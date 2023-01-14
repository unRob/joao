// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

var Diff = &command.Command{
	Path:        []string{"diff"},
	Summary:     "Shows differences between local and remote configs",
	Description: `Fetches remote and compares against local, ignoring comments but respecting order.`,
	Arguments: command.Arguments{
		{
			Name:        "config",
			Description: "The configuration file(s) to diff",
			Required:    false,
			Variadic:    true,
			Values: &command.ValueSource{
				Files: &fileExtensions,
			},
		},
	},
	Options: command.Options{
		"output": {
			Description: "How to format the differences",
			Type:        "string",
			Default:     "auto",
			Values: &command.ValueSource{
				Static: &[]string{
					"auto", "patch", "exit-code", "short",
				},
			},
		},
	},
	Action: func(cmd *command.Command) error {
		paths := cmd.Arguments[0].ToValue().([]string)
		for _, path := range paths {

			local, err := config.Load(path, false)
			if err != nil {
				return err
			}

			if err := local.DiffRemote(path, cmd.Cobra.OutOrStdout(), cmd.Cobra.OutOrStderr()); err != nil {
				return err
			}
		}

		logrus.Info("Done")
		return nil
	},
}
