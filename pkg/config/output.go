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

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const YAMLTypeSecret string = "!!secret"
const YAMLTypeMetaConfig string = "!!joao"

type outputOptions struct {
	mode OutputMode
}

func (flag *outputOptions) Set(modes ...OutputMode) {
	for _, mode := range modes {
		flag.mode = mode | flag.mode
	}
}
func (flag *outputOptions) Clear(mode OutputMode)    { flag.mode = mode &^ flag.mode }
func (flag *outputOptions) Toggle(mode OutputMode)   { flag.mode = mode ^ flag.mode }
func (flag *outputOptions) Has(mode OutputMode) bool { return mode&flag.mode != 0 }

type OutputMode uint8

const (
	// OutputModeRoundTrip outputs the input as-is.
	OutputModeRoundTrip OutputMode = iota
	// OutputModeRedacted prints empty secret values.
	OutputModeRedacted
	// OutputModeNoComments does not output comments.
	OutputModeNoComments OutputMode = 2
	// OutputModeSorted outputs map keys in alphabetical order.
	OutputModeSorted OutputMode = 4
	// OutputModeNoConfig does not output the _config key if any.
	OutputModeNoConfig OutputMode = 8
)

var defaultYamlOutput = &outputOptions{OutputModeRoundTrip}
var yamlOutput = defaultYamlOutput

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
func (cfg *Config) AsYAML(modes ...OutputMode) ([]byte, error) {
	if len(modes) > 0 {
		defer func() { yamlOutput = defaultYamlOutput }()
		yamlOutput = &outputOptions{}
		yamlOutput.Set(modes...)
	}

	logrus.Debug("Printing as yaml with modes %v", modes)

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
		repr = cfg.ToMap()
	}

	bytes, err := json.Marshal(repr)
	if err != nil {
		return nil, fmt.Errorf("could not serialize config as json: %w", err)
	}
	return bytes, nil
}
