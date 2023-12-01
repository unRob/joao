package opclient_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"git.rob.mx/nidito/joao/internal/testdata"
	opclient "git.rob.mx/nidito/joao/pkg/op-client"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	code := m.Run()
	opclient.Exec = opclient.DefaultExec
	os.Exit(code)
}

func fieldify(item *onepassword.Item) *onepassword.Item {
	item.Fields = []*onepassword.ItemField{
		{
			ID:      "password",
			Type:    "CONCEALED",
			Purpose: "PASSWORD",
			Label:   "password",
			Value:   "checksum",
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
	}

	return item
}

func TestDeprecatedUpdate(t *testing.T) {
	client := &opclient.CLI{}
	queriedVersion := false
	var calledArgs []string
	var calledStdin *bytes.Buffer
	opclient.Exec = func(program string, args []string, stdin *bytes.Buffer) (bytes.Buffer, error) {
		if strings.Contains(strings.Join(args, " "), "--version") {
			queriedVersion = true
			return *bytes.NewBufferString("1.0"), nil
		}

		calledArgs = args
		calledStdin = stdin
		return *bytes.NewBufferString("updated"), nil
	}

	original := fieldify(testdata.NewTestConfig("some:test"))

	updated := fieldify(testdata.NewTestConfig(original.Title))
	updated.Fields[2].Value = "42"
	updated.Fields[4].Value = "42"

	err := client.Update(updated, original)
	if err != nil {
		t.Fatalf("Failed in test update: %s", err)
	}

	if !queriedVersion {
		t.Fatalf("client did not query for version")
	}

	gotArgs := strings.Join(calledArgs, " ")
	wantedArgs := "--vault example item edit some:test -- password[password]=checksum ~annotations.int[text]=int int[text]=42 ~annotations.nested\\.int[text]=int nested.int[text]=42"
	if gotArgs != wantedArgs {
		t.Fatalf("client called unexpected arguments.\nwant: %s\n got: %s", wantedArgs, gotArgs)
	}

	var wantedStdin *bytes.Buffer
	if wantedStdin.String() != calledStdin.String() {
		t.Fatalf("client called unexpected stdin.\nwant: %s\n got: %s", wantedStdin, calledStdin)
	}
}

func TestUpdate(t *testing.T) {
	testdata.EnableDebugLogging()
	client := &opclient.CLI{}
	queriedVersion := false
	var calledArgs []string
	var calledStdin *bytes.Buffer
	execCalled := false
	opclient.Exec = func(program string, args []string, stdin *bytes.Buffer) (bytes.Buffer, error) {
		logrus.Debugf("Called exec with %+v", args)
		execCalled = true
		if strings.Contains(strings.Join(args, " "), "--version") {
			queriedVersion = true
			return *bytes.NewBufferString("2.30"), nil
		}

		calledArgs = args
		calledStdin = stdin
		return *bytes.NewBufferString("updated"), nil
	}

	original := fieldify(testdata.NewTestConfig("some:test"))

	updated := fieldify(testdata.NewTestConfig(original.Title))
	updated.Fields[2].Value = "42"
	updated.Fields[4].Value = "42"

	err := client.Update(updated, original)
	if err != nil {
		t.Fatalf("Failed in test update: %s", err)
	}

	if !execCalled {
		t.Fatalf("client did not query for version")
	}

	if !queriedVersion {
		t.Fatalf("client did not query for version")
	}

	gotArgs := strings.Join(calledArgs, " ")
	wantedArgs := "--vault example item edit some:test"
	if gotArgs != wantedArgs {
		t.Fatalf("client called unexpected arguments.\nwant: %s\n got: %s", wantedArgs, gotArgs)
	}

	updatedJSON, err := json.Marshal(updated)
	if err != nil {
		t.Fatalf("Could not encode json: %s", err)
	}
	wantedStdin := bytes.NewBuffer(updatedJSON)
	if wantedStdin.String() != calledStdin.String() {
		t.Fatalf("client called unexpected stdin.\nwant: %s\n got: %s", wantedStdin, calledStdin)
	}
}
