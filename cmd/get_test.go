// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd_test

import (
	"bytes"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	. "git.rob.mx/nidito/joao/cmd"
	opclient "git.rob.mx/nidito/joao/internal/op-client"
	"git.rob.mx/nidito/joao/internal/op-client/mock"
	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var testConfig = &onepassword.Item{
	Title:    "some:test",
	Vault:    onepassword.ItemVault{ID: "example"},
	Category: "PASSWORD",
	Sections: []*onepassword.ItemSection{
		{ID: "~annotations", Label: "~annotations"},
		// {ID: "nested", Label: "nested"},
		{ID: "list", Label: "list"},
	},
	Fields: []*onepassword.ItemField{
		{
			ID:      "password",
			Type:    "CONCEALED",
			Purpose: "PASSWORD",
			Label:   "password",
			Value:   "56615e9be5f0ce5f97d5b446faaa1d39f95a13a1ea8326ae933c3d29eb29735c",
		},
		{
			ID:      "notesPlain",
			Type:    "STRING",
			Purpose: "NOTES",
			Label:   "notesPlain",
			Value:   "flushed by joao",
		},
		{
			ID:      "~annotations.int",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "int",
			Value:   "int",
		},
		{
			ID:    "int",
			Type:  "STRING",
			Label: "int",
			Value: "1",
		},
		{
			ID:    "string",
			Type:  "STRING",
			Label: "string",
			Value: "pato",
		},
		{
			ID:      "~annotations.bool",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "bool",
			Value:   "bool",
		},
		{
			ID:    "bool",
			Type:  "STRING",
			Label: "bool",
			Value: "false",
		},
		{
			ID:      "~annotations.secret",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "secret",
			Value:   "secret",
		},
		{
			ID:    "secret",
			Type:  "CONCEALED",
			Label: "secret",
			Value: "very secret",
		},
		{
			ID:      "nested.string",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "STRING",
			Label:   "string",
			Value:   "quem",
		},
		{
			ID:      "~annotations.nested.int",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "nested.int",
			Value:   "int",
		},
		{
			ID:      "nested.int",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "STRING",
			Label:   "int",
			Value:   "1",
		},
		{
			ID:      "~annotations.nested.secret",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "nested.secret",
			Value:   "secret",
		},
		{
			ID:      "nested.secret",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "CONCEALED",
			Label:   "secret",
			Value:   "very secret",
		},
		{
			ID:      "~annotations.nested.bool",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "nested.bool",
			Value:   "bool",
		},
		{
			ID:      "nested.bool",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "STRING",
			Label:   "bool",
			Value:   "true",
		},
		{
			ID:      "list.0",
			Section: &onepassword.ItemSection{ID: "list", Label: "list"},
			Type:    "STRING",
			Label:   "0",
			Value:   "one",
		},
		{
			ID:      "list.1",
			Section: &onepassword.ItemSection{ID: "list", Label: "list"},
			Type:    "STRING",
			Label:   "1",
			Value:   "two",
		},
		{
			ID:      "list.2",
			Section: &onepassword.ItemSection{ID: "list", Label: "list"},
			Type:    "STRING",
			Label:   "2",
			Value:   "three",
		},
	},
}

func mockOPConnect(t *testing.T) {
	t.Helper()
	opclient.ConnectClientFactory = func(host, token, userAgent string) connect.Client {
		return &mock.Client{}
	}
	client := opclient.NewConnect("", "")
	opclient.Use(client)
	mock.Add(testConfig)
}

func fromProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	wd, _ := os.Getwd()
	return wd
}

func TestGetBadYAML(t *testing.T) {
	root := fromProjectRoot()
	Get.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/bad-test.yaml", "."})
	if err == nil {
		t.Fatalf("Did not throw on bad path: %s", out.String())
	}
	wantedPrefix := "could not parse file"
	wantedSuffix := "/bad-test.yaml as yaml: line 4: mapping values are not allowed in this context"
	if got := err.Error(); !(strings.HasPrefix(got, wantedPrefix) && strings.HasSuffix(got, wantedSuffix)) {
		t.Fatalf("Failed with bad error, wanted %s /some-path%s, got %s", wantedPrefix, wantedSuffix, got)
	}
}

func TestGetBadPath(t *testing.T) {
	root := fromProjectRoot()
	Get.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/does-not-exist.yaml", "."})
	if err == nil {
		t.Fatalf("Did not throw on bad path: %s", out.String())
	}
	wantedPrefix := "could not read file"
	wantedSuffix := "/does-not-exist.yaml"

	if got := err.Error(); !(strings.HasPrefix(got, wantedPrefix) && strings.HasSuffix(got, wantedSuffix)) {
		t.Fatalf("Failed with bad error, wanted %s /some-path%s, got %s", wantedPrefix, wantedSuffix, got)
	}
}

