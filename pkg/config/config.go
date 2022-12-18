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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
	"gopkg.in/yaml.v3"
)

const YAMLTypeSecret string = "!!secret"
const YAMLTypeMetaConfig string = "!!joao"

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

type Config struct {
	Vault string
	Name  string
	Tree  *Entry
}

// ToMap turns a config into a dictionary of strings to values.
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

func checksum(fields []*op.ItemField) string {
	newHash, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	// newHash := md5.New()
	df := []string{}
	for _, field := range fields {
		if field.ID == "password" || field.ID == "notesPlain" || (field.Section != nil && field.Section.ID == "~annotations") {
			continue
		}
		label := field.Label
		if field.Section != nil && field.Section.ID != "" {
			label = field.Section.ID + "." + label
		}
		df = append(df, label+field.Value)
	}
	sort.Strings(df)
	newHash.Write([]byte(strings.Join(df, "")))
	checksum := newHash.Sum(nil)

	return fmt.Sprintf("%x", checksum)
}

// ToOp turns a config into an 1Password Item.
func (cfg *Config) ToOP() *op.Item {
	sections := []*op.ItemSection{annotationsSection}
	fields := append([]*op.ItemField{}, defaultItemFields...)

	datafields := cfg.Tree.ToOP()
	cs := checksum(datafields)

	fields[0].Value = cs
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

// MarshalYAML implements `yaml.Marshal``.
func (cfg *Config) MarshalYAML() (any, error) {
	return cfg.Tree.MarshalYAML()
}

// AsYAML returns the config encoded as YAML.
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

// AsJSON returns the config enconded as JSON, optionally encoding as a 1Password item.
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

// Delete a value at path.
func (cfg *Config) Delete(path []string) error {
	parent := cfg.Tree

	for idx, key := range path {
		if len(path)-1 == idx {
			newContents := []*Entry{}
			found := false
			for idx, child := range parent.Content {
				if child.Name() == key {
					found = true
					logrus.Debugf("Deleting %s", strings.Join(path, "."))
					if parent.Kind == yaml.DocumentNode || parent.Kind == yaml.MappingNode {
						newContents = newContents[0 : idx-1]
					}
					continue
				}
				newContents = append(newContents, child)
			}

			if !found {
				return fmt.Errorf("no value found at %s", key)
			}

			parent.Content = newContents
			break
		}

		parent = parent.ChildNamed(key)
		if parent == nil {
			return fmt.Errorf("no value found at %s", key)
		}
	}

	return nil
}

// Set a new value, optionally parsing the supplied bytes as a secret or a JSON-encoded value.
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
			entry = child
			continue
		}

		kind := yaml.MappingNode
		if isNumeric(key) {
			kind = yaml.SequenceNode
		}
		sub := NewEntry(key, kind)
		sub.Path = append(entry.Path, key) // nolint: gocritic
		entry.Content = append(entry.Content, sub)
		entry = sub
	}

	return nil
}

func (cfg *Config) Merge(other *Config) error {
	return cfg.Tree.Merge(other.Tree)
}
