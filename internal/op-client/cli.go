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
package opclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/alessio/shellescape"
	"github.com/sirupsen/logrus"
)

type CLI struct{}

func invoke(vault string, args ...string) (bytes.Buffer, error) {
	if vault != "" {
		args = append([]string{"--vault", shellescape.Quote(vault)}, args...)
	}

	argString := ""
	for _, arg := range args {
		parts := strings.Split(arg, "]=")
		if strings.HasSuffix(parts[0], "[password") {
			parts[1] = "*****"
			argString += fmt.Sprintf("%s]=%v", parts[0], parts[1])
		} else {
			argString += " " + arg
		}
	}
	logrus.Debugf("invoking op with args: %s", argString)
	cmd := exec.Command("op", args...)

	cmd.Env = os.Environ()
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr, err
	}
	if cmd.ProcessState.ExitCode() > 0 {
		return stderr, fmt.Errorf("op exited with %d: %s", cmd.ProcessState.ExitCode(), stderr.Bytes())
	}

	return stdout, nil
}

func (b *CLI) Get(vault, name string) (*op.Item, error) {
	stdout, err := invoke(vault, "item", "--format", "json", "get", name)
	if err != nil {
		return nil, err
	}

	var item *op.Item
	if err := json.Unmarshal(stdout.Bytes(), &item); err != nil {
		return nil, err
	}

	return item, nil
}

func (b *CLI) create(item *op.Item) error {
	logrus.Infof("Creating new item: %s/%s", item.Vault.ID, item.Title)
	cmd := exec.Command("op", "--vault", shellescape.Quote(item.Vault.ID), "item", "create")

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("could not serialize op item into json: %w", err)
	}

	cmd.Stdin = bytes.NewBuffer(itemJSON)
	cmd.Env = os.Environ()
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not create item: %w", err)
	}

	if cmd.ProcessState.ExitCode() > 0 {
		return fmt.Errorf("op exited with %d: %s", cmd.ProcessState.ExitCode(), stderr.Bytes())
	}
	logrus.Infof("Item %s/%s created", item.Vault.ID, item.Title)
	return nil
}

type hashResult int

const (
	HashItemError hashResult = iota
	HashItemMissing
	HashMatch
	HashMismatch
)

func keyForField(field *op.ItemField) string {
	name := strings.ReplaceAll(field.Label, ".", "\\.")
	if field.Section != nil {
		name = field.Section.ID + "." + name
	}
	return name
}

func (b *CLI) Update(vault, name string, item *op.Item) error {
	remote, err := b.Get(vault, name)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf("\"%s\" isn't an item in the ", name)) {
			return b.create(item)
		}

		return fmt.Errorf("could not fetch remote 1password item to compare against: %w", err)
	}

	if remote.GetValue("password") == item.GetValue("password") {
		logrus.Warn("item is already up to date")
		return nil
	}

	logrus.Infof("Item %s/%s already exists, updating", item.Vault.ID, item.Title)

	args := []string{"item", "edit", name, "--"}
	localKeys := map[string]int{}

	for _, field := range item.Fields {
		kind := ""
		if field.Type == "CONCEALED" {
			kind = "password"
		} else {
			kind = "text"
		}
		keyName := keyForField(field)
		key := fmt.Sprintf("%s[%s]", keyName, kind)
		args = append(args, fmt.Sprintf("%s=%s", key, field.Value))
		localKeys[keyName] = 1
	}

	for _, field := range remote.Fields {
		key := keyForField(field)
		if _, exists := localKeys[key]; !exists {
			logrus.Debugf("Deleting remote key %s", key)
			args = append(args, key+"[delete]=")
		}
	}

	stdout, err := invoke(vault, args...)
	if err != nil {
		logrus.Errorf("op stderr: %s", stdout.String())
		return err
	}
	logrus.Infof("Item %s/%s updated", item.Vault.ID, item.Title)
	return nil
}

func (b *CLI) List(vault, prefix string) ([]string, error) {
	return nil, nil
}
