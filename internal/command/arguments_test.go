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
package command_test

import (
	"reflect"
	"strings"
	"testing"

	. "git.rob.mx/nidito/joao/internal/command"
)

func testCommand() *Command {
	return (&Command{
		Arguments: []*Argument{
			{
				Name:    "first",
				Default: "default",
			},
			{
				Name:     "variadic",
				Default:  []any{"defaultVariadic0", "defaultVariadic1"},
				Variadic: true,
			},
		},
		Options: Options{
			"option": {
				Default: "default",
				Type:    "string",
			},
			"bool": {
				Type:    "bool",
				Default: false,
			},
		},
	}).SetBindings()
}

func TestParse(t *testing.T) {
	cmd := testCommand()
	cmd.Arguments.Parse([]string{"asdf", "one", "two", "three"})
	known := cmd.Arguments.AllKnown()

	if !cmd.Arguments[0].IsKnown() {
		t.Fatalf("first argument isn't known")
	}
	val, exists := known["first"]
	if !exists {
		t.Fatalf("first argument isn't on AllKnown map: %v", known)
	}

	if val != "asdf" {
		t.Fatalf("first argument does not match. expected: %s, got %s", "asdf", val)
	}

	if !cmd.Arguments[1].IsKnown() {
		t.Fatalf("variadic argument isn't known")
	}
	val, exists = known["variadic"]
	if !exists {
		t.Fatalf("variadic argument isn't on AllKnown map: %v", known)
	}

	if !reflect.DeepEqual(val, []string{"one", "two", "three"}) {
		t.Fatalf("Known argument does not match. expected: %s, got %s", "one two three", val)
	}

	cmd = testCommand()
	cmd.Arguments.Parse([]string{"asdf"})
	known = cmd.Arguments.AllKnown()

	if !cmd.Arguments[0].IsKnown() {
		t.Fatalf("first argument is not known")
	}

	val, exists = known["first"]
	if !exists {
		t.Fatalf("first argument isn't on AllKnown map: %v", known)
	}

	if val != "asdf" {
		t.Fatalf("first argument does not match. expected: %s, got %s", "asdf", val)
	}

	val, exists = known["variadic"]
	if !exists {
		t.Fatalf("variadic argument isn't on AllKnown map: %v", known)
	}

	expected := []string{"defaultVariadic0", "defaultVariadic1"}
	if !reflect.DeepEqual(val, expected) {
		t.Fatalf("variadic argument does not match. expected: %s, got %s", expected, val)
	}
}

func TestBeforeParse(t *testing.T) {
	cmd := testCommand()
	known := cmd.Arguments.AllKnown()

	if cmd.Arguments[0].IsKnown() {
		t.Fatalf("first argument is known")
	}

	val, exists := known["first"]
	if !exists {
		t.Fatalf("first argument isn't on AllKnown map: %v", known)
	}

	if val != "default" {
		t.Fatalf("first argument does not match. expected: %s, got %s", "asdf", val)
	}

	val, exists = known["variadic"]
	if !exists {
		t.Fatalf("variadic argument isn't on AllKnown map: %v", known)
	}

	expected := []string{"defaultVariadic0", "defaultVariadic1"}
	if !reflect.DeepEqual(val, expected) {
		t.Fatalf("variadic argument does not match. expected: %s, got %s", expected, val)
	}
}

func TestArgumentsValidate(t *testing.T) {
	staticArgument := func(name string, def string, values []string, variadic bool) *Argument {
		return &Argument{
			Name:     name,
			Default:  def,
			Variadic: variadic,
			Required: def == "",
			Values: &ValueSource{
				Static: &values,
			},
		}
	}

	cases := []struct {
		Command     *Command
		Args        []string
		ErrorSuffix string
		Env         []string
	}{
		{
			Command: (&Command{
				// Name: []string{"test", "required", "failure"},
				Arguments: []*Argument{
					{
						Name:     "first",
						Required: true,
					},
				},
			}).SetBindings(),
			ErrorSuffix: "Missing argument for FIRST",
		},
		{
			Args:        []string{"bad"},
			ErrorSuffix: "bad is not a valid value for argument <first>. Valid options are: good, default",
			Command: (&Command{
				// Name: []string{"test", "script", "bad"},
				Arguments: []*Argument{
					{
						Name:    "first",
						Default: "default",
						Values: &ValueSource{
							Script: "echo good; echo default",
						},
					},
				},
			}).SetBindings(),
		},
		{
			Args:        []string{"bad"},
			ErrorSuffix: "bad is not a valid value for argument <first>. Valid options are: default, good",
			Command: (&Command{
				// Name: []string{"test", "static", "errors"},
				Arguments: []*Argument{staticArgument("first", "default", []string{"default", "good"}, false)},
			}).SetBindings(),
		},
		{
			Args:        []string{"default", "good", "bad"},
			ErrorSuffix: "bad is not a valid value for argument <first>. Valid options are: default, good",
			Command: (&Command{
				// Name:      []string{"test", "static", "errors"},
				Arguments: []*Argument{staticArgument("first", "default", []string{"default", "good"}, true)},
			}).SetBindings(),
		},
		{
			Args:        []string{"good"},
			ErrorSuffix: "could not validate argument for command test script bad-exit, ran",
			Command: (&Command{
				Path: []string{"test", "script", "bad-exit"},
				Arguments: []*Argument{
					{
						Name:    "first",
						Default: "default",
						Values: &ValueSource{
							Script: "echo good; echo default; exit 2",
						},
					},
				},
			}).SetBindings(),
		},
	}

	t.Run("good command is good", func(t *testing.T) {
		cmd := testCommand()
		cmd.Arguments[0] = staticArgument("first", "default", []string{"default", "good"}, false)
		cmd.Arguments[1] = staticArgument("second", "", []string{"one", "two", "three"}, true)
		cmd.SetBindings()

		cmd.Arguments.Parse([]string{"first", "one", "three", "two"})

		err := cmd.Arguments.AreValid()
		if err == nil {
			t.Fatalf("Unexpected failure validating: %s", err)
		}
	})

	for _, c := range cases {
		t.Run(c.Command.FullName(), func(t *testing.T) {
			c.Command.Arguments.Parse(c.Args)

			err := c.Command.Arguments.AreValid()
			if err == nil {
				t.Fatalf("Expected failure but got none")
			}
			if !strings.HasPrefix(err.Error(), c.ErrorSuffix) {
				t.Fatalf("Could not find error <%s> got <%s>", c.ErrorSuffix, err)
			}
		})
	}
}

