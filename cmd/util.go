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
	"io/ioutil"
	"strings"

	opClient "git.rob.mx/nidito/joao/internal/op-client"
	"git.rob.mx/nidito/joao/pkg/config"
)

func argIsYAMLFile(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

func pathToName(path string) string {
	comps := strings.Split(path, "config/")
	return strings.ReplaceAll(strings.Replace(comps[len(comps)-1], ".yaml", "", 1), "/", ":")
}

func nameToPath(name string) string {
	return "config/" + strings.ReplaceAll(name, ":", "/") + ".yaml"
}

func loadExisting(ref string, preferRemote bool) (*config.Config, error) {
	isYaml := argIsYAMLFile(ref)
	if preferRemote {
		name := ref
		if isYaml {
			name = pathToName(ref)
		}

		item, err := opClient.Get("nidito-admin", name)
		if err != nil {
			return nil, err
		}

		return config.ConfigFromOP(item)
	}

	path := ref
	var name string
	if !isYaml {
		path = nameToPath(ref)
		name = ref
	} else {
		name = pathToName(ref)
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s", ref)
	}

	if len(buf) == 0 {
		buf = []byte("{}")
	}

	return config.ConfigFromYAML(buf, name)
}
