package cmd_test

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	. "git.rob.mx/nidito/joao/cmd"
	"git.rob.mx/nidito/joao/internal/testdata"
	"git.rob.mx/nidito/joao/internal/testdata/opconnect"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/spf13/cobra"
)

func TestFlush(t *testing.T) {
	testdata.EnableDebugLogging()
	testdata.MockOPConnect(t)
	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("redact", false, "")
	cmd.SetOut(out)
	cmd.SetErr(out)

	Flush.SetBindings()
	Flush.Cobra = cmd
	err := Flush.Run(cmd, []string{testdata.YAML("test")})

	if err != nil {
		t.Fatalf("could not flush: %s", err)
	}

	expected := ""

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}

	item, err := opconnect.Get("some:test", "example")
	if err != nil {
		t.Fatalf("unexpected error getting flushed config: %s", err)
	}

	cfg, err := config.FromOP(item)
	if err != nil {
		t.Fatalf("unexpected error translating flushed config: %s", err)
	}

	serialized, err := cfg.AsYAML()
	if err != nil {
		t.Fatalf("unexpected error serializing config as yaml: %s", err)
	}

	data, err := os.ReadFile(testdata.YAML("test"))
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %s", err)
	}
	if bytes.Equal(serialized, data) {
		t.Fatalf("did not get expected serialization after flush.\n wanted:\n%s\n\ngot:\n%s", serialized, data)
	}
}

func TestFlushRedacted(t *testing.T) {
	testdata.EnableDebugLogging()
	testdata.MockOPConnect(t)
	out := &bytes.Buffer{}
	cmd := &cobra.Command{}
	cmd.Flags().Bool("dry-run", false, "")
	cmd.Flags().Bool("redact", true, "")
	cmd.SetOut(out)
	cmd.SetErr(out)

	Flush.SetBindings()
	Flush.Cobra = cmd
	path, cleanup := testdata.TempYAML(t, "test")
	defer cleanup()
	err := Flush.Run(cmd, []string{path})

	if err != nil {
		t.Fatalf("could not flush: %s", err)
	}

	expected := ""

	if got := out.String(); strings.TrimSpace(got) != expected {
		t.Fatalf("did not get expected output:\nwanted: %s\ngot: %s", expected, got)
	}

	item, err := opconnect.Get("some:test", "example")
	if err != nil {
		t.Fatalf("unexpected error getting flushed config: %s", err)
	}

	cfg, err := config.FromOP(item)
	if err != nil {
		t.Fatalf("unexpected error translating flushed config: %s", err)
	}

	serialized, err := cfg.AsYAML(config.OutputModeRedacted)
	if err != nil {
		t.Fatalf("unexpected error serializing redacted config as yaml: %s", err)
	}

	data, err := os.ReadFile(testdata.YAML("test"))
	if err != nil {
		t.Fatalf("unexpected error reading fixture: %s", err)
	}

	pat := regexp.MustCompile(`!!secret\.+\n`)
	redactedData := pat.ReplaceAll(data, []byte("!!secret\n"))
	if bytes.Equal(serialized, redactedData) {
		t.Fatalf("did not get expected redacted serialization after flush.\n wanted:\n%s\n\ngot:\n%s", serialized, redactedData)
	}
}
