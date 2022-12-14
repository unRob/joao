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
package command

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type HelpFunc func(printLinks bool) string
type Action func(cmd *Command) error

type Command struct {
	Path []string
	// Summary is a short description of a command, on supported shells this is part of the autocomplete prompt
	Summary string `json:"summary" yaml:"summary" validate:"required"`
	// Description is a long form explanation of how a command works its magic. Markdown is supported
	Description string `json:"description" yaml:"description" validate:"required"`
	// A list of arguments for a command
	Arguments Arguments `json:"arguments" yaml:"arguments" validate:"dive"`
	// A map of option names to option definitions
	Options  Options  `json:"options" yaml:"options" validate:"dive"`
	HelpFunc HelpFunc `json:"-" yaml:"-"`
	// The action to take upon running
	Action       Action
	runtimeFlags *pflag.FlagSet
	Cobra        *cobra.Command
}

func (cmd *Command) SetBindings() *Command {
	ptr := cmd
	for _, opt := range cmd.Options {
		opt.Command = ptr
		if opt.Validates() {
			opt.Values.command = ptr
		}
	}

	for _, arg := range cmd.Arguments {
		arg.Command = ptr
		if arg.Validates() {
			arg.Values.command = ptr
		}
	}
	return ptr
}

func (cmd *Command) Name() string {
	return cmd.Path[len(cmd.Path)-1]
}

func (cmd *Command) FullName() string {
	return strings.Join(cmd.Path, " ")
}

func (cmd *Command) FlagSet() *pflag.FlagSet {
	if cmd.runtimeFlags == nil {
		fs := pflag.NewFlagSet(strings.Join(cmd.Path, " "), pflag.ContinueOnError)
		fs.SortFlags = false
		fs.Usage = func() {}

		for name, opt := range cmd.Options {
			switch opt.Type {
			case ValueTypeBoolean:
				def := false
				if opt.Default != nil {
					def = opt.Default.(bool)
				}
				fs.Bool(name, def, opt.Description)
			case ValueTypeDefault, ValueTypeString:
				opt.Type = ValueTypeString
				def := ""
				if opt.Default != nil {
					def = fmt.Sprintf("%s", opt.Default)
				}
				fs.String(name, def, opt.Description)
			default:
				// ignore flag
				logrus.Warnf("Ignoring unknown option type <%s> for option <%s>", opt.Type, name)
				continue
			}
		}

		cmd.runtimeFlags = fs
	}
	return cmd.runtimeFlags
}

func (cmd *Command) ParseInput(cc *cobra.Command, args []string) error {
	cmd.Arguments.Parse(args)
	skipValidation, _ := cc.Flags().GetBool("skip-validation")
	cmd.Options.Parse(cc.Flags())
	if !skipValidation {
		logrus.Debug("Validating arguments")
		if err := cmd.Arguments.AreValid(); err != nil {
			return err
		}

		logrus.Debug("Validating flags")
		if err := cmd.Options.AreValid(); err != nil {
			return err
		}
	}

	return nil
}

func (cmd *Command) Run(cc *cobra.Command, args []string) error {
	logrus.Debugf("running command %s", cmd.FullName())

	if err := cmd.ParseInput(cc, args); err != nil {
		return err
	}

	return cmd.Action(cmd)
}

func (cmd *Command) SetCobra(cc *cobra.Command) {
	cmd.Cobra = cc
}
