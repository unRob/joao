// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd_test

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	. "git.rob.mx/nidito/joao/cmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func tempTestYaml(root, name string, data []byte) (string, func(), error) {
	path := fmt.Sprintf("%s/test-%s.yaml", root, name)
	if err := ioutil.WriteFile(path, data, fs.FileMode(0644)); err != nil {
		return path, nil, fmt.Errorf("could not create test file")
	}
	return path, func() { os.Remove(path) }, nil
}

func TestSet(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	stdin := bytes.Buffer{}
	stdin.Write([]byte("pato\nganso\nmarreco\n"))
	cmd.SetIn(&stdin)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", false, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-plain", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = Set.Run(cmd, []string{path, "string"})
	if err != nil {
		t.Fatalf("Threw on good set: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if !strings.Contains(string(changed), `
string: |-
  pato
  ganso
  marreco`) {
		t.Fatalf("Did not contain expected new string, got:\n%s", changed)
	}
}

func TestSetSecret(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	stdin := bytes.Buffer{}
	stdin.Write([]byte("new secret\n"))
	cmd.SetIn(&stdin)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", true, "")
	cmd.Flags().Bool("delete", false, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-plain", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	os.Args = []string{path, "secret", "--secret"}
	err = Set.Run(cmd, []string{path, "secret"})
	if err != nil {
		t.Fatalf("Threw on good set: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if !strings.Contains(string(changed), "\nsecret: !!secret new secret\n") {
		t.Fatalf("Did not contain expected new string, got:\n%s", changed)
	}
}

func TestSetFromFile(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", false, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}
	dataPath, dataCleanup, err := tempTestYaml(root, "set-from-file-data", []byte("ganso"))
	if err != nil {
		t.Fatal(err)
	}
	defer dataCleanup()
	cmd.Flags().StringP("input", "i", dataPath, "")

	path, cleanup, err := tempTestYaml(root, "set-from-file", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	os.Args = []string{path, "string", "--input", dataPath}
	err = Set.Run(cmd, []string{path, "string"})
	if err != nil {
		t.Fatalf("Threw on good set: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if !strings.Contains(string(changed), "\nstring: ganso\n") {
		t.Fatalf("Did not contain expected new string, got:\n%s", changed)
	}
}

func TestSetNew(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	stdin := bytes.Buffer{}
	stdin.Write([]byte("pato\nganso\nmarreco\ncisne\n"))
	cmd.SetIn(&stdin)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", false, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	cmd.Flags().StringP("input", "i", "/dev/stdin", "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-new-key", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = Set.Run(cmd, []string{path, "quarteto"})
	if err != nil {
		t.Fatalf("Threw on good new set: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if !strings.Contains(string(changed), `
quarteto: |-
  pato
  ganso
  marreco
  cisne`) {
		t.Fatalf("Did not contain expected new string, got:\n%s", changed)
	}
}

func TestSetNested(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	stdin := bytes.Buffer{}
	stdin.Write([]byte("tico"))
	cmd.SetIn(&stdin)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", false, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	cmd.Flags().StringP("input", "i", "/dev/stdin", "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-nested-key", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = Set.Run(cmd, []string{path, "nested.tico"})
	if err != nil {
		t.Fatalf("Threw on good nested set: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if !strings.Contains(string(changed), `
  tico: tico
`) {
		t.Fatalf("Did not contain expected new string, got:\n%s", changed)
	}
}

func TestSetJSON(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	stdin := bytes.Buffer{}
	stdin.Write([]byte(`{"foram": "ensaiar", "para": "começar"}`))
	cmd.SetIn(&stdin)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", false, "")
	cmd.Flags().Bool("json", true, "")
	cmd.Flags().Bool("flush", false, "")
	cmd.Flags().StringP("input", "i", "/dev/stdin", "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-json", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = Set.Run(cmd, []string{path, "na-beira-da-lagoa"})
	if err != nil {
		t.Fatalf("Threw on good nested set: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if !strings.Contains(string(changed), `
na-beira-da-lagoa:
  foram: ensaiar
  para: começar
`) {
		t.Fatalf("Did not contain expected new entry tree, got:\n%s", changed)
	}
}

func TestSetList(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	stdin := bytes.Buffer{}
	stdin.Write([]byte("um"))
	cmd.SetIn(&stdin)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", false, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	cmd.Flags().StringP("input", "i", "/dev/stdin", "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-list-key", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	err = Set.Run(cmd, []string{path, "asdf.0"})
	if err != nil {
		t.Fatalf("Threw on good nested set: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if !strings.Contains(string(changed), `
asdf:
  - um
`) {
		t.Fatalf("Did not contain expected new string, got:\n%s", changed)
	}
}

func TestDelete(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", true, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	cmd.Flags().StringP("input", "i", "/dev/stdin", "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-delete-key", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	os.Args = []string{path, "string", "--delete"}
	err = Set.Run(cmd, []string{path, "string"})
	if err != nil {
		t.Fatalf("Threw on good set delete: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if strings.Contains(string(changed), `
string: pato
`) {
		t.Fatalf("Still contains deleted key, got:\n%s", changed)
	}
}

func TestDeleteNested(t *testing.T) {
	root := fromProjectRoot()
	Set.SetBindings()
	out := bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Set.Cobra = cmd
	cmd.Flags().Bool("secret", false, "")
	cmd.Flags().Bool("delete", true, "")
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("flush", false, "")
	cmd.Flags().StringP("input", "i", "/dev/stdin", "")
	original, err := ioutil.ReadFile(root + "/test.yaml")
	if err != nil {
		t.Fatalf("could not read file")
	}

	path, cleanup, err := tempTestYaml(root, "set-delete-nested-key", original)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	os.Args = []string{path, "nested.string", "--delete"}
	err = Set.Run(cmd, []string{path, "nested.string"})
	if err != nil {
		t.Fatalf("Threw on good set delete nested: %s", err)
	}

	changed, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file")
	}

	if string(changed) == string(original) {
		t.Fatal("Did not change file")
	}

	if strings.Contains(string(changed), `
  string: quem
`) {
		t.Fatalf("Still contains deleted nested key, got:\n%s", changed)
	}
}
