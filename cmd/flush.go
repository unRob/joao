// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"fmt"

	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
	opclient "git.rob.mx/nidito/joao/pkg/op-client"
	"github.com/sirupsen/logrus"
)

var Flush = &command.Command{
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
				Files: &fileExtensions,
			},
		},
	},
	Options: command.Options{
		"dry-run": {
			Description: "Don't persist to 1Password",
			Type:        "bool",
		},
		"redact": {
			Description: "Redact local file after flushing",
			Type:        "bool",
		},
	},
	Action: func(cmd *command.Command) error {
		paths := cmd.Arguments[0].ToValue().([]string)
		dryRun := cmd.Options["dry-run"].ToValue().(bool)

		if dryRun {
			opclient.Use(&opclient.CLI{DryRun: true})
		}

		for _, path := range paths {
			cfg, err := config.Load(path, false)
			if err != nil {
				return err
			}

			if dryRun {
				logrus.Warnf("dry-run: comparing %s to %s", path, cfg.OPURL())
				if err := cfg.DiffRemote(path, false, false, cmd.Cobra.OutOrStdout(), cmd.Cobra.OutOrStderr()); err != nil {
					return err
				}
				logrus.Warnf("dry-run: did not update %s", cfg.OPURL())
				continue
			}

			if err := opclient.Update(cfg.Vault, cfg.Name, cfg.ToOP()); err != nil {
				return fmt.Errorf("could not flush to 1password: %w", err)
			}

			if cmd.Options["redact"].ToValue().(bool) {
				if err := cfg.AsFile(path, config.OutputModeRedacted); err != nil {
					return err
				}
			}
			logrus.Infof("Flushed %s to %s", path, cfg.OPURL())
		}

		logrus.Info("Done")
		return nil
	},
}
