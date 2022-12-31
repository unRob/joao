// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package opclient

import (
	"github.com/1Password/connect-sdk-go/connect"
	op "github.com/1Password/connect-sdk-go/onepassword"
)

// UUIDLength defines the required length of UUIDs.
const UUIDLength = 26

// IsValidClientUUID returns true if the given client uuid is valid.
func IsValidClientUUID(uuid string) bool {
	if len(uuid) != UUIDLength {
		return false
	}

	for _, c := range uuid {
		valid := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
		if !valid {
			return false
		}
	}

	return true
}

type Connect struct {
	client connect.Client
}

const userAgent = "nidito-joao"

func NewConnect(host, token string) *Connect {
	client := connect.NewClientWithUserAgent(host, token, userAgent)
	return &Connect{client: client}
}

// func (b *Connect) getVaultId(vaultIdentifier string) (string, error) {
// 	if !IsValidClientUUID(vaultIdentifier) {
// 		vaults, err := b.client.GetVaultsByTitle(vaultIdentifier)
// 		if err != nil {
// 			return "", err
// 		}

// 		if len(vaults) == 0 {
// 			return "", fmt.Errorf("no vaults found with identifier %q", vaultIdentifier)
// 		}

// 		oldestVault := vaults[0]
// 		if len(vaults) > 1 {
// 			for _, returnedVault := range vaults {
// 				if returnedVault.CreatedAt.Before(oldestVault.CreatedAt) {
// 					oldestVault = returnedVault
// 				}
// 			}
// 			logrus.Infof("%v 1Password vaults found with the title %q. Will use vault %q as it is the oldest.", len(vaults), vaultIdentifier, oldestVault.ID)
// 		}
// 		vaultIdentifier = oldestVault.ID
// 	}
// 	return vaultIdentifier, nil
// }

func (b *Connect) Get(vault, name string) (*op.Item, error) {
	return b.client.GetItem(name, vault)
}

func (b *Connect) Update(item *op.Item) error {
	_, err := b.client.UpdateItem(item, item.Vault.ID)
	return err
}

func (b *Connect) List(vault, prefix string) ([]string, error) {
	// TODO: get this done
	return nil, nil
}
