// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type opDetails struct {
	Vault        string `yaml:"vault"`
	Name         string `yaml:"name"`
	NameTemplate string `yaml:"nameTemplate"` // nolint: tagliatelle
	Repo         string
}

type singleModeConfig struct {
	Config *opDetails `yaml:"_config,omitempty"` // nolint: tagliatelle
}

func argIsYAMLFile(path string) bool {
	return strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
}

// VaultAndNameFrom path/buffer reads a path (unless a buffer is provided) and gets the 1Password
// item name and vault name:
// first, it looks at the embedded `_config: !!joao` YAML item.
// if it still needs a vault or name, it looks for the repo config, erroring if none found
// otherwise, it'll fill in missing values from the found repo config
func VaultAndNameFrom(path string, buf []byte) (name string, vault string, err error) {
	smc := &singleModeConfig{}

	// if a buffer was not provided, read from filesystem
	if buf == nil {
		var err error
		buf, err = os.ReadFile(path)
		if err != nil {
			return "", "", fmt.Errorf("could not read file %s", path)
		}
	}

	// decode single-mode config
	if err = yaml.Unmarshal(buf, &smc); err == nil && smc.Config != nil {
		name = smc.Config.Name
		vault = smc.Config.Vault
	}

	// if we have both name and vault, return early
	if name != "" && vault != "" {
		return name, vault, nil
	}

	// look for whole-repo config
	rmc, err := findRepoConfig(path)
	if err != nil {
		return name, vault, err
	}

	if rmc == nil {
		// no repo config found
		return name, vault, fmt.Errorf("could not find repo config for %s", path)
	}
	logrus.Debugf("Found repo config at %s", rmc.Repo)

	if name == "" {
		if rmc.NameTemplate == "" {
			rmc.NameTemplate = "{{ DirName }}:{{ FileName}}"
		}
		logrus.Tracef("Generating name for path %s from template %s", path, rmc.NameTemplate)

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
			return "", "", fmt.Errorf("could not generate item name for %s using template %s: %s", path, rmc.NameTemplate, err)
		}
		name = nameBuf.String()
		logrus.Tracef("Setting name for path %s from repo config %s", path, name)
	}

	if rmc.Vault != "" && vault == "" {
		logrus.Tracef("Setting vault for path %s from repo config %s", path, rmc.Vault)
		vault = rmc.Vault
	}

	return name, vault, nil
}

func isNumeric(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}
