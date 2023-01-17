// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	. "git.rob.mx/nidito/joao/cmd"
	"git.rob.mx/nidito/joao/internal/op-client/mock"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func TestFetch(t *testing.T) {
	mockOPConnect(t)
	f := testConfig.Fields
	s := testConfig.Sections
	defer func() { testConfig.Fields = f; testConfig.Sections = s }()

	testConfig.Sections = append(testConfig.Sections,
		&onepassword.ItemSection{ID: "o", Label: "o"},
		&onepassword.ItemSection{ID: "e-fez-tambem", Label: "e-fez-tambem"},
	)
	testConfig.Fields = append(testConfig.Fields,
		&onepassword.ItemField{
			ID:      "o.ganso.gosto",
			Section: &onepassword.ItemSection{ID: "o", Label: "o"},
			Type:    "STRING",
			Label:   "ganso.gosto",
			Value:   "da dupla",
		},
		&onepassword.ItemField{
			ID:      "e-fez-tambem.0",
			Section: &onepassword.ItemSection{ID: "e-fez-tambem", Label: "e-fez-tambem"},
			Type:    "STRING",
			Label:   "0",
			Value:   "quém!",
		},
		&onepassword.ItemField{
			ID:      "e-fez-tambem.1",
			Section: &onepassword.ItemSection{ID: "e-fez-tambem", Label: "e-fez-tambem"},
			Type:    "STRING",
			Label:   "1",
			Value:   "quém!",
		},
		&onepassword.ItemField{
			ID:      "e-fez-tambem.2",
			Section: &onepassword.ItemSection{ID: "e-fez-tambem", Label: "e-fez-tambem"},
			Type:    "STRING",
			Label:   "2",
			Value:   "quém!",
		})
	mock.Update(testConfig)
	root := fromProjectRoot()
	out := bytes.Buffer{}
	Fetch.SetBindings()
	cmd := &cobra.Command{}
	cmd.Flags().Bool("dry-run", true, "")
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	Fetch.Cobra = cmd
	logrus.SetLevel(logrus.DebugLevel)
	err := Fetch.Run(cmd, []string{root + "/test.yaml"})

	if err != nil {
		t.Fatalf("could not get: %s", err)
	}

	expected := `--- /Users/roberto/src/joao/test.yaml
+++ op://example/some:test
@@ -1,4 +1,8 @@
 bool: false
+e-fez-tambem:
+  - quém!
+  - quém!
+  - quém!
 int: 1
 list:
   - one
@@ -14,5 +18,8 @@
   second_secret: !!secret very secret
   secret: !!secret very secret
   string: quem
+o:
+  ganso:
+    gosto: da dupla
 secret: !!secret very secret
 string: pato`

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}
}
