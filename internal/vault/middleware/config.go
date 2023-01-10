// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

const (
	ConfigPath = "1password"
)

type Config struct {
	Host  string `json:"host"`
	Token string `json:"token"`
	Vault string `json:"vault"`
}

func ConfigFromStorage(ctx context.Context, s logical.Storage) (*Config, error) {
	entry, err := s.Get(ctx, ConfigPath)
	if err != nil || entry == nil {
		return nil, err
	}

	var config Config
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func ReadConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	cfg, err := ConfigFromStorage(ctx, req.Storage)
	if err != nil || cfg == nil {
		return nil, err
	}

	return &logical.Response{
		Data: map[string]any{
			"host":  cfg.Host,
			"token": cfg.Token,
			"vault": cfg.Vault,
		},
	}, nil
}

func WriteConfig(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	existing, err := ConfigFromStorage(ctx, req.Storage)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		existing = &Config{}
	}

	if host, ok := data.GetOk("host"); ok {
		existing.Host = host.(string)
	}

	if token, ok := data.GetOk("token"); ok {
		existing.Token = token.(string)
	}

	if opVault, ok := data.GetOk("vault"); ok {
		existing.Vault = opVault.(string)
	}

	entry, err := logical.StorageEntryJSON(ConfigPath, existing)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	return nil, nil
}
