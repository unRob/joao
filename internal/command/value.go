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
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	_c "git.rob.mx/nidito/joao/internal/constants"
	"git.rob.mx/nidito/joao/internal/exec"
	"github.com/spf13/cobra"
)

// ValueType represent the kinds of or option.
type ValueType string

const (
	// ValueTypeDefault is the empty string, maps to ValueTypeString.
	ValueTypeDefault ValueType = ""
	// ValueTypeString a value treated like a string.
	ValueTypeString ValueType = "string"
	// ValueTypeBoolean is a value treated like a boolean.
	ValueTypeBoolean ValueType = "bool"
)

type SourceCommand struct {
	Path []string
	Args string
}

type CompletionFunc func(cmd *Command, currentValue string) (values []string, flag cobra.ShellCompDirective, err error)

// ValueSource represents the source for an auto-completed and/or validated option/argument.
type ValueSource struct {
	// Directories prompts for directories with the given prefix.
	Directories *string `json:"dirs,omitempty" yaml:"dirs,omitempty" validate:"omitempty,excluded_with=Command Files Func Script Static"`
	// Files prompts for files with the given extensions
	Files *[]string `json:"files,omitempty" yaml:"files,omitempty" validate:"omitempty,excluded_with=Command Func Directories Script Static"`
	// Script runs the provided command with `bash -c "$script"` and returns an option for every line of stdout.
	Script string `json:"script,omitempty" yaml:"script,omitempty" validate:"omitempty,excluded_with=Command Directories Files Func Static"`
	// Static returns the given list.
	Static *[]string `json:"static,omitempty" yaml:"static,omitempty" validate:"omitempty,excluded_with=Command Directories Files Func Script"`
	// Command runs a subcommand and returns an option for every line of stdout.
	Command *SourceCommand `json:"command,omitempty" yaml:"command,omitempty" validate:"omitempty,excluded_with=Directories Files Func Script Static"`
	// Func runs a function
	Func CompletionFunc `json:"func,omitempty" yaml:"func,omitempty" validate:"omitempty,excluded_with=Command Directories Files Script Static"`
	// Timeout is the maximum amount of time we will wait for a Script, Command, or Func before giving up on completions/validations.
	Timeout int `json:"timeout,omitempty" yaml:"timeout,omitempty" validate:"omitempty,excluded_with=Directories Files Static"`
	// Suggestion if provided will only suggest autocomplete values but will not perform validation of a given value
	Suggestion bool `json:"suggest-only" yaml:"suggest-only" validate:"omitempty"` // nolint:tagliatelle
	// SuggestRaw if provided the shell will not add a space after autocompleting
	SuggestRaw bool     `json:"suggest-raw" yaml:"suggest-raw" validate:"omitempty"` // nolint:tagliatelle
	command    *Command `json:"-" yaml:"-" validate:"-"`
	computed   *[]string
	flag       cobra.ShellCompDirective
}

// Validates tells if a value needs to be validated.
func (vs *ValueSource) Validates() bool {
	if vs.Directories != nil || vs.Files != nil {
		return false
	}

	return !vs.Suggestion
}

// Resolve returns the values for autocomplete and validation.
func (vs *ValueSource) Resolve(currentValue string) (values []string, flag cobra.ShellCompDirective, err error) {
	if vs.computed != nil {
		return *vs.computed, vs.flag, nil
	}

	if vs.Timeout == 0 {
		vs.Timeout = 5
	}

	flag = cobra.ShellCompDirectiveDefault
	timeout := time.Duration(vs.Timeout)

	switch {
	case vs.Static != nil:
		values = *vs.Static
	case vs.Files != nil:
		flag = cobra.ShellCompDirectiveFilterFileExt
		values = *vs.Files
	case vs.Directories != nil:
		flag = cobra.ShellCompDirectiveFilterDirs
		values = []string{*vs.Directories}
	case vs.Func != nil:
		ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
		defer cancel()

		done := make(chan error, 1)
		panicChan := make(chan any, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()

			values, flag, err = vs.Func(vs.command, currentValue)
			done <- err
		}()
		select {
		case err = <-done:
			return
		case p := <-panicChan:
			panic(p)
		case <-ctx.Done():
			flag = cobra.ShellCompDirectiveError
			err = ctx.Err()
			return
		}

	case vs.Command != nil:
		if vs.command == nil {
			return nil, cobra.ShellCompDirectiveError, fmt.Errorf("bug: command is nil")
		}
		argString, err := vs.command.ResolveTemplate(vs.Command.Args, currentValue)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError, err
		}
		args := strings.Split(argString, " ")
		sub, _, err := vs.command.Cobra.Root().Find(vs.Command.Path)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError, fmt.Errorf("could not find a command named %s", vs.Command.Path)
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
		defer cancel() // The cancel should be deferred so resources are cleaned up

		sub.SetArgs(args)
		var stdout bytes.Buffer
		sub.SetOut(&stdout)
		var stderr bytes.Buffer
		sub.SetErr(&stderr)
		err = sub.ExecuteContext(ctx)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError, err
		}

		values = strings.Split(stdout.String(), "\n")
		flag = cobra.ShellCompDirectiveDefault
		err = nil
	case vs.Script != "":
		if vs.command == nil {
			return nil, cobra.ShellCompDirectiveError, fmt.Errorf("bug: command is nil")
		}
		cmd, err := vs.command.ResolveTemplate(vs.Script, currentValue)
		if err != nil {
			return nil, cobra.ShellCompDirectiveError, err
		}

		args := append([]string{"/bin/bash", "-c"}, cmd)

		values, flag, err = exec.Exec(vs.command.FullName(), args, os.Environ(), timeout*time.Second)
		if err != nil {
			return nil, flag, err
		}
	}

	vs.computed = &values

	if vs.SuggestRaw {
		flag = flag | cobra.ShellCompDirectiveNoSpace
	}

	vs.flag = flag
	return values, flag, err
}

type AutocompleteTemplate struct {
	Args map[string]string
	Opts map[string]string
}

func (tpl *AutocompleteTemplate) Opt(name string) string {
	if val, ok := tpl.Opts[name]; ok {
		return fmt.Sprintf("--%s %s", name, val)
	}

	return ""
}

func (tpl *AutocompleteTemplate) Arg(name string) string {
	return tpl.Args[name]
}

func (cmd *Command) ResolveTemplate(templateString string, currentValue string) (string, error) {
	var buf bytes.Buffer

	tplData := &AutocompleteTemplate{
		Args: cmd.Arguments.AllKnownStr(),
		Opts: cmd.Options.AllKnownStr(),
	}

	fnMap := template.FuncMap{
		"Opt":     tplData.Opt,
		"Arg":     tplData.Arg,
		"Current": func() string { return currentValue },
	}

	for k, v := range _c.TemplateFuncs {
		fnMap[k] = v
	}

	tpl, err := template.New("subcommand").Funcs(fnMap).Parse(templateString)

	if err != nil {
		return "", err
	}

	err = tpl.Execute(&buf, tplData)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
