// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
		if strings.Contains(err.Error(), fmt.Sprintf("\"%s\" isn't an item in the ", name)) {
			return client.Create(item)
		}

		return fmt.Errorf("could not fetch remote 1password item to compare against: %w", err)
	}

	if remote.GetValue("password") == item.GetValue("password") {
		logrus.Warn("item is already up to date")
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
