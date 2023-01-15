// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/blake2b"
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

func VaultAndNameFrom(path string, buf []byte) (name string, vault string, err error) {
	smc := &singleModeConfig{}
	if buf == nil {
		var err error
		buf, err = ioutil.ReadFile(path)
		if err != nil {
			return "", "", fmt.Errorf("could not read file %s", path)
		}
	}

	if err = yaml.Unmarshal(buf, &smc); err == nil && smc.Config != nil {
		return smc.Config.Name, smc.Config.Vault, nil
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

func checksum(fields []*op.ItemField) string {
	newHash, err := blake2b.New256(nil)
	if err != nil {
		panic(err)
	}
	df := []string{}
	for _, field := range fields {
		if field.ID == "password" || field.ID == "notesPlain" || (field.Section != nil && field.Section.ID == "~annotations") {
			continue
		}
		label := field.Label
		if field.Section != nil && field.Section.ID != "" {
			label = field.Section.ID + "." + label
		}
		df = append(df, label+field.Value)
	}
	sort.Strings(df)
	newHash.Write([]byte(strings.Join(df, "")))
	checksum := newHash.Sum(nil)

	return fmt.Sprintf("%x", checksum)
}

func isNumeric(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}
