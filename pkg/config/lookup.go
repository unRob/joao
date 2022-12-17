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
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func (cfg *Config) Lookup(query []string) (*Entry, error) {
	if len(query) == 0 || len(query) == 1 && query[0] == "." {
		return cfg.Tree, nil
	}

	entry := cfg.Tree
	for _, part := range query {
		entry = entry.ChildNamed(part)
		if entry == nil {
			return nil, fmt.Errorf("value not found at %s of %s", part, query)
		}
	}

	return entry, nil
}

func findRepoConfig(from string) (*opDetails, error) {
	parts := strings.Split(from, "/")
	for i := len(parts); i > 0; i-- {
		query := strings.Join(parts[0:i], "/")
		if bytes, err := os.ReadFile(query + "/.joao.yaml"); err == nil {
			rmc := &opDetails{Repo: query}
			err := yaml.Unmarshal(bytes, rmc)
			if err != nil {
				return nil, err
			}
			return rmc, nil
		}
	}

	return nil, nil
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
