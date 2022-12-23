// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
