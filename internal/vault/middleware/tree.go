// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"context"
	"fmt"
	"strings"

	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/1Password/connect-sdk-go/connect"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

var ErrorNoVaultProvided = fmt.Errorf("no vault has been specified, provide one reading from MOUNT/tree/VAULT/ITEM, or configure a default writing to MOUNT/1password")

func vaultName(data *framework.FieldData, storage logical.Storage) (vault string, err error) {
	if vaultI, ok := data.GetOk("vault"); ok {
		vault := strings.TrimSuffix(vaultI.(string), "/")
		if vault != "" {
			return vault, nil
		}
	}

	config, err := ConfigFromStorage(context.Background(), storage)
	if err != nil {
		return "", fmt.Errorf("could not get config from storage: %w", err)
	}

	if config != nil && config.Vault != "" {
		return config.Vault, nil
	}

	return "", ErrorNoVaultProvided
}

func ReadTree(client connect.Client, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vault, err := vaultName(data, req.Storage)
	if err != nil {
		return nil, err
	}

	item, err := client.GetItem(data.Get("id").(string), vault)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve item: %w", err)
	}

	tree := config.NewEntry("root", yaml.MappingNode)

	if err := tree.FromOP(item.Fields); err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: tree.AsMap().(map[string]any),
	}, nil
}

func ListTrees(client connect.Client, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	vault, err := vaultName(data, req.Storage)
	if err != nil {
		return nil, err
	}

	items, err := client.GetItems(vault)
	if err != nil {
		return nil, fmt.Errorf("could not list items: %w", err)
	}

	retMap := map[string]any{}
	retList := []string{}
	for _, item := range items {
		key := fmt.Sprintf("%s %s", item.Title, item.ID)
		retMap[key] = item.ID
		retList = append(retList, key)
	}

	return logical.ListResponseWithInfo(retList, retMap), nil
}
