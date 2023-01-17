// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

var Fetch = &command.Command{
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
				Files: &fileExtensions,
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
			local, err := config.Load(path, false)
			if err != nil {
				return err
			}

			if dryRun := cmd.Options["dry-run"].ToValue().(bool); dryRun {
				logrus.Warnf("dry-run: comparing %s to %s", local.OPURL(), path)
				if err := local.DiffRemote(path, false, true, cmd.Cobra.OutOrStdout(), cmd.Cobra.OutOrStderr()); err != nil {
					return err
				}
				logrus.Warnf("dry-run: did not update %s", path)
				continue
			}

			remote, err := config.Load(path, true)
			if err != nil {
				return err
			}

			if err = local.Merge(remote); err != nil {
				return err
			}

			if err := local.AsFile(path); err != nil {
				return err
			}

			logrus.Infof("Fetched %s => %s", remote.OPURL(), path)
		}

		logrus.Info("Done")
		return nil
	},
}
