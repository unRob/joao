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
package opClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	op "github.com/1Password/connect-sdk-go/onepassword"
)

func Fetch(vault, name string) (*op.Item, error) {
	return fetchRemote(name, vault)
}

func fetchRemote(name, vault string) (*op.Item, error) {
	cmd := exec.Command("op", "item", "--format", "json", "--vault", vault, "get", name)

	cmd.Env = os.Environ()
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	if cmd.ProcessState.ExitCode() > 0 {
		return nil, fmt.Errorf("op exited with %d: %s", cmd.ProcessState.ExitCode(), stderr.Bytes())
	}

	var item *op.Item
	if err := json.Unmarshal(stdout.Bytes(), &item); err != nil {
		return nil, err
	}

	return item, nil
}
