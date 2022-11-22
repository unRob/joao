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
package registry

import (
	"strings"

	"git.rob.mx/nidito/joao/internal/command"
	_c "git.rob.mx/nidito/joao/internal/constants"
	"git.rob.mx/nidito/joao/internal/errors"
	"git.rob.mx/nidito/joao/internal/runtime"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func toCobra(cmd *command.Command, globalOptions command.Options) *cobra.Command {
	localName := cmd.Name()
	useSpec := []string{localName, "[options]"}
	for _, arg := range cmd.Arguments {
		useSpec = append(useSpec, arg.ToDesc())
	}

	cc := &cobra.Command{
		Use:               strings.Join(useSpec, " "),
		Short:             cmd.Summary,
		DisableAutoGenTag: true,
		SilenceUsage:      true,
		SilenceErrors:     true,
		Annotations: map[string]string{
			_c.ContextKeyRuntimeIndex: cmd.FullName(),
		},
		Args: func(cc *cobra.Command, supplied []string) error {
			skipValidation, _ := cc.Flags().GetBool("skip-validation")
			if !skipValidation && runtime.ValidationEnabled() {
				cmd.Arguments.Parse(supplied)
				return cmd.Arguments.AreValid()
			}
			return nil
		},
		RunE: cmd.Run,
	}

	cc.SetFlagErrorFunc(func(c *cobra.Command, e error) error {
		return errors.BadArguments{Msg: e.Error()}
	})

	cc.ValidArgsFunction = cmd.Arguments.CompletionFunction

	cc.Flags().AddFlagSet(cmd.FlagSet())

	for name, opt := range cmd.Options {
		if err := cc.RegisterFlagCompletionFunc(name, opt.CompletionFunction); err != nil {
			logrus.Errorf("Failed setting up autocompletion for option <%s> of command <%s>", name, cmd.FullName())
		}
	}

	cc.SetHelpFunc(cmd.HelpRenderer(globalOptions))
	cmd.SetCobra(cc)
	return cc
}

func fromCobra(cc *cobra.Command) *command.Command {
	rtidx, hasAnnotation := cc.Annotations[_c.ContextKeyRuntimeIndex]
	if hasAnnotation {
		return Get(rtidx)
	}
	return nil
}
