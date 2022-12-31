// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config_test

import (
	"encoding/json"
	"strings"
	"testing"

	"git.rob.mx/nidito/joao/pkg/config"
)

const testYAML = `
_config: !!joao
  vault: example
  name: test
string: asdf
int: 1
float: 3.14
bool: true
secret: !!secret --secret--
list:
  - zero
  - one
map:
  key: value
`

func TestAsJSON(t *testing.T) {
	cfg, err := config.FromYAML([]byte(testYAML))
	if err != nil {
		t.Fatalf("Could not initialize test config: %s", err)
	}

	bytes, err := cfg.AsJSON(false, false)
	if err != nil {
		t.Fatalf("could not encode as json: %s", err)
	}

	expected, err := json.Marshal(map[string]any{
		"string": "asdf",
		"int":    1,
		"float":  3.14,
		"bool":   true,
		"secret": "--secret--",
		"list":   []any{"zero", "one"},
		"map":    map[string]string{"key": "value"},
	})
	if err != nil {
		t.Fatalf("could not encode as json: %s", err)
	}

	if string(bytes) != string(expected) {
		t.Fatalf("wanted secrets:\n%s\n---\ngot:\n%s", expected, bytes)
	}

	redactedBytes, err := cfg.AsJSON(true, false)
	if err != nil {
		t.Fatalf("could not encode as json: %s", err)
	}

	expectedRedacted := strings.Replace(string(expected), "--secret--", "", 1)
	if string(redactedBytes) != expectedRedacted {
		t.Fatalf("wanted redacted:\n %s\n---\ngot:\n%s", expectedRedacted, redactedBytes)
	}
}
