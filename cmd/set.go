// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"git.rob.mx/nidito/chinampa"
	"git.rob.mx/nidito/chinampa/pkg/command"
	opclient "git.rob.mx/nidito/joao/internal/op-client"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

func init() {
	chinampa.Register(setCommand)
}

var setCommand = (&command.Command{
	Path:    []string{"set"},
	Summary: "updates configuration values",
	Description: `
Updates the value at ﹅PATH﹅ in a local ﹅CONFIG﹅ file. Specify ﹅--secret﹅ to keep the value secret, or ﹅--delete﹅ to delete the key at PATH.

Will read values from stdin (or ﹅--from﹅ a file) and store it at the ﹅PATH﹅ of ﹅CONFIG﹅, optionally ﹅--flush﹅ing to 1Password.
`,
	Arguments: command.Arguments{
		{
			Name:        "config",
			Description: "The configuration file to modify",
			Required:    true,
			Values: &command.ValueSource{
				Files: &[]string{"yaml"},
			},
		},
		{
			Name:        "path",
			Required:    true,
			Description: "A dot-delimited path to set in CONFIG",
			Values: &command.ValueSource{
				SuggestRaw: true,
				Suggestion: true,
				Func:       config.AutocompleteKeys,
			},
		},
	},
	Options: command.Options{
		"input": {
			ShortName:   "i",
			Description: "the file to read input from",
			Default:     "/dev/stdin",
			Values: &command.ValueSource{
				Files: &[]string{},
			},
		},
		"secret": {
			Description: "Store value as a secret string",
			Type:        "bool",
		},
		"delete": {
			Description: "Delete the value at the given PATH",
			Type:        "bool",
		},
		"json": {
			Description: "Treat input as JSON-encoded",
			Type:        "bool",
		},
		"flush": {
			Description: "Save to 1Password after saving to PATH",
			Type:        "bool",
		},
	},
	Action: func(cmd *command.Command) error {
		path := cmd.Arguments[0].ToValue().(string)
		query := cmd.Arguments[1].ToValue().(string)

		var cfg *config.Config
		var err error
		secret := cmd.Options["secret"].ToValue().(bool)
		delete := cmd.Options["delete"].ToValue().(bool)
		input := cmd.Options["input"].ToValue().(string)
		parseJSON := cmd.Options["json"].ToValue().(bool)
		flush := cmd.Options["flush"].ToValue().(bool)

		if secret && delete {
			return fmt.Errorf("cannot --delete and set a --secret at the same time")
		}

		if secret && parseJSON {
			return fmt.Errorf("cannot set a --secret that is JSON encoded, encode individual values instead")
		}

		if delete && input != "/dev/stdin" {
			logrus.Warn("Ignoring --file while deleting")
		}

		cfg, err = config.Load(path, false)
		if err != nil {
			return err
		}

		parts := strings.Split(query, ".")

		if delete {
			if err := cfg.Delete(parts); err != nil {
				return err
			}
		} else {
			valueBytes, err := os.ReadFile(input)
			if err != nil {
				return err
			}
			if err := cfg.Set(parts, valueBytes, secret, parseJSON); err != nil {
				return err
			}
		}

		b, err := cfg.AsYAML()
		if err != nil {
			return err
		}

		var mode fs.FileMode = 0644
		if info, err := os.Stat(path); err == nil {
			mode = info.Mode().Perm()
		}

		if err := os.WriteFile(path, b, mode); err != nil {
			return fmt.Errorf("could not save changes to %s: %w", path, err)
		}

		if flush {
			if err := opclient.Update(cfg.Vault, cfg.Name, cfg.ToOP()); err != nil {
				return fmt.Errorf("could not flush to 1password: %w", err)
			}
		}

		logrus.Info("Done")
		return err
	},
}).SetBindings()
