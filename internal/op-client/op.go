// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package opclient

import (
	"fmt"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
)

var client opClient

type opClient interface {
	Get(vault, name string) (*op.Item, error)
	Update(item *op.Item, remote *op.Item) error
	Create(item *op.Item) error
	List(vault, prefix string) ([]string, error)
}

func init() {
	client = &CLI{}
}

func Use(newClient opClient) {
	client = newClient
}

func Get(vault, name string) (*op.Item, error) {
	return client.Get(vault, name)
}

func Update(vault, name string, item *op.Item) error {
	remote, err := client.Get(vault, name)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("\"%s\" isn't an item in ", name)) {
			return client.Create(item)
		}

		return fmt.Errorf("could not fetch remote 1password item to compare against: %w", err)
	}

	remoteCS := Checksum(remote.Fields)
	// we're checking the checksum we just calculated matches the stored on remote
	// and that remoteCS matching the current item's stored password
	// nolint:gocritic
	if remoteCS == item.GetValue("password") && remoteCS == remote.GetValue("password") {
		logrus.Debugf("remote %s\nlocal %s", remoteCS, item.GetValue("password"))
		logrus.Warnf("item %s/%s is already up to date", item.Vault.ID, item.Title)
		return nil
	}

	logrus.Infof("Item %s/%s already exists, updating", item.Vault.ID, item.Title)
	return client.Update(item, remote)
}

func List(vault, prefix string) ([]string, error) {
	return client.List(vault, prefix)
}

func keyForField(field *op.ItemField) string {
	name := strings.ReplaceAll(field.Label, ".", "\\.")
	if field.Section != nil {
		name = field.Section.ID + "." + name
	}
	return name
}
