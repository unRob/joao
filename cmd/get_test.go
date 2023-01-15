// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd_test

import (
	"bytes"
	"encoding/json"
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
		{ID: "nested", Label: "nested"},
		{ID: "list", Label: "list"},
	},
	Fields: []*onepassword.ItemField{
		{
			ID:      "password",
			Type:    "CONCEALED",
			Purpose: "PASSWORD",
			Label:   "password",
			Value:   "8b23de7705b79b73d9f75b120651bc162859e45a732b764362feaefc882eab5d",
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
			ID:      "~annotations.nested.list.0",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "nested.list.0",
			Value:   "int",
		},
		{
			ID:      "nested.list.0",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "STRING",
			Label:   "list.0",
			Value:   "1",
		},
		{
			ID:      "~annotations.nested.list.1",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "nested.list.1",
			Value:   "int",
		},
		{
			ID:      "nested.list.1",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "STRING",
			Label:   "list.1",
			Value:   "2",
		},
		{
			ID:      "~annotations.nested.list.2",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "nested.list.2",
			Value:   "int",
		},
		{
			ID:      "nested.list.2",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "STRING",
			Label:   "list.2",
			Value:   "3",
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
			ID:      "~annotations.nested.second_secret",
			Section: &onepassword.ItemSection{ID: "~annotations", Label: "~annotations"},
			Type:    "STRING",
			Label:   "nested.second_secret",
			Value:   "secret",
		},
		{
			ID:      "nested.second_secret",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "CONCEALED",
			Label:   "second_secret",
			Value:   "very secret",
		},
		{
			ID:      "nested.string",
			Section: &onepassword.ItemSection{ID: "nested", Label: "nested"},
			Type:    "STRING",
			Label:   "string",
			Value:   "quem",
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
second_secret: very secret
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
second_secret: very secret
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
  second_secret: !!secret very secret
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

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"list":[1,2,3],"second_secret":"very secret","secret":"very secret","string":"quem"},"secret":"very secret","string":"pato"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetDeepJSON(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/deeply-nested.test.yaml", ".", "--output", "json"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"root":{"of":{"deeply":{"nested_list":["10.42.31.42/32"],"nested_map":"asdf"}}}}`

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

	expected := `{"bool":true,"int":1,"list":[1,2,3],"second_secret":"very secret","secret":"very secret","string":"quem"}`

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

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"list":[1,2,3],"second_secret":"","secret":"","string":"quem"},"secret":"","string":"pato"}`

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

	id := testConfig.ID
	testConfig.ID = ""
	defer func() { testConfig.ID = id }()
	expected, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	if got := out.String(); strings.TrimSpace(got) != strings.TrimSpace(string(expected)) {
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
  list:
    - 1
    - 2
    - 3
  second_secret: !!secret very secret
  secret: !!secret very secret
  string: quem
secret: !!secret very secret
string: pato`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}
