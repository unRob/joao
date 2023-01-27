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
	Description: `Fetches remote and compares against local, ignoring comments but respecting order. The diff output shows what would happen upon running ﹅joao fetch﹅. Specify ﹅--remote﹅ to show what would happen upon ﹅joao flush﹅`,
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
			Type:        command.ValueTypeString,
			Default:     "auto",
			Values: &command.ValueSource{
				Static: &[]string{
					"auto", "patch", "exit-code", "short",
				},
			},
		},
		"remote": {
			Description: "Shows what would happen on `flush` instead of `fetch`",
			Type:        command.ValueTypeBoolean,
			Default:     false,
		},
		"redacted": {
			Description: "Compare redacted versions",
			Type:        command.ValueTypeBoolean,
			Default:     false,
		},
	},
	Action: func(cmd *command.Command) error {
		paths := cmd.Arguments[0].ToValue().([]string)
		redacted := cmd.Options["redacted"].ToValue().(bool)
		remote := cmd.Options["remote"].ToValue().(bool)
		for _, path := range paths {

			local, err := config.Load(path, false)
			if err != nil {
				return err
			}

			if err := local.DiffRemote(path, redacted, remote, cmd.Cobra.OutOrStdout(), cmd.Cobra.OutOrStderr()); err != nil {
				return err
			}
		}

		logrus.Info("Done")
		return nil
	},
}
