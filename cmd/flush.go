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

	"git.rob.mx/nidito/joao/internal/command"
	opclient "git.rob.mx/nidito/joao/internal/op-client"
	"git.rob.mx/nidito/joao/internal/registry"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

func init() {
	registry.Register(flushCommand)
}

var flushCommand = (&command.Command{
	Path:        []string{"flush"},
	Summary:     "flush configuration values to 1Password",
	Description: `Creates or updates existing items for every ﹅CONFIG﹅ file provided. Does not delete 1Password items.`,
	Arguments: command.Arguments{
		{
			Name:        "config",
			Description: "The configuration file(s) to flush",
			Required:    false,
			Variadic:    true,
			Values: &command.ValueSource{
				Files: &[]string{"yaml"},
			},
		},
	},
	Options: command.Options{
		"dry-run": {
			Description: "Don't persist to 1Password",
			Type:        "bool",
		},
	},
	Action: func(cmd *command.Command) error {
		paths := cmd.Arguments[0].ToValue().([]string)

		if dryRun := cmd.Options["dry-run"].ToValue().(bool); dryRun {
			opclient.Use(&opclient.CLI{DryRun: true})
		}

		for _, path := range paths {
			cfg, err := config.Load(path, false)
			if err != nil {
				return err
			}

			if err := opclient.Update(cfg.Vault, cfg.Name, cfg.ToOP()); err != nil {
				return fmt.Errorf("could not flush to 1password: %w", err)
			}
		}

		logrus.Info("Done")
		return nil
	},
}).SetBindings()
