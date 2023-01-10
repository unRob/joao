// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package mock

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
)

const Host = "http://localhost:8080"
const Token = "test_token"

var randPool = []rune("0123456789abcdefghijklmnopqrstuvwxyz")
var items = map[string]*onepassword.Item{}

func Add(item *onepassword.Item) *onepassword.Item {
	item.ID = itemID()
	items[item.ID] = item
	return item
}

func Update(item *onepassword.Item) *onepassword.Item {
	items[item.ID] = item
	return item
}

func Clear() {
	items = map[string]*onepassword.Item{}
}

func Delete(key string) {
	delete(items, key)
}

var (
	Vaults = []onepassword.Vault{
		{
			ID:   "aabbccddeeffgghhiijjkkllmm",
			Name: "Zeroth Vault",
		},
		{
			ID:   "00011122233344455566677788",
			Name: "First Vault",
		},
	}
)

type Client struct{}

func (m *Client) GetVaults() ([]onepassword.Vault, error) {
	return Vaults, nil
}

func (m *Client) GetVault(uuid string) (*onepassword.Vault, error) {
	for _, v := range Vaults {
		if v.Name == uuid || v.ID == uuid {
			return &v, nil
		}
	}

	return nil, nil
}

func (m *Client) GetVaultByUUID(uuid string) (*onepassword.Vault, error) {
	return m.GetVault(uuid)
}

func (m *Client) GetVaultByTitle(title string) (*onepassword.Vault, error) {
	return m.GetVault(title)
}

func (m *Client) GetVaultsByTitle(uuid string) ([]onepassword.Vault, error) {
	res := []onepassword.Vault{}
	for _, v := range Vaults {
		if v.Name == uuid || v.ID == uuid {
			res = append(res, v)
		}
	}

	return res, nil
}

func (m *Client) GetItems(vaultQuery string) ([]onepassword.Item, error) {
	res := []onepassword.Item{}
	for _, item := range items {
		if item.Vault.ID == vaultQuery {
			res = append(res, *item)
		}
	}
	return res, nil
}

func (m *Client) GetItem(itemQuery, vaultQuery string) (*onepassword.Item, error) {
	return get(itemQuery, vaultQuery)
}

func (m *Client) GetItemByUUID(uuid string, vaultQuery string) (*onepassword.Item, error) {
	return get(uuid, vaultQuery)
}

func (m *Client) GetItemByTitle(title string, vaultQuery string) (*onepassword.Item, error) {
	return get(title, vaultQuery)
}

func (m *Client) GetItemsByTitle(title string, vaultQuery string) ([]onepassword.Item, error) {
	res := []onepassword.Item{}
	for _, v := range items {
		if v.Title == title {
			res = append(res, *v)
		}
	}

	return res, nil
}

func (m *Client) CreateItem(item *onepassword.Item, vaultQuery string) (*onepassword.Item, error) {
	item.CreatedAt = time.Now()
	item.Vault.ID = vaultQuery
	return Add(item), nil
}

func (m *Client) UpdateItem(item *onepassword.Item, vaultQuery string) (*onepassword.Item, error) {
	return Update(item), nil
}

func (m *Client) DeleteItem(item *onepassword.Item, vaultQuery string) error {
	return deleteItem(item, vaultQuery)
}

func (m *Client) DeleteItemByID(itemUUID string, vaultQuery string) error {
	item, err := get(itemUUID, vaultQuery)
	if err != nil {
		return err
	}
	return deleteItem(item, vaultQuery)
}

func (m *Client) DeleteItemByTitle(title string, vaultQuery string) error {
	item, err := get(title, vaultQuery)
	if err != nil {
		return err
	}
	return deleteItem(item, vaultQuery)
}

func (m *Client) GetFiles(itemQuery string, vaultQuery string) ([]onepassword.File, error) {
	return nil, nil
}

func (m *Client) GetFile(uuid string, itemQuery string, vaultQuery string) (*onepassword.File, error) {
	return nil, nil
}

func (m *Client) GetFileContent(file *onepassword.File) ([]byte, error) {
	return nil, nil
}

func (m *Client) DownloadFile(file *onepassword.File, targetDirectory string, overwrite bool) (string, error) {
	return "", nil
}

func (m *Client) LoadStructFromItemByUUID(config any, itemUUID string, vaultQuery string) error {
	return nil
}

func (m *Client) LoadStructFromItemByTitle(config any, itemTitle string, vaultQuery string) error {
	return nil
}

func (m *Client) LoadStructFromItem(config any, itemQuery string, vaultQuery string) error {
	return nil
}

func (m *Client) LoadStruct(config any) error {
	return nil
}

func itemID() string {
	b := make([]rune, 26)
	for i := range b {
		b[i] = randPool[rand.Intn(len(randPool))] // nolint: gosec
	}
	return string(b)
}

func get(itemUUID, vaultUUID string) (*onepassword.Item, error) {
	for _, item := range items {
		if (item.ID == itemUUID || item.Title == itemUUID) && item.Vault.ID == vaultUUID {
			return item, nil
		}
	}

	return nil, fmt.Errorf("could not retrieve item with id %s in vault %s", itemUUID, vaultUUID)
}

func deleteItem(item *onepassword.Item, vaultUUID string) error {
	if item.Vault.ID != vaultUUID {
		return fmt.Errorf("could not delete item: %s: not found in vault %s", item.Title, vaultUUID)
	}
	Delete(item.ID)
	return nil
}

var _ connect.Client = &Client{}