func TestGetNormal(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", "."})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected, err := os.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file: %s", err)
	}

	if got := out.String(); strings.TrimSpace(got) != strings.TrimSpace(string(expected)) {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetRedacted(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", true, "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", ".", "--redacted"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected, err := os.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file: %s", err)
	}

	if got := out.String(); strings.TrimSpace(got) != strings.ReplaceAll(strings.TrimSpace(string(expected)), " very secret", "") {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetPath(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "yaml", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", "nested.secret"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := "very secret"
	if got := out.String(); strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Fatalf("did not get expected scalar output:\nwanted: %s\ngot: %s", expected, got)
	}

	out = bytes.Buffer{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	err = Get.Run(cmd, []string{root + "/test.yaml", "nested", "--output", "diff-yaml"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected = `bool: true
int: 1
list:
  - 1
  - 2
  - 3
secret: very secret
string: quem`

	if got := out.String(); strings.TrimSpace(got) != strings.TrimSpace(expected) {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetPathCollection(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "yaml", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", "nested", "--output", "yaml"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `bool: true
int: 1
list:
  - 1
  - 2
  - 3
secret: very secret
string: quem`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetDiff(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "diff-yaml", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", ".", "--output", "diff-yaml"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `_config: !!joao
  name: some:test
  vault: example
bool: false
int: 1
list:
  - one
  - two
  - three
nested:
  bool: true
  int: 1
  list:
    - 1
    - 2
    - 3
  secret: !!secret very secret
  string: quem
secret: !!secret very secret
string: pato`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSON(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", ".", "--output", "json"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"list":[1,2,3],"secret":"very secret","string":"quem"},"secret":"very secret","string":"pato"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONPathScalar(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", "nested.secret", "--output", "json"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `very secret` // nolint: ifshort

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONPathCollection(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", "nested", "--output", "json"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"bool":true,"int":1,"list":[1,2,3],"secret":"very secret","string":"quem"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONRedacted(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", true, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", ".", "--output", "json", "--redacted"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"list":[1,2,3],"secret":"","string":"quem"},"secret":"","string":"pato"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONOP(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "op", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", ".", "--output", "op"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"id":"","title":"some:test","vault":{"id":"example"},"category":"PASSWORD","sections":[{"id":"~annotations","label":"~annotations"},{"id":"nested","label":"nested"},{"id":"list","label":"list"}],"fields":[{"id":"password","type":"CONCEALED","purpose":"PASSWORD","label":"password","value":"cedbdf86fb15cf1237569e9b3188372d623aea9d6a707401aca656645590227c"},{"id":"notesPlain","type":"STRING","purpose":"NOTES","label":"notesPlain","value":"flushed by joao"},{"id":"~annotations.int","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"int","value":"int"},{"id":"int","type":"STRING","label":"int","value":"1"},{"id":"string","type":"STRING","label":"string","value":"pato"},{"id":"~annotations.bool","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"bool","value":"bool"},{"id":"bool","type":"STRING","label":"bool","value":"false"},{"id":"~annotations.secret","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"secret","value":"secret"},{"id":"secret","type":"CONCEALED","label":"secret","value":"very secret"},{"id":"nested.string","section":{"id":"nested"},"type":"STRING","label":"string","value":"quem"},{"id":"~annotations.nested.int","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.int","value":"int"},{"id":"nested.int","section":{"id":"nested"},"type":"STRING","label":"int","value":"1"},{"id":"~annotations.nested.secret","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.secret","value":"secret"},{"id":"nested.secret","section":{"id":"nested"},"type":"CONCEALED","label":"secret","value":"very secret"},{"id":"~annotations.nested.bool","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.bool","value":"bool"},{"id":"nested.bool","section":{"id":"nested"},"type":"STRING","label":"bool","value":"true"},{"id":"~annotations.nested.list.0","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.list.0","value":"int"},{"id":"nested.list.0","section":{"id":"nested"},"type":"STRING","label":"list.0","value":"1"},{"id":"~annotations.nested.list.1","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.list.1","value":"int"},{"id":"nested.list.1","section":{"id":"nested"},"type":"STRING","label":"list.1","value":"2"},{"id":"~annotations.nested.list.2","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.list.2","value":"int"},{"id":"nested.list.2","section":{"id":"nested"},"type":"STRING","label":"list.2","value":"3"},{"id":"list.0","section":{"id":"list"},"type":"STRING","label":"0","value":"one"},{"id":"list.1","section":{"id":"list"},"type":"STRING","label":"1","value":"two"},{"id":"list.2","section":{"id":"list"},"type":"STRING","label":"2","value":"three"}],"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetRemote(t *testing.T) {
	mockOPConnect(t)
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().Bool("remote", true, "")
	cmd.Flags().StringP("output", "o", "diff-yaml", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	logrus.SetLevel(logrus.DebugLevel)
	err := Get.Run(cmd, []string{root + "/test.yaml", ".", "--output", "diff-yaml", "--remote"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `bool: false
int: 1
list:
  - one
  - two
  - three
nested:
  bool: true
  int: 1
  secret: !!secret very secret
  string: quem
secret: !!secret very secret
string: pato`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}