// func TestArgumentsToEnv(t *testing.T) {
// 	cases := []struct {
// 		Command *Command
// 		Args    []string
// 		Expect  []string
// 		Env     []string
// 	}{
// 		{
// 			Args:   []string{"something"},
// 			Expect: []string{"export MILPA_ARG_FIRST=something"},
// 			Command: &Command{
// 				// Name: []string{"test", "required", "present"},
// 				Arguments: []*Argument{
// 					{
// 						Name:     "first",
// 						Required: true,
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Args:   []string{},
// 			Expect: []string{"export MILPA_ARG_FIRST=default"},
// 			Command: &Command{
// 				// Name: []string{"test", "default", "present"},
// 				Arguments: []*Argument{
// 					{
// 						Name:    "first",
// 						Default: "default",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Args: []string{"zero", "one", "two", "three"},
// 			Expect: []string{
// 				"export MILPA_ARG_FIRST=zero",
// 				"declare -a MILPA_ARG_VARIADIC='( one two three )'",
// 			},
// 			Command: &Command{
// 				// Name: []string{"test", "variadic"},
// 				Arguments: []*Argument{
// 					{
// 						Name:    "first",
// 						Default: "default",
// 					},
// 					{
// 						Name:     "variadic",
// 						Variadic: true,
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Args:   []string{},
// 			Expect: []string{"export MILPA_ARG_FIRST=default"},
// 			Command: &Command{
// 				// Name: []string{"test", "static", "default"},
// 				Arguments: []*Argument{
// 					{
// 						Name:    "first",
// 						Default: "default",
// 						Values: &ValueSource{
// 							Static: &[]string{
// 								"default",
// 								"good",
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Args:   []string{"good"},
// 			Expect: []string{"export MILPA_ARG_FIRST=good"},
// 			Command: &Command{
// 				// Name: []string{"test", "static", "good"},
// 				Arguments: []*Argument{
// 					{
// 						Name:    "first",
// 						Default: "default",
// 						Values: &ValueSource{
// 							Static: &[]string{
// 								"default",
// 								"good",
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Args:   []string{"good"},
// 			Expect: []string{"export MILPA_ARG_FIRST=good"},
// 			Command: &Command{
// 				// Name: []string{"test", "script", "good"},
// 				Arguments: []*Argument{
// 					{
// 						Name:    "first",
// 						Default: "default",
// 						Values: &ValueSource{
// 							Script: "echo good; echo default",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	for _, c := range cases {
// 		t.Run(c.Command.FullName(), func(t *testing.T) {
// 			dst := []string{}
// 			c.Command.SetBindings()
// 			c.Command.Arguments.Parse(c.Args)
// 			c.Command.Arguments.ToEnv(c.Command, &dst, "export ")

// 			err := c.Command.Arguments.AreValid()
// 			if err != nil {
// 				t.Fatalf("Unexpected failure validating: %s", err)
// 			}

// 			for _, expected := range c.Expect {
// 				found := false
// 				for _, actual := range dst {
// 					if strings.HasPrefix(actual, expected) {
// 						found = true
// 						break
// 					}
// 				}

// 				if !found {
// 					t.Fatalf("Expected line %v not found in %v", expected, dst)
// 				}
// 			}
// 		})
// 	}
// }

func TestArgumentToDesc(t *testing.T) {
	cases := []struct {
		Arg  *Argument
		Spec string
	}{
		{
			Arg: &Argument{
				Name: "regular",
			},
			Spec: "[REGULAR]",
		},
		{
			Arg: &Argument{
				Name:     "required",
				Required: true,
			},
			Spec: "REQUIRED",
		},
		{
			Arg: &Argument{
				Name:     "variadic-regular",
				Variadic: true,
			},
			Spec: "[VARIADIC_REGULAR...]",
		},
		{
			Arg: &Argument{
				Name:     "variadic-required",
				Variadic: true,
				Required: true,
			},
			Spec: "VARIADIC_REQUIRED...",
		},
	}

	for _, c := range cases {
		t.Run(c.Arg.Name, func(t *testing.T) {
			res := c.Arg.ToDesc()
			if res != c.Spec {
				t.Fatalf("Expected %s got %s", c.Spec, res)
			}
		})
	}
}
