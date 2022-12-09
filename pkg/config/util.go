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
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	opClient "git.rob.mx/nidito/joao/internal/op-client"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func argIsYAMLFile(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

// func pathToName(path string, prefix string) string {
// 	comps := strings.SplitN(path, prefix+"/", 2)
// 	return strings.ReplaceAll(strings.Replace(comps[1], ".yaml", "", 1), "/", ":")
// }

// func nameToPath(name string) string {
// 	return "config/" + strings.ReplaceAll(name, ":", "/") + ".yaml"
// }

func vaultAndNameFrom(path string, buf []byte) (string, string, error) {
	smc := &singleModeConfig{}
	if buf == nil {
		var err error
		buf, err = ioutil.ReadFile(path)
		if err != nil {
			return "", "", fmt.Errorf("could not read file %s", path)
		}
	}

	if err := yaml.Unmarshal(buf, &smc); err == nil && smc.Config != nil {
		return smc.Config.Vault, smc.Config.Name, nil
	}

	rmc, err := findRepoConfig(path)
	if err != nil {
		return "", "", err
	}

	if rmc == nil {
		return "", "", fmt.Errorf("could not find repo config for %s", path)
	}

	if rmc.NameTemplate == "" {
		rmc.NameTemplate = "{{ DirName }}:{{ FileName}}"
	}

	logrus.Debugf("Found repo config at %s", rmc.Repo)

	tpl := template.Must(template.New("help").Funcs(template.FuncMap{
		"DirName": func() string {
			return filepath.Base(filepath.Dir(path))
		},
		"FileName": func() string {
			return strings.Split(filepath.Base(path), ".")[0]
		},
	}).Parse(rmc.NameTemplate))

	var nameBuf bytes.Buffer
	err = tpl.Execute(&nameBuf, nil)
	if err != nil {
		return "", "", err
	}
	return nameBuf.String(), rmc.Vault, nil
}

func Load(ref string, preferRemote bool) (*Config, error) {
	isYaml := argIsYAMLFile(ref)
	if preferRemote {
		name := ref
		vault := ""

		if isYaml {
			var err error
			name, vault, err = vaultAndNameFrom(ref, nil)
			if err != nil {
				return nil, err
			}
		} else {
			parts := strings.SplitN(ref, "/", 2)
			if len(parts) > 1 {
				vault = parts[0]
				name = parts[1]
			}
		}

		item, err := opClient.Get(vault, name)
		if err != nil {
			return nil, err
		}

		return ConfigFromOP(item)
	}

	if !isYaml {
		return nil, fmt.Errorf("could not load %s from local as it's not a path", ref)
	}

	return ConfigFromFile(ref)
}
