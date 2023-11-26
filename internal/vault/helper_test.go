// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package vault_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"git.rob.mx/nidito/joao/internal/testdata/opconnect"
	"git.rob.mx/nidito/joao/internal/vault"
	"git.rob.mx/nidito/joao/internal/vault/middleware"
	"github.com/1Password/connect-sdk-go/connect"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/logical"
)

func mapsEqual(t *testing.T, actual, expected map[string]any) {
	for key, want := range expected {
		if have, ok := actual[key]; !ok || want != have {
			t.Fail()
			t.Errorf(`field mismatch for "%v". \nwanted: %v\ngot: %v"`, key, want, have)
		}
	}
	if t.Failed() {
		t.FailNow()
	}
}

func testConfig() *logical.BackendConfig {
	config := logical.TestBackendConfig()
	config.StorageView = new(logical.InmemStorage)
	config.Logger = hclog.NewNullLogger()
	config.System = &logical.StaticSystemView{
		DefaultLeaseTTLVal: 1 * time.Hour,
		MaxLeaseTTLVal:     2 * time.Hour,
	}
	return config
}

func getBackend(tb testing.TB) (logical.Backend, logical.Storage) {
	tb.Helper()
	cfg := testConfig()
	ctx := context.Background()

	data, err := json.Marshal(map[string]string{"host": opconnect.Host, "token": opconnect.Token, "vault": opconnect.Vaults[0].ID})
	if err != nil {
		tb.Fatalf("Could not serialize config for client: %s", err)
	}

	_ = cfg.StorageView.Put(ctx, &logical.StorageEntry{
		Key:   middleware.ConfigPath,
		Value: data,
	})

	setOnePassswordConnectMocks()
	b, err := vault.Factory(context.Background(), cfg)

	if err != nil {
		tb.Fatal(err)
	}
	return b, cfg.StorageView
}

func getUnconfiguredBackend(tb testing.TB) (logical.Backend, logical.Storage) {
	tb.Helper()
	cfg := testConfig()
	setOnePassswordConnectMocks()
	b, err := vault.Factory(context.Background(), cfg)

	if err != nil {
		tb.Fatal(err)
	}
	return b, cfg.StorageView
}

func setOnePassswordConnectMocks() {
	vault.ConnectClientFactory = func(s logical.Storage) (connect.Client, error) {
		return &opconnect.Client{}, nil
	}
}
