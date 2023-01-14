// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

var Redact = &command.Command{
	Path:        []string{"redact"},
	Summary:     "removes secrets from configuration",
	Description: `Removes secret values (not the keys) from existing items for every ﹅CONFIG﹅ file provided.`,
	Arguments: command.Arguments{
		{
			Name:        "config",
			Description: "The configuration file(s) to redact",
			Required:    false,
			Variadic:    true,
			Values: &command.ValueSource{
				Files: &fileExtensions,
			},
		},
	},
	Action: func(cmd *command.Command) error {
		paths := cmd.Arguments[0].ToValue().([]string)

		for _, path := range paths {
			cfg, err := config.Load(path, false)
			if err != nil {
				return err
			}

			if err := cfg.AsFile(path); err != nil {
				return err
			}
		}

		logrus.Info("Done")
		return nil
	},
}
