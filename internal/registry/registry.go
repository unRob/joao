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
	"fmt"
	"os"
	"sort"
	"strings"

	"git.rob.mx/nidito/joao/internal/command"
	_c "git.rob.mx/nidito/joao/internal/constants"
	"git.rob.mx/nidito/joao/internal/errors"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var registry = &CommandRegistry{
	kv: map[string]*command.Command{},
}

type ByPath []*command.Command

func (cmds ByPath) Len() int           { return len(cmds) }
func (cmds ByPath) Swap(i, j int)      { cmds[i], cmds[j] = cmds[j], cmds[i] }
func (cmds ByPath) Less(i, j int) bool { return cmds[i].FullName() < cmds[j].FullName() }

type CommandTree struct {
	Command  *command.Command `json:"command"`
	Children []*CommandTree   `json:"children"`
}

func (t *CommandTree) Traverse(fn func(cmd *command.Command) error) error {
	for _, child := range t.Children {
		if err := fn(child.Command); err != nil {
			return err
		}

		if err := child.Traverse(fn); err != nil {
			return err
		}
	}
	return nil
}

type CommandRegistry struct {
	kv     map[string]*command.Command
	byPath []*command.Command
	tree   *CommandTree
}

func Register(cmd *command.Command) {
	logrus.Debugf("Registering %s", cmd.FullName())
	registry.kv[cmd.FullName()] = cmd
}

func Get(id string) *command.Command {
	return registry.kv[id]
}

func CommandList() []*command.Command {
	if len(registry.byPath) == 0 {
		list := []*command.Command{}
		for _, v := range registry.kv {
			list = append(list, v)
		}
		sort.Sort(ByPath(list))
		registry.byPath = list
	}

	return registry.byPath
}

func BuildTree(cc *cobra.Command, depth int) {
	tree := &CommandTree{
		Command:  fromCobra(cc),
		Children: []*CommandTree{},
	}

	var populateTree func(cmd *cobra.Command, ct *CommandTree, maxDepth int, depth int)
	populateTree = func(cmd *cobra.Command, ct *CommandTree, maxDepth int, depth int) {
		newDepth := depth + 1
		for _, subcc := range cmd.Commands() {
			if subcc.Hidden {
				continue
			}

			if cmd := fromCobra(subcc); cmd != nil {
				leaf := &CommandTree{Children: []*CommandTree{}}
				leaf.Command = cmd
				ct.Children = append(ct.Children, leaf)

				if newDepth < maxDepth {
					populateTree(subcc, leaf, maxDepth, newDepth)
				}
			}
		}
	}
	populateTree(cc, tree, depth, 0)

	registry.tree = tree
}

func SerializeTree(serializationFn func(any) ([]byte, error)) (string, error) {
	bytes, err := serializationFn(registry.tree)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ChildrenNames() []string {
	if registry.tree == nil {
		return []string{}
	}

	ret := make([]string, len(registry.tree.Children))
	for idx, cmd := range registry.tree.Children {
		ret[idx] = cmd.Command.Name()
	}
	return ret
}

func Execute(version string) error {
	cmdRoot := command.Root
	ccRoot.Short = cmdRoot.Summary
	ccRoot.Long = cmdRoot.Description
	ccRoot.Annotations["version"] = version
	ccRoot.CompletionOptions.DisableDefaultCmd = true
	ccRoot.Flags().AddFlagSet(cmdRoot.FlagSet())

	for name, opt := range cmdRoot.Options {
		if err := ccRoot.RegisterFlagCompletionFunc(name, opt.CompletionFunction); err != nil {
			logrus.Errorf("Failed setting up autocompletion for option <%s> of command <%s>", name, cmdRoot.FullName())
		}
	}
	ccRoot.SetHelpFunc(cmdRoot.HelpRenderer(cmdRoot.Options))

	for _, cmd := range CommandList() {
		cmd := cmd
		leaf := toCobra(cmd, cmdRoot.Options)
		container := ccRoot
		for idx, cp := range cmd.Path {
			if idx == len(cmd.Path)-1 {
				logrus.Debugf("adding command %s to %s", leaf.Name(), cmd.Path[0:idx])
				container.AddCommand(leaf)
				break
			}

			query := []string{cp}
			if cc, _, err := container.Find(query); err == nil && cc != container {
				container = cc
			} else {
				groupName := strings.Join(query, " ")
				groupPath := append(cmd.Path[0:idx], query...) // nolint:gocritic
				cc := &cobra.Command{
					Use:                        cp,
					Short:                      fmt.Sprintf("%s subcommands", groupName),
					DisableAutoGenTag:          true,
					SuggestionsMinimumDistance: 2,
					SilenceUsage:               true,
					SilenceErrors:              true,
					Annotations: map[string]string{
						_c.ContextKeyRuntimeIndex: strings.Join(groupPath, " "),
					},
					Args: func(cmd *cobra.Command, args []string) error {
						if err := cobra.OnlyValidArgs(cmd, args); err == nil {
							return nil
						}

						suggestions := []string{}
						bold := color.New(color.Bold)
						for _, l := range cmd.SuggestionsFor(args[len(args)-1]) {
							suggestions = append(suggestions, bold.Sprint(l))
						}
						last := len(args) - 1
						parent := cmd.CommandPath()
						errMessage := fmt.Sprintf("Unknown subcommand %s of known command %s", bold.Sprint(args[last]), bold.Sprint(parent))
						if len(suggestions) > 0 {
							errMessage += ". Perhaps you meant " + strings.Join(suggestions, ", ") + "?"
						}
						return errors.NotFound{Msg: errMessage, Group: []string{}}
					},
					ValidArgs: []string{""},
					RunE: func(cc *cobra.Command, args []string) error {
						if len(args) == 0 {
							return errors.NotFound{Msg: "No subcommand provided", Group: []string{}}
						}
						os.Exit(_c.ExitStatusNotFound)
						return nil
					},
				}

				groupParent := &command.Command{
					Path:        cmd.Path[0 : len(cmd.Path)-1],
					Summary:     fmt.Sprintf("%s subcommands", groupName),
					Description: fmt.Sprintf("Runs subcommands within %s", groupName),
					Arguments:   command.Arguments{},
					Options:     command.Options{},
				}
				Register(groupParent)
				cc.SetHelpFunc(groupParent.HelpRenderer(command.Options{}))
				container.AddCommand(cc)
				container = cc
			}
		}
	}
	cmdRoot.SetCobra(ccRoot)

	return ccRoot.Execute()
}

var ccRoot = &cobra.Command{
	Use: "joao [--silent|-v|--verbose] [--[no-]color] [-h|--help] [--version]",
	Annotations: map[string]string{
		_c.ContextKeyRuntimeIndex: "joao",
	},
	DisableAutoGenTag: true,
	SilenceUsage:      true,
	SilenceErrors:     true,
	ValidArgs:         []string{""},
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
			return errors.NotFound{Msg: errMessage, Group: []string{}}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// if ok, err := cmd.Flags().GetBool("version"); err == nil && ok {
			// 	vc, _, err := cmd.Root().Find([]string{versionName()})

			// 	if err != nil {
			// 		return err
			// 	}
			// 	return vc.RunE(vc, []string{})
			// }
			return errors.NotFound{Msg: "No subcommand provided", Group: []string{}}
		}

		return nil
	},
}
