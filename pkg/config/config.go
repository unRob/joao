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
	"encoding/json"
	"fmt"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Path  string
	Vault string
	Name  string
	Tree  *Entry
}

var redactOutput = false

func (cfg *Config) ToMap() map[string]interface{} {
	ret := map[string]interface{}{}
	for _, child := range cfg.Tree.Children {
		ret[child.Key] = child.AsMap()
	}
	return ret
}

func (cfg *Config) ToOP() *op.Item {
	annotationsSection := &op.ItemSection{
		ID:    "~annotations",
		Label: "~annotations",
	}
	sections := []*op.ItemSection{annotationsSection}
	fields := []*op.ItemField{
		{
			ID:      "password",
			Type:    "CONCEALED",
			Purpose: "PASSWORD",
			Label:   "password",
			Value:   "hash",
		}, {
			ID:      "notesPlain",
			Type:    "STRING",
			Purpose: "NOTES",
			Label:   "notesPlain",
			Value:   "flushed by joao",
		},
	}

	for key, leaf := range cfg.Tree.Children {
		if len(leaf.Children) == 0 {
			fields = append(fields, leaf.ToOP(annotationsSection)...)
			continue
		}

		if !leaf.isSequence {
			sections = append(sections, &op.ItemSection{
				ID:    key,
				Label: key,
			})
		} else {
			fmt.Printf("Found sequence for %s", leaf.Key)
		}

		for _, child := range leaf.Children {
			fields = append(fields, child.ToOP(annotationsSection)...)
		}
	}

	return &op.Item{
		Title:    cfg.Name,
		Sections: sections,
		Vault:    op.ItemVault{ID: "nidito-admin"},
		Category: op.Password,
		Fields:   fields,
	}
}

func ConfigFromYAML(data []byte) (*Config, error) {
	cfg := &Config{
		Vault: "vault",
		Name:  "title",
		Tree:  NewEntry("root"),
	}

	yaml.Unmarshal(data, cfg.Tree.Children)

	for k, leaf := range cfg.Tree.Children {
		leaf.SetKey(k, []string{})
	}

	return cfg, nil
}

func ConfigFromOP(item *op.Item) (*Config, error) {
	cfg := &Config{
		Vault: item.Vault.ID,
		Name:  item.Title,
		Tree:  NewEntry("root"),
	}

	err := cfg.Tree.FromOP(item.Fields)
	return cfg, err
}

func (cfg *Config) MarshalYAML() (interface{}, error) {
	return cfg.Tree.MarshalYAML()
}

func (cfg *Config) AsYAML(redacted bool) ([]byte, error) {
	redactOutput = redacted
	bytes, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("could not serialize config as yaml: %w", err)
	}
	return bytes, nil
}

func (cfg *Config) AsJSON(redacted bool, item bool) ([]byte, error) {
	var repr interface{}
	if item {
		repr = cfg.ToOP()
	} else {
		redactOutput = redacted
		repr = cfg.ToMap()
	}

	bytes, err := json.Marshal(repr)
	if err != nil {
		return nil, fmt.Errorf("could not serialize config as json: %w", err)
	}
	return bytes, nil
}
