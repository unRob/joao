// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"
	"io/fs"
	"os"

	"git.rob.mx/nidito/chinampa"
	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

func init() {
	chinampa.Register(fetchCommand)
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

			b, err := local.AsYAML()
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
