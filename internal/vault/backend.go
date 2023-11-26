// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package vault

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"git.rob.mx/nidito/joao/internal/vault/middleware"
	"git.rob.mx/nidito/joao/pkg/version"
	"github.com/1Password/connect-sdk-go/connect"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	ttlcache "github.com/jellydator/ttlcache/v3"
)

const (
	userAgent = "joao/%s"
)

type backend struct {
	*framework.Backend
	configCache *ttlcache.Cache[string, string]
	client      *connect.Client
}

var ConnectClientFactory func(s logical.Storage) (connect.Client, error) = onePasswordConnectClient

// Factory returns a new backend as logical.Backend.
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := newBackend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

type clientCallback func(client connect.Client, r *logical.Request, fd *framework.FieldData) (*logical.Response, error)

func withClient(b *backend, callback clientCallback) framework.OperationFunc {
	return func(ctx context.Context, r *logical.Request, fd *framework.FieldData) (*logical.Response, error) {
		client, err := b.Client(r.Storage)
		if err != nil {
			return nil, fmt.Errorf("plugin is not configured: %s", err)
		}

		return callback(client, r, fd)
	}
}

func itemPattern(name string) string {
	return fmt.Sprintf("(?P<%s>\\w(([\\w-.:]+)?\\w)?)", name)
}

func optionalVaultPattern(suffix string) string {
	return fmt.Sprintf("(?P<vault>([\\w:]+)%s)?", suffix)
}

func newBackend() *backend {
	var b = &backend{
		configCache: ttlcache.New(
			ttlcache.WithTTL[string, string](5 * time.Minute),
		),
	}

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Help:        "joao reads configuration entries from 1Password Connect",
		PathsSpecial: &logical.Paths{
			SealWrapStorage: []string{
				middleware.ConfigPath,
			},
		},
		Paths: framework.PathAppend(
			[]*framework.Path{
				{
					Pattern:         middleware.ConfigPath,
					HelpSynopsis:    "Configures the connection to a 1Password Connect Server",
					HelpDescription: "Provide a `host` and `token`, with an optional default `vault` to query 1Password Connect at",
					Fields: map[string]*framework.FieldSchema{
						"host": {
							Type:        framework.TypeString,
							Description: "The address for the 1Password Connect server",
						},
						"token": {
							Type:        framework.TypeString,
							Description: "A 1Password Connect token",
						},
						"vault": {
							Type:        framework.TypeString,
							Description: "An optional vault id or name to use for queries",
						},
					},
					Operations: map[logical.Operation]framework.OperationHandler{
						logical.ReadOperation: &framework.PathOperation{
							Callback: middleware.ReadConfig,
						},
						logical.UpdateOperation: &framework.PathOperation{
							Callback: func(ctx context.Context, r *logical.Request, fd *framework.FieldData) (*logical.Response, error) {
								res, err := middleware.WriteConfig(ctx, r, fd)
								if err != nil {
									return nil, err
								}

								b.client = nil
								if _, err := b.Client(r.Storage); err != nil {
									return nil, err
								}
								return res, nil
							},
						},
					},
				},
				{
					Pattern:      "trees/" + optionalVaultPattern(""),
					HelpSynopsis: `List configuration trees`,
					Operations: map[logical.Operation]framework.OperationHandler{
						logical.ListOperation: &framework.PathOperation{
							Callback: withClient(b, middleware.ListTrees),
							Summary:  "List available entries",
						},
					},
					Fields: map[string]*framework.FieldSchema{
						"vault": {
							Type:        framework.TypeString,
							Description: "Specifies the id of the vault to list from.",
							Required:    true,
						},
					},
				},
				{
					Pattern:      "tree/" + optionalVaultPattern("/") + itemPattern("id"),
					HelpSynopsis: `Returns a configuration tree`,
					Operations: map[logical.Operation]framework.OperationHandler{
						logical.ReadOperation: &framework.PathOperation{
							Callback: withClient(b, middleware.ReadTree),
							Summary:  "Retrieve nested key values from specified item",
						},
					},
					Fields: map[string]*framework.FieldSchema{
						"id": {
							Type:        framework.TypeString,
							Description: "The item name or id to read",
							Required:    true,
						},
						"vault": {
							Type:        framework.TypeString,
							Description: "The vault name or id to read from",
							Required:    true,
						},
					},
				},
			},
		),
		Secrets: []*framework.Secret{},
	}

	return b
}

func (b *backend) Client(s logical.Storage) (connect.Client, error) {
	if b.client != nil {
		return *b.client, nil
	}

	client, err := ConnectClientFactory(s)
	if err != nil {
		return nil, err
	}
	b.client = &client
	return client, nil
}

func onePasswordConnectClient(s logical.Storage) (connect.Client, error) {
	config, err := middleware.ConfigFromStorage(context.Background(), s)
	if err != nil {
		return nil, fmt.Errorf("error retrieving config for client: %w", err)
	}

	if config == nil {
		return nil, fmt.Errorf("no config set for backend, write host, token and vault to [mount]/1password")
	}

	http.DefaultClient.Timeout = 15 * time.Second
	client := connect.NewClientWithUserAgent(config.Host, config.Token, fmt.Sprintf(userAgent, version.Version))

	return client, nil
}
