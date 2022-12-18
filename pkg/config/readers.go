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
package config

import (
	"fmt"
	"io/ioutil"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// FromFile reads a path and returns a config.
func FromFile(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s", path)
	}

	if len(buf) == 0 {
		buf = []byte("{}")
	}

	name, vault, err := vaultAndNameFrom(path, buf)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Found name: %s and vault: %s", name, vault)

	cfg, err := FromYAML(buf)
	if err != nil {
		return nil, err
	}
	cfg.Name = name
	cfg.Vault = vault

	return cfg, nil
}

// FromYAML reads yaml bytes and returns a config.
func FromYAML(data []byte) (*Config, error) {
	cfg := &Config{
		Tree: NewEntry("root", yaml.MappingNode),
	}

	err := yaml.Unmarshal(data, &cfg.Tree)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// FromOP reads a config from an op item and returns a config.
func FromOP(item *op.Item) (*Config, error) {
	cfg := &Config{
		Vault: item.Vault.ID,
		Name:  item.Title,
		Tree:  NewEntry("root", yaml.MappingNode),
	}

	if cs := checksum(item.Fields); cs != item.GetValue("password") {
		logrus.Warnf("1Password item changed and checksum was not updated. Expected %s, found %s", cs, item.GetValue("password"))
	}
	err := cfg.Tree.FromOP(item.Fields)
	return cfg, err
}
