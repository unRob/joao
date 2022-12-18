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

	"git.rob.mx/nidito/joao/internal/command"
	"git.rob.mx/nidito/joao/internal/registry"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

func init() {
	registry.Register(fetchCommand)
}

var fetchCommand = (&command.Command{
	Path:        []string{"fetch"},
	Summary:     "fetches configuration values from 1Password",
	Description: `Fetches secrets for local ﹅CONFIG﹅ files from 1Password.`,
	Arguments: command.Arguments{
		{
			Name:        "config",
			Description: "The configuration file(s) to fetch",
			Required:    false,
			Variadic:    true,
			Values: &command.ValueSource{
				Files: &[]string{"joao.yaml"},
			},
		},
	},
	Options: command.Options{
		"dry-run": {
			Description: "Don't persist to the filesystem",
			Type:        "bool",
		},
	},
	Action: func(cmd *command.Command) error {
		paths := cmd.Arguments[0].ToValue().([]string)
		for _, path := range paths {
			remote, err := config.Load(path, true)
			if err != nil {
				return err
			}
			local, err := config.Load(path, false)
			if err != nil {
				return err
			}

			if err = local.Merge(remote); err != nil {
				return err
			}

			b, err := local.AsYAML(false)
			if err != nil {
				return err
			}

			if dryRun := cmd.Options["dry-run"].ToValue().(bool); dryRun {
				logrus.Warnf("dry-run: would write to %s", path)
				_, _ = cmd.Cobra.OutOrStdout().Write(b)
			} else {
				var mode fs.FileMode = 0644
				if info, err := os.Stat(path); err == nil {
					mode = info.Mode().Perm()
				}

				if err := os.WriteFile(path, b, mode); err != nil {
					return fmt.Errorf("could not save changes to %s: %w", path, err)
				}
			}

			logrus.Infof("Updated %s", path)
		}

		logrus.Info("Done")
		return nil
	},
}).SetBindings()
