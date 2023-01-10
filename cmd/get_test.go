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
	"github.com/spf13/cobra"
)

func fromProjectRoot() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../")
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	wd, _ := os.Getwd()
	return wd
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

	got := out.String()
	if strings.TrimSpace(got) != strings.ReplaceAll(strings.TrimSpace(string(expected)), " very secret", "") {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
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

	got := out.String()
	if strings.TrimSpace(got) != strings.TrimSpace(string(expected)) {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetPath(t *testing.T) {
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{root + "/test.yaml", "nested.secret"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := "very secret"
	got := out.String()
	if strings.TrimSpace(got) != strings.TrimSpace(expected) {
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
secret: very secret
string: quem`

	got := out.String()
	if strings.TrimSpace(got) != expected {
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
  secret: !!secret very secret
  string: quem
secret: !!secret very secret
string: "pato"`

	got := out.String()
	if strings.TrimSpace(got) != expected {
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

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"secret":"very secret","string":"quem"},"secret":"very secret","string":"pato"}`

	got := out.String()
	if strings.TrimSpace(got) != expected {
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

	expected := `very secret`

	got := out.String()
	if strings.TrimSpace(got) != expected {
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

	expected := `{"bool":true,"int":1,"secret":"very secret","string":"quem"}`

	got := out.String()
	if strings.TrimSpace(got) != expected {
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

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"secret":"","string":"quem"},"secret":"","string":"pato"}`

	got := out.String()
	if strings.TrimSpace(got) != expected {
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

	expected := `{"id":"","title":"some:test","vault":{"id":"example"},"category":"PASSWORD","sections":[{"id":"~annotations","label":"~annotations"},{"id":"nested","label":"nested"}],"fields":[{"id":"password","type":"CONCEALED","purpose":"PASSWORD","label":"password","value":"56615e9be5f0ce5f97d5b446faaa1d39f95a13a1ea8326ae933c3d29eb29735c"},{"id":"notesPlain","type":"STRING","purpose":"NOTES","label":"notesPlain","value":"flushed by joao"},{"id":"~annotations.int","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"int","value":"int"},{"id":"int","type":"STRING","label":"int","value":"1"},{"id":"string","type":"STRING","label":"string","value":"pato"},{"id":"~annotations.bool","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"bool","value":"bool"},{"id":"bool","type":"STRING","label":"bool","value":"false"},{"id":"~annotations.secret","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"secret","value":"secret"},{"id":"secret","type":"CONCEALED","label":"secret","value":"very secret"},{"id":"nested.string","section":{"id":"nested"},"type":"STRING","label":"string","value":"quem"},{"id":"~annotations.nested.int","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.int","value":"int"},{"id":"nested.int","section":{"id":"nested"},"type":"STRING","label":"int","value":"1"},{"id":"~annotations.nested.secret","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.secret","value":"secret"},{"id":"nested.secret","section":{"id":"nested"},"type":"CONCEALED","label":"secret","value":"very secret"},{"id":"~annotations.nested.bool","section":{"id":"~annotations","label":"~annotations"},"type":"STRING","label":"nested.bool","value":"bool"},{"id":"nested.bool","section":{"id":"nested"},"type":"STRING","label":"bool","value":"true"},{"id":"list.0","section":{"id":"list"},"type":"STRING","label":"0","value":"one"},{"id":"list.1","section":{"id":"list"},"type":"STRING","label":"1","value":"two"},{"id":"list.2","section":{"id":"list"},"type":"STRING","label":"2","value":"three"}],"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}`

	got := out.String()
	if strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}
