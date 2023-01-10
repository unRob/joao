// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package vault_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"git.rob.mx/nidito/joao/internal/op-client/mock"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/hashicorp/vault/sdk/logical"
)

func getTestBackendWithConfig(t *testing.T) (logical.Backend, logical.Storage) {
	t.Helper()
	return getBackend(t)
}

func TestReadEntry(t *testing.T) {
	b, reqStorage := getTestBackendWithConfig(t)
	mock.Clear()
	item := mock.Add(generateConfigItem("service:test"))
	expected := map[string]any{
		"boolean": false,
		"integer": 42,
		"list":    []string{"first item", "second item"},
		"nested": map[string]any{
			"boolean": true,
			"integer": 42,
			"string":  "this is a string",
		},
	}
	expectedJSON, _ := json.Marshal(expected)

	t.Run("with default vault", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("tree/%v", item.Title),
			Storage:   reqStorage,
		})

		if err != nil {
			t.Fatal("read request failed:", err)
		}

		if resp == nil {
			t.Fatal("Item missing")
		}

		if resp.IsError() {
			t.Fatal(resp.Error())
		}

		gotJSON, _ := json.Marshal(resp.Data)

		if string(expectedJSON) != string(gotJSON) {
			t.Fatalf("unexpectedJSON response.\nwanted: %s\ngot: %s", string(expectedJSON), string(gotJSON))
		}
	})

	t.Run("with explicit vault", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ReadOperation,
			Path:      fmt.Sprintf("tree/%s/%s", item.Vault.ID, item.Title),
			Storage:   reqStorage,
		})

		if err != nil {
			t.Fatal("read request failed:", err)
		}

		if resp == nil {
			t.Fatal("Item missing")
		}

		if resp.IsError() {
			t.Fatal(resp.Error())
		}

		gotJSON, _ := json.Marshal(resp.Data)

		if string(expectedJSON) != string(gotJSON) {
			t.Fatalf("unexpectedJSON response.\nwanted: %s\ngot: %s", string(expectedJSON), string(gotJSON))
		}
	})
}

func TestListEntries(t *testing.T) {
	b, reqStorage := getTestBackendWithConfig(t)
	mock.Clear()
	item := mock.Add(generateConfigItem("service:test"))

	expected := map[string]any{
		"keys": []string{
			"service:test " + item.ID,
		},
		"key_info": map[string]string{
			"service:test " + item.ID: item.ID,
		},
	}

	expectedJSON, _ := json.Marshal(expected)
	t.Run("with default vault", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "trees/",
			Storage:   reqStorage,
		})

		if err != nil {
			t.Fatal(err)
		}

		if resp.IsError() {
			t.Fatal(resp.Error())
		}

		gotJSON, _ := json.Marshal(resp.Data)

		if string(expectedJSON) != string(gotJSON) {
			t.Fatalf("unexpectedJSON response.\nwanted: %s\ngot: %s", string(expectedJSON), string(gotJSON))
		}
	})

	t.Run("with explicit vault", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "trees/" + item.Vault.ID,
			Storage:   reqStorage,
		})

		if err != nil {
			t.Fatal(err)
		}

		if resp.IsError() {
			t.Fatal(resp.Error())
		}

		gotJSON, _ := json.Marshal(resp.Data)

		if string(expectedJSON) != string(gotJSON) {
			t.Fatalf("unexpectedJSON response.\nwanted: %s\ngot: %s", string(expectedJSON), string(gotJSON))
		}
	})

	t.Run("with explicit unknown vault", func(t *testing.T) {
		resp, err := b.HandleRequest(context.Background(), &logical.Request{
			Operation: logical.ListOperation,
			Path:      "trees/asdf",
			Storage:   reqStorage,
		})

		if err != nil {
			t.Fatal(err)
		}

		if resp.IsError() {
			t.Fatal(resp.Error())
		}

		gotJSON, _ := json.Marshal(resp.Data)

		if string(gotJSON) != "{}" {
			t.Fatalf("unexpectedJSON response, wanted %s, got %s", "{}", string(gotJSON))
		}
	})
}

func generateConfigItem(title string) *onepassword.Item {
	return &onepassword.Item{
		Category: "password",
		Title:    title,
		Vault: onepassword.ItemVault{
			ID: mock.Vaults[0].ID,
		},
		Fields: []*onepassword.ItemField{
			{
				ID:      "nested.string",
				Type:    "STRING",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Label:   "string",
				Value:   "this is a string",
			},
			{
				ID:      "nested.boolean",
				Type:    "STRING",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Label:   "boolean",
				Value:   "true",
			},
			{
				ID:      "nested.integer",
				Type:    "STRING",
				Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
				Label:   "integer",
				Value:   "42",
			},
			{
				ID:      "list.0",
				Type:    "STRING",
				Section: &onepassword.ItemSection{ID: "list", Label: "list"},
				Label:   "0",
				Value:   "first item",
			},
			{
				ID:      "list.1",
				Type:    "STRING",
				Section: &onepassword.ItemSection{ID: "list", Label: "list"},
				Label:   "1",
				Value:   "second item",
			},
			{
				ID:    "boolean",
				Type:  "STRING",
				Label: "boolean",
				Value: "false",
			},
			{
				ID:    "integer",
				Type:  "STRING",
				Label: "integer",
				Value: "42",
			},
			{
				ID:      "~annotations.integer",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "integer",
				Value:   "int",
			},
			{
				ID:      "~annotations.boolean",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "boolean",
				Value:   "bool",
			},
			{
				ID:      "~annotations.nested.integer",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.integer",
				Value:   "int",
			},
			{
				ID:      "~annotations.nested.boolean",
				Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
				Type:    "STRING",
				Label:   "nested.boolean",
				Value:   "bool",
			},
		},
		Sections: []*onepassword.ItemSection{
			{
				ID:    "~annotations",
				Label: "~annotations",
			},
			{
				ID:    "nested",
				Label: "nested",
			},
			{
				ID:    "list",
				Label: "list",
			},
		},
	}
}
