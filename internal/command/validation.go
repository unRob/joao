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

	"github.com/go-playground/validator/v10"
)

type varSearchMap struct {
	Status int
	Name   string
	Usage  string
}

func (cmd *Command) Validate() (report map[string]int) {
	report = map[string]int{}

	validate := validator.New()
	if err := validate.Struct(cmd); err != nil {
		verrs := err.(validator.ValidationErrors)
		for _, issue := range verrs {
			// todo: output better errors, see validator.FieldError
			report[fmt.Sprint(issue)] = 1
		}
	}

	vars := map[string]map[string]*varSearchMap{
		"argument": {},
		"option":   {},
	}

	for _, arg := range cmd.Arguments {
		vars["argument"][strings.ToUpper(strings.ReplaceAll(arg.Name, "-", "_"))] = &varSearchMap{2, arg.Name, ""}
	}

	for name := range cmd.Options {
		vars["option"][strings.ToUpper(strings.ReplaceAll(name, "-", "_"))] = &varSearchMap{2, name, ""}
	}

	return report
}
