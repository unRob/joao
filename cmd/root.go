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
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var Root = &cobra.Command{
	Use:   "joao [--silent|-v|--verbose] [--[no-]color] [-h|--help] [--version]",
	Short: "does config",
	Long:  `does config with 1password and stuff`,
	// DisableAutoGenTag: true,
	// SilenceUsage:      true,
	// SilenceErrors:     true,
	ValidArgs:   []string{""},
	Annotations: map[string]string{},
	Args: func(cmd *cobra.Command, args []string) error {
		err := cobra.OnlyValidArgs(cmd, args)
		if err != nil {

			suggestions := []string{}
			bold := color.New(color.Bold)
			for _, l := range cmd.SuggestionsFor(args[len(args)-1]) {
				suggestions = append(suggestions, bold.Sprint(l))
			}
			errMessage := fmt.Sprintf("Unknown subcommand %s", bold.Sprint(strings.Join(args, " ")))
			if len(suggestions) > 0 {
				errMessage += ". Perhaps you meant " + strings.Join(suggestions, ", ") + "?"
			}
			return fmt.Errorf("command not found")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			if ok, err := cmd.Flags().GetBool("version"); err == nil && ok {
				vc, _, err := cmd.Root().Find([]string{"__version"})

				if err != nil {
					return err
				}
				return vc.RunE(vc, []string{})
			}
			return fmt.Errorf("no command provided")

		}

		return nil
	},
}

func RootCommand(version string) *cobra.Command {
	Root.Annotations["version"] = version
	rootFlagset := pflag.NewFlagSet("joao", pflag.ContinueOnError)
	// for name, opt := range Root.Options {
	// 	def, ok := opt.Default.(bool)
	// 	if !ok {
	// 		def = false
	// 	}

	// 	if opt.ShortName != "" {
	// 		rootFlagset.BoolP(name, opt.ShortName, def, opt.Description)
	// 	} else {
	// 		rootFlagset.Bool(name, def, opt.Description)
	// 	}
	// }

	rootFlagset.Usage = func() {}
	rootFlagset.SortFlags = false
	Root.PersistentFlags().AddFlagSet(rootFlagset)

	Root.Flags().Bool("version", false, "Display the version")

	// Root.CompletionOptions.DisableDefaultCmd = true

	Root.AddCommand(getCommand)
	// Root.AddCommand(completionCommand)
	// Root.AddCommand(generateDocumentationCommand)
	// Root.AddCommand(doctorCommand)

	// Root.SetHelpCommand(helpCommand)
	// helpCommand.AddCommand(docsCommand)
	// docsCommand.SetHelpFunc(docs.HelpRenderer(Root.Options))
	// Root.SetHelpFunc(Root.HelpRenderer(Root.Options))

	return Root
}
