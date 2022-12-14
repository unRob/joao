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
package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	"git.rob.mx/nidito/joao/internal/command"
	opclient "git.rob.mx/nidito/joao/internal/op-client"
	"git.rob.mx/nidito/joao/internal/registry"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

func init() {
	registry.Register(setCommand)
}

var setCommand = (&command.Command{
	Path:    []string{"set"},
	Summary: "updates configuration values",
	Description: `
looks at the filesystem or remotely, using 1password (over the CLI if available, or 1password-connect, if configured).

Will read from stdin (or ﹅--from﹅ a file) and store it at the ﹅PATH
﹅ of ﹅CONFIG﹅, optionally ﹅--flush﹅ing to 1Password.`,
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
				Func:       keyFinder,
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
		"json": {
			Description: "Treat input as JSON-encoded",
			Type:        "bool",
		},
		"flush": {
			Description: "Save to 1Password after saving to file",
			Type:        "bool",
		},
	},
	Action: func(cmd *command.Command) error {
		path := cmd.Arguments[0].ToValue().(string)
		query := cmd.Arguments[1].ToValue().(string)

		var cfg *config.Config
		var err error
		secret := cmd.Options["secret"].ToValue().(bool)
		input := cmd.Options["input"].ToValue().(string)
		parseJSON := cmd.Options["json"].ToValue().(bool)
		flush := cmd.Options["flush"].ToValue().(bool)

		cfg, err = config.Load(path, false)
		if err != nil {
			return err
		}

		parts := strings.Split(query, ".")

		valueBytes, err := os.ReadFile(input)
		if err != nil {
			return err
		}

		if err := cfg.Set(parts, valueBytes, secret, parseJSON); err != nil {
			return err
		}

		// b, err := cfg.AsJSON(false, true)
		b, err := cfg.AsYAML(false)
		if err != nil {
			return err
		}

		var mode fs.FileMode = 644
		// var mode uint32 =
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
