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
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const YAMLTypeSecret string = "!!secret"
const YAMLTypeMetaConfig string = "!!joao"

type Config struct {
	Vault string
	Name  string
	Tree  *Entry
}

var redactOutput = false
var annotationsSection = &op.ItemSection{
	ID:    "~annotations",
	Label: "~annotations",
}
var defaultItemFields = []*op.ItemField{
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

func (cfg *Config) ToMap() map[string]any {
	ret := map[string]any{}
	for _, child := range cfg.Tree.Content {
		if child.Name() == "" {
			continue
		}
		ret[child.Name()] = child.AsMap()
	}
	return ret
}

func (cfg *Config) ToOP() *op.Item {
	sections := []*op.ItemSection{annotationsSection}
	fields := append([]*op.ItemField{}, defaultItemFields...)

	newHash := md5.New()
	datafields := cfg.Tree.ToOP()
	for _, field := range datafields {
		newHash.Write([]byte(field.ID + field.Value))
	}
	fields[0].Value = fmt.Sprintf("%x", newHash.Sum(nil))
	fields = append(fields, datafields...)

	for i := 0; i < len(cfg.Tree.Content); i += 2 {
		value := cfg.Tree.Content[i+1]
		if value.Type == YAMLTypeMetaConfig {
			continue
		}

		if value.Kind == yaml.MappingNode {
			sections = append(sections, &op.ItemSection{
				ID:    value.Name(),
				Label: value.Name(),
			})
		}
	}

	return &op.Item{
		Title:    cfg.Name,
		Sections: sections,
		Vault:    op.ItemVault{ID: cfg.Vault},
		Category: op.Password,
		Fields:   fields,
	}
}

type opDetails struct {
	Vault        string `yaml:"vault"`
	Name         string `yaml:"name"`
	NameTemplate string `yaml:"nameTemplate"`
	Repo         string
}

// type opConfig interface {
// 	Name() string
// 	Vault() string
// }

// type inFileConfig struct {
// 	*opDetails
// 	*yaml.Node
// }

// type virtualConfig struct {
// 	*opDetails
// }

// func (ifc *inFileConfig) MarshalYAML() (any, error) {
// 	return ifc.Node, nil
// }

// func (vc *virtualConfig) MarshalYAML() (any, error) {
// 	return nil, nil
// }

// func (ifc *inFileConfig) UnmarshalYAML(node *yaml.Node) error {
// 	ifc.Node = node
// 	d := &opDetails{}

// 	if err := node.Decode(&d); err != nil {
// 		return err
// 	}
// 	ifc.opDetails = d

// 	return nil
// }

// func (ifc *inFileConfig) Name() string {
// 	return ifc.opDetails.Name
// }

// func (ifc *inFileConfig) Vault() string {
// 	return ifc.opDetails.Name
// }

type singleModeConfig struct {
	Config *opDetails `yaml:"_config,omitempty"`
}

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

func scalarsIn(data map[string]yaml.Node, parents []string) ([]string, error) {
	keys := []string{}
	for key, leaf := range data {
		if key == "_config" && len(parents) == 0 {
			continue
		}
		switch leaf.Kind {
		case yaml.ScalarNode:
			newKey := strings.Join(append(parents, key), ".")
			keys = append(keys, newKey)
		case yaml.MappingNode, yaml.DocumentNode, yaml.SequenceNode:
			sub := map[string]yaml.Node{}
			if leaf.Kind == yaml.SequenceNode {
				list := []yaml.Node{}
				if err := leaf.Decode(&list); err != nil {
					return keys, err
				}

				for idx, child := range list {
					sub[fmt.Sprintf("%d", idx)] = child
				}
			} else {
				if err := leaf.Decode(&sub); err != nil {
					return keys, err
				}
			}
			ret, err := scalarsIn(sub, append(parents, key))
			if err != nil {
				return keys, err
			}
			keys = append(keys, ret...)
		default:
			logrus.Fatalf("found unknown %v at %s", leaf.Kind, key)
		}
	}

	sort.Strings(keys)
	return keys, nil
}

func KeysFromYAML(data []byte) ([]string, error) {
	cfg := map[string]yaml.Node{}
	err := yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return scalarsIn(cfg, []string{})
}

// FromOP reads a config from an op item and returns a config.
func FromOP(item *op.Item) (*Config, error) {
	cfg := &Config{
		Vault: item.Vault.ID,
		Name:  item.Title,
		Tree:  NewEntry("root", yaml.MappingNode),
	}

	err := cfg.Tree.FromOP(item.Fields)
	return cfg, err
}

func (cfg *Config) MarshalYAML() (any, error) {
	return cfg.Tree.MarshalYAML()
}

func (cfg *Config) AsYAML(redacted bool) ([]byte, error) {
	redactOutput = redacted
	var out bytes.Buffer
	enc := yaml.NewEncoder(&out)
	enc.SetIndent(2)
	if err := enc.Encode(cfg); err != nil {
		return nil, fmt.Errorf("could not serialize config as yaml: %w", err)
	}
	return out.Bytes(), nil
}

func (cfg *Config) AsJSON(redacted bool, item bool) ([]byte, error) {
	var repr any
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

func (cfg *Config) Set(path []string, data []byte, isSecret, parseEntry bool) error {
	newEntry := NewEntry(path[len(path)-1], yaml.ScalarNode)
	newEntry.Path = path
	valueStr := string(data)
	newEntry.Value = valueStr

	if parseEntry {
		if err := yaml.Unmarshal(data, newEntry); err != nil {
			return err
		}
	} else {
		valueStr = strings.Trim(valueStr, "\n")
		if isSecret {
			newEntry.Style = yaml.TaggedStyle
			newEntry.Tag = YAMLTypeSecret
		}
		newEntry.Kind = yaml.ScalarNode
		newEntry.Value = valueStr

		if !strings.Contains(valueStr, "\n") {
			newEntry.Style &= yaml.LiteralStyle
		} else {
			newEntry.Style &= yaml.FlowStyle
		}
	}

	entry := cfg.Tree
	for idx, key := range path {
		if len(path)-1 == idx {
			dst := entry.ChildNamed(key)
			if dst == nil {
				if entry.Kind == yaml.MappingNode {
					key := NewEntry(key, yaml.ScalarNode)
					entry.Content = append(entry.Content, key, newEntry)
				} else {
					entry.Content = append(entry.Content, newEntry)
				}
			} else {
				logrus.Infof("setting %v", newEntry.Path)
				dst.Value = newEntry.Value
				dst.Tag = newEntry.Tag
				dst.Style = newEntry.Style
			}
			break
		}

		if child := entry.ChildNamed(key); child != nil {
			logrus.Infof("found child named %s, with len %v", key, len(child.Content))
			entry = child
			continue
		}

		logrus.Infof("no child named %s found in %s", key, entry.Name())
		kind := yaml.MappingNode
		if isNumeric(key) {
			kind = yaml.SequenceNode
		}
		sub := NewEntry(key, kind)
		sub.Path = append(entry.Path, key)
		entry.Content = append(entry.Content, sub)
		entry = sub
	}

	return nil
}
