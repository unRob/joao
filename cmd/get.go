// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
	"gopkg.in/yaml.v3"
)

var Get = &command.Command{
	Path:    []string{"get"},
	Summary: "retrieves configuration",
	Description: `
looks at the filesystem or remotely, using 1password (over the CLI if available, or 1password-connect, if configured).

` + "`--output`" + ` can be one of:
- **raw**:
  - when querying for scalar values this will return a non-quoted version of the values
  - when querying for trees or lists, this will output JSON
- **yaml**: formats the value at the given path as YAML
- **json**: formats the value at the given path as JSON
- **op**: formats the whole configuration as a 1Password item`,
	Arguments: command.Arguments{
		{
			Name:        "config",
			Description: "The configuration to get from",
			Required:    true,
			Values: &command.ValueSource{
				Files: &[]string{"yaml", "yml"},
			},
		},
		{
			Name:        "path",
			Default:     ".",
			Description: "A dot-delimited path to extract from CONFIG",
			Values: &command.ValueSource{
				Func: config.AutocompleteKeysAndParents,
			},
		},
	},
	Options: command.Options{
		"output": {
			ShortName:   "o",
			Description: "the format to use for rendering output",
			Default:     "raw",
			Values: &command.ValueSource{
				Static: &[]string{"raw", "json", "yaml", "diff-yaml", "op"},
			},
		},
		"redacted": {
			Description: "Do not print secret values",
			Type:        "bool",
			Default:     false,
		},
		"remote": {
			Description: "Get values from 1password",
			Type:        "bool",
			Default:     false,
		},
	},
	Action: func(cmd *command.Command) error {
		path := cmd.Arguments[0].ToValue().(string)
		query := cmd.Arguments[1].ToValue().(string)

		remote := cmd.Options["remote"].ToValue().(bool)
		format := cmd.Options["output"].ToValue().(string)
		redacted := cmd.Options["redacted"].ToValue().(bool)

		cfg, err := config.Load(path, remote)
		if err != nil {
			return err
		}

		if query == "" || query == "." {
			switch format {
			case "yaml", "raw", "diff-yaml":
				modes := []config.OutputMode{}
				if redacted {
					modes = append(modes, config.OutputModeRedacted)
				}
				if format == "diff-yaml" {
					modes = append(modes, config.OutputModeNoComments, config.OutputModeSorted)
				}
				bytes, err := cfg.AsYAML(modes...)
				if err != nil {
					return err
				}
				_, err = cmd.Cobra.OutOrStdout().Write(bytes)
				return err
			case "json", "op":
				bytes, err := cfg.AsJSON(redacted, format == "op")
				if err != nil {
					return err
				}
				_, err = cmd.Cobra.OutOrStdout().Write(bytes)
				return err
			}
			return fmt.Errorf("unknown format %s", format)
		}

		parts := strings.Split(query, ".")

		entry := cfg.Tree
		for _, part := range parts {
			entry = entry.ChildNamed(part)
			if entry == nil {
				return fmt.Errorf("value not found at %s of %s", part, query)
			}
		}

		var bytes []byte
		if len(entry.Content) > 0 {
			val := entry.AsMap()
			if format == "yaml" {
				enc := yaml.NewEncoder(cmd.Cobra.OutOrStdout())
				enc.SetIndent(2)
				return enc.Encode(val)
			}

			bytes, err = json.Marshal(val)
			if err != nil {
				return err
			}
		} else {
			bytes = []byte(entry.String())
		}

		_, err = cmd.Cobra.OutOrStdout().Write(bytes)
		return err
	},
}
