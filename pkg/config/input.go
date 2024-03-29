// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	opClient "git.rob.mx/nidito/joao/pkg/op-client"
	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const warnChecksumMismatch = `1Password item %s changed and checksum was not updated.
Expected: %s
found   : %s`

func Load(ref string, preferRemote bool) (*Config, error) {
	if preferRemote {
		name := ref
		vault := ""

		if argIsYAMLFile(ref) {
			path, err := filepath.Abs(ref)
			if err != nil {
				return nil, fmt.Errorf("could not find asbolute path to file %s: %w", ref, err)
			}
			name, vault, err = VaultAndNameFrom(path, nil)
			if err != nil {
				return nil, err
			}
		} else {
			parts := strings.SplitN(ref, "/", 2)
			if len(parts) > 1 {
				vault = parts[0]
				name = parts[1]
			}
		}

		item, err := opClient.Get(vault, name)
		if err != nil {
			return nil, err
		}

		return FromOP(item)
	}

	if !argIsYAMLFile(ref) {
		return nil, fmt.Errorf("could not load %s from local as it's not a path", ref)
	}

	path, err := filepath.Abs(ref)
	if err != nil {
		return nil, fmt.Errorf("could not find asbolute path to file %s: %w", ref, err)
	}
	cfg, err := FromFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not load file %s: %w", path, err)
	}
	return cfg, nil
}

// FromFile reads a path and returns a config.
func FromFile(path string) (*Config, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s", path)
	}

	if len(buf) == 0 {
		buf = []byte("{}")
	}

	name, vault, err := VaultAndNameFrom(path, buf)
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
		return nil, fmt.Errorf("could not parse %w", err)
	}
	cfg.Tree.SetPath([]string{}, ".")

	return cfg, nil
}

// FromOP reads a config from an op item and returns a config.
func FromOP(item *op.Item) (*Config, error) {
	cfg := &Config{
		Vault: item.Vault.ID,
		Name:  item.Title,
		Tree:  NewEntry("root", yaml.MappingNode),
	}

	if cs := opClient.Checksum(item.Fields); cs != item.GetValue("password") {
		logrus.Warnf(warnChecksumMismatch, fmt.Sprintf("%s/%s", item.Vault.ID, item.Title), cs, item.GetValue("password"))
	}
	err := cfg.Tree.FromOP(item.Fields)
	return cfg, err
}
