// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	. "git.rob.mx/nidito/joao/cmd"
	"git.rob.mx/nidito/joao/internal/testdata"
	"git.rob.mx/nidito/joao/internal/testdata/opconnect"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func TestGetBadYAML(t *testing.T) {
	Get.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{testdata.YAML("bad-test"), "."})
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
	root := testdata.FromProjectRoot()
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
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	err := Get.Run(cmd, []string{testdata.YAML("test"), "."})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected, err := os.ReadFile(testdata.YAML("test"))
	if err != nil {
		t.Fatalf("could not read file: %s", err)
	}

	if got := out.String(); strings.TrimSpace(got) != strings.TrimSpace(string(expected)) {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetRedacted(t *testing.T) {
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", true, "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	os.Args = []string{testdata.YAML("test"), ".", "--redacted"}
	err := Get.Run(cmd, []string{testdata.YAML("test"), "."})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected, err := os.ReadFile(testdata.YAML("test"))
	if err != nil {
		t.Fatalf("could not read file: %s", err)
	}

	if got := out.String(); strings.TrimSpace(got) != strings.ReplaceAll(strings.TrimSpace(string(expected)), " very secret", "") {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetPath(t *testing.T) {
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "yaml", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	os.Args = []string{testdata.YAML("test"), "nested.secret"}
	err := Get.Run(cmd, os.Args)

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
	os.Args = []string{testdata.YAML("test"), "nested", "--output", "diff-yaml"}
	err = Get.Run(cmd, []string{testdata.YAML("test"), "nested"})

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
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "yaml", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	os.Args = []string{testdata.YAML("test"), "nested", "--output", "yaml"}
	err := Get.Run(cmd, []string{testdata.YAML("test"), "nested"})

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
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "diff-yaml", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	os.Args = []string{testdata.YAML("test"), ".", "--output", "diff-yaml"}
	err := Get.Run(cmd, []string{testdata.YAML("test"), "."})

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
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	os.Args = []string{testdata.YAML("test"), ".", "--output", "json"}
	err := Get.Run(cmd, []string{testdata.YAML("test"), "."})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"list":[1,2,3],"second_secret":"very secret","secret":"very secret","string":"quem"},"secret":"very secret","string":"pato"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetDeepJSON(t *testing.T) {
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	file := testdata.YAML("deeply-nested.test")
	os.Args = []string{file, ".", "--output", "json"}
	err := Get.Run(cmd, []string{file, "."})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"root":{"of":{"deeply":{"nested_list":["10.42.31.42/32"],"nested_map":"asdf"}}}}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONPathScalar(t *testing.T) {
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	file := testdata.YAML("test")
	os.Args = []string{file, "nested.secret", "--output", "json"}
	err := Get.Run(cmd, []string{file, "nested.secret"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `very secret` // nolint: ifshort

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONPathCollection(t *testing.T) {
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", false, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	file := testdata.YAML("test")
	os.Args = []string{file, "nested", "--output", "json"}
	err := Get.Run(cmd, []string{file, "nested"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"bool":true,"int":1,"list":[1,2,3],"second_secret":"very secret","secret":"very secret","string":"quem"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONRedacted(t *testing.T) {
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("redacted", true, "")
	cmd.Flags().StringP("output", "o", "json", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	file := testdata.YAML("test")
	os.Args = []string{file, ".", "--output", "json", "--redacted"}
	err := Get.Run(cmd, []string{file, "."})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `{"bool":false,"int":1,"list":["one","two","three"],"nested":{"bool":true,"int":1,"list":[1,2,3],"second_secret":"","secret":"","string":"quem"},"secret":"","string":"pato"}`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetJSONOP(t *testing.T) {
	out := bytes.Buffer{}
	Get.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().StringP("output", "o", "op", "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Get.Cobra = cmd
	os.Args = []string{testdata.YAML("test"), ".", "--output", "op"}
	err := Get.Run(cmd, []string{testdata.YAML("test"), "."})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	cfg := testdata.NewTestConfig("some:test")
	id := cfg.ID
	cfg.ID = ""
	defer func() { cfg.ID = id }()
	expected, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if got := out.String(); strings.TrimSpace(got) != strings.TrimSpace(string(expected)) {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}

func TestGetRemote(t *testing.T) {
	testdata.MockOPConnect(t)
	opconnect.Add(testdata.NewTestConfig("some:test"))
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
	os.Args = []string{testdata.YAML("test"), "."}
	err := Get.Run(cmd, []string{testdata.YAML("test"), "."})

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
