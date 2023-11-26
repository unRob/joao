// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package vault_test

import (
	"context"
	"testing"

	"git.rob.mx/nidito/joao/internal/testdata/opconnect"
	"github.com/hashicorp/vault/sdk/logical"
)

func TestConfigEmpty(t *testing.T) {
	b, reqStorage := getUnconfiguredBackend(t)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "1password",
		Storage:   reqStorage,
	})

	if err != nil {
		t.Fatalf("Could not issue request: %s", err)
	}

	if resp != nil {
		t.Fatalf("got response, expected none %v", resp)
	}
}

func TestConfigDefault(t *testing.T) {
	b, reqStorage := getBackend(t)

	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "1password",
		Storage:   reqStorage,
	})

	if err != nil {
		t.Fatalf("Could not issue request: %s", err)
	}

	if resp.IsError() {
		t.Fatalf("get request threw error: %s", resp.Error())
	}

	if len(resp.Data) == 0 {
		t.Fatal("got no response, expected something!")
	}

	mapsEqual(t, resp.Data, map[string]any{"host": opconnect.Host, "token": opconnect.Token, "vault": opconnect.Vaults[0].ID})
}

func TestConfigUpdate(t *testing.T) {
	b, reqStorage := getBackend(t)
	expected := map[string]any{
		"host":  "mira",
		"token": "un",
		"vault": "salmón",
	}
	resp, err := b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.UpdateOperation,
		Path:      "1password",
		Data:      expected,
		Storage:   reqStorage,
	})

	if err != nil {
		t.Fatalf("Could not issue update request: %s", err)
	}

	if resp != nil && resp.IsError() {
		t.Fatal(resp.Error())
	}

	resp, err = b.HandleRequest(context.Background(), &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "1password",
		Storage:   reqStorage,
	})
	if err != nil {
		t.Fatalf("Could not issue read after update request: %s", err)
	}

	if resp.IsError() {
		t.Fatalf("get after update request threw error: %s", resp.Error())
	}

	if len(resp.Data) == 0 {
		t.Fatal("got no response on get after update, expected something!")
	}

	mapsEqual(t, resp.Data, expected)
}
