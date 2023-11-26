// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package vault_test

import (
	"context"
	"testing"

	"git.rob.mx/nidito/joao/internal/testdata/opconnect"
	"git.rob.mx/nidito/joao/internal/vault"
	"git.rob.mx/nidito/joao/internal/vault/middleware"
	"github.com/1Password/connect-sdk-go/connect"
	"github.com/hashicorp/vault/sdk/logical"
)

func init() {
	vault.ConnectClientFactory = func(s logical.Storage) (connect.Client, error) {
		return &opconnect.Client{}, nil
	}
}

func TestConfiguredBackend(t *testing.T) {
	b, reqStorage := getBackend(t)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      middleware.ConfigPath,
		Storage:   reqStorage,
	})

	if err != nil && resp.IsError() {
		t.Fatalf("Unexpected error with config set: %s => %v", err, resp)
	}

	if resp.Data["token"] != opconnect.Token {
		t.Errorf("Found unknown token: %s", resp.Data["token"])
	}

	if resp.Data["host"] != opconnect.Host {
		t.Errorf("Found unknown host: %s", resp.Data["host"])
	}

	if resp.Data["vault"] != opconnect.Vaults[0].ID {
		t.Errorf("Found unknown vault: %s", resp.Data["vault"])
	}
}

func TestUnconfiguredBackend(t *testing.T) {
	b, reqStorage := getUnconfiguredBackend(t)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      middleware.ConfigPath,
		Storage:   reqStorage,
	})

	if err != nil && resp.IsError() {
		t.Fatalf("Unexpected error with unconfigured: %s => %v", err, resp)
	}

	if resp != nil {
		t.Fatalf("Found a response where none was expected: %v", resp)
	}

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "tree/someItem",
		Storage:   reqStorage,
	})

	if err == nil && !resp.IsError() {
		t.Fatalf("Expected error with no config set: %v", resp)
	}

	expected := middleware.ErrorNoVaultProvided.Error()
	if actual := err.Error(); actual != expected {
		t.Fatalf("unconfigured client threw wrong error: \nwanted: %s\ngot: %s", expected, actual)
	}
}
