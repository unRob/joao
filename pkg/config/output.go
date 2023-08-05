// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"

	opClient "git.rob.mx/nidito/joao/internal/op-client"
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
	// OutputModeStandardYAML formats strings and arrays uniformly.
	OutputModeStandardYAML OutputMode = 16
)

var defaultYamlOutput = &outputOptions{OutputModeRoundTrip}
var yamlOutput = defaultYamlOutput

func setOutputMode(modes []OutputMode) func() {
	if len(modes) > 0 {
		yamlOutput = &outputOptions{}
		yamlOutput.Set(modes...)
	}
	return func() { yamlOutput = defaultYamlOutput }
}

// ToMap turns a config into a dictionary of strings to values.
func (cfg *Config) ToMap(modes ...OutputMode) map[string]any {
	defer setOutputMode(modes)()
	ret := map[string]any{}
	for _, child := range cfg.Tree.Content {
		if child.Name() == "" || (yamlOutput.Has(OutputModeNoConfig) && child.Name() == "_config") {
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
	cs := opClient.Checksum(datafields)

	fields[0].Value = cs
	fields = append(fields, datafields...)

	for i := 0; i < len(cfg.Tree.Content); i += 2 {
		value := cfg.Tree.Content[i+1]
		if value.Type == YAMLTypeMetaConfig {
			continue
		}

		if value.Kind == yaml.MappingNode || value.Kind == yaml.SequenceNode {
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

// MarshalYAML implements `yaml.Marshal“.
func (cfg *Config) MarshalYAML() (any, error) {
	return cfg.Tree.MarshalYAML()
}

// AsYAML returns the config encoded as YAML.
func (cfg *Config) AsYAML(modes ...OutputMode) ([]byte, error) {
	defer setOutputMode(modes)()
	logrus.Debugf("Printing as yaml with modes %v", yamlOutput)

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
		modes := []OutputMode{OutputModeNoConfig}
		if redacted {
			modes = append(modes, OutputModeRedacted)
		}

		repr = cfg.ToMap(modes...)
	}

	bytes, err := json.Marshal(repr)
	if err != nil {
		return nil, fmt.Errorf("could not serialize config as json: %w", err)
	}
	return bytes, nil
}

func (cfg *Config) AsFile(path string, modes ...OutputMode) error {
	b, err := cfg.AsYAML(modes...)
	if err != nil {
		return err
	}

	var mode fs.FileMode = 0644
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}

	if err := os.WriteFile(path, b, mode); err != nil {
		return fmt.Errorf("could not save config to file %s: %w", path, err)
	}

	return nil
}
