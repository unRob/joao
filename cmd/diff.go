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
	"git.rob.mx/nidito/chinampa"
	"git.rob.mx/nidito/chinampa/pkg/command"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/sirupsen/logrus"
)

func init() {
	chinampa.Register(diffCommand)
}

var diffCommand = (&command.Command{
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
				Files: &[]string{"joao.yaml"},
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
}).SetBindings()
