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
	"bytes"

	_c "git.rob.mx/nidito/joao/internal/constants"
	"git.rob.mx/nidito/joao/internal/render"
	"git.rob.mx/nidito/joao/internal/runtime"
	"github.com/spf13/cobra"
)

type combinedCommand struct {
	Spec          *Command
	Command       *cobra.Command
	GlobalOptions Options
	HTMLOutput    bool
}

func (cmd *Command) HasAdditionalHelp() bool {
	return cmd.HelpFunc != nil
}

func (cmd *Command) AdditionalHelp(printLinks bool) *string {
	if cmd.HelpFunc != nil {
		str := cmd.HelpFunc(printLinks)
		return &str
	}
	return nil
}

func (cmd *Command) HelpRenderer(globalOptions Options) func(cc *cobra.Command, args []string) {
	return func(cc *cobra.Command, args []string) {
		// some commands don't have a binding until help is rendered
		// like virtual ones (sub command groups)
		cmd.SetCobra(cc)
		content, err := cmd.ShowHelp(globalOptions, args)
		if err != nil {
			panic(err)
		}
		_, err = cc.OutOrStderr().Write(content)
		if err != nil {
			panic(err)
		}
	}
}

func (cmd *Command) ShowHelp(globalOptions Options, args []string) ([]byte, error) {
	var buf bytes.Buffer
	c := &combinedCommand{
		Spec:          cmd,
		Command:       cmd.Cobra,
		GlobalOptions: globalOptions,
		HTMLOutput:    runtime.UnstyledHelpEnabled(),
	}
	err := _c.TemplateCommandHelp.Execute(&buf, c)
	if err != nil {
		return nil, err
	}

	colorEnabled := runtime.ColorEnabled()
	flags := cmd.Cobra.Flags()
	ncf := cmd.Cobra.Flag("no-color") // nolint:ifshort
	cf := cmd.Cobra.Flag("color")     // nolint:ifshort

	if noColorFlag, err := flags.GetBool("no-color"); err == nil && ncf.Changed {
		colorEnabled = !noColorFlag
	} else if colorFlag, err := flags.GetBool("color"); err == nil && cf.Changed {
		colorEnabled = colorFlag
	}

	content, err := render.Markdown(buf.Bytes(), colorEnabled)
	if err != nil {
		return nil, err
	}
	return content, nil
}
