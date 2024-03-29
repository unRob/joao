// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"git.rob.mx/nidito/chinampa/pkg/command"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// returns repo config details, error on failure to open/parse found files
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

func AutocompleteKeys(cmd *command.Command, currentValue, config string) ([]string, cobra.ShellCompDirective, error) {
	flag := cobra.ShellCompDirectiveError
	file := cmd.Arguments[0].ToString()
	buf, err := os.ReadFile(file)
	if err != nil {
		return nil, flag, fmt.Errorf("could not read file %s", file)
	}

	keys, err := KeysFromYAML(buf)
	if err != nil {
		return nil, flag, fmt.Errorf("could not parse file %s as %w", file, err)
	}

	sort.Strings(keys)

	return keys, cobra.ShellCompDirectiveDefault, nil
}

func AutocompleteKeysAndParents(cmd *command.Command, currentValue string, config string) (values []string, flag cobra.ShellCompDirective, err error) {
	opts := map[string]bool{".": true}
	options, flag, err := AutocompleteKeys(cmd, currentValue, "")
	for _, opt := range options {
		parts := strings.Split(opt, ".")
		sub := []string{parts[0]}
		for idx, p := range parts {
			key := strings.Join(sub, ".")
			opts[key] = true

			if idx > 0 && idx < len(parts)-1 {
				sub = append(sub, p)
			}
		}
	}

	for k := range opts {
		options = append(options, k)
	}
	sort.Strings(options)

	return options, flag, err
}
