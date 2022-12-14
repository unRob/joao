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
	op "github.com/1Password/connect-sdk-go/onepassword"
)

var client opClient

type opClient interface {
	Get(vault, name string) (*op.Item, error)
	Update(vault, name string, item *op.Item) error
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
	return client.Update(vault, name, item)
}

func List(vault, prefix string) ([]string, error) {
	return client.List(vault, prefix)
}
