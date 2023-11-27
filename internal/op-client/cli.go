// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
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
	"github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

// Path points to the op binary.
var Path = "op"
var probedVersionModern = false
var versionConstraint = version.MustConstraints(version.NewConstraint(">= 2.23"))
var Exec ExecFunc = DefaultExec

type ExecFunc func(program string, args []string, stdin *bytes.Buffer) (bytes.Buffer, error)

func DefaultExec(program string, args []string, stdin *bytes.Buffer) (stdout bytes.Buffer, err error) {
	cmd := exec.Command(Path, args...)

	cmd.Env = os.Environ()
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if stdin != nil {
		cmd.Stdin = stdin
	}

	if err = cmd.Run(); err != nil {
		return stderr, fmt.Errorf("op exited with %s:\n%s", err, stderr.Bytes())
	}
	if cmd.ProcessState.ExitCode() > 0 {
		return stderr, fmt.Errorf("op exited with %d: %s", cmd.ProcessState.ExitCode(), stderr.Bytes())
	}

	return stdout, nil
}

type CLI struct {
	DryRun bool // Won't write to 1Password
}

func invoke(dryRun bool, vault string, stdin *bytes.Buffer, args ...string) (bytes.Buffer, error) {
	if vault != "" {
		args = append([]string{"--vault", shellescape.Quote(vault)}, args...)
	}

	argString := strings.Join(args, " ")
	if dryRun {
		logrus.Warnf("dry-run: Would have invoked `op %s`", argString)
		logrus.Tracef("dry-run: stdin `%s`", stdin)
		return bytes.Buffer{}, nil
	}

	logrus.Debugf("running `%s %s`", Path, argString)
	logrus.Tracef("stdin `%s`", stdin)
	return Exec(Path, args, stdin)
}

func (b *CLI) Get(vault, name string) (*op.Item, error) {
	stdout, err := invoke(false, vault, nil, "item", "--format", "json", "get", name)
	if err != nil {
		return nil, err
	}

	var item *op.Item
	if err := json.Unmarshal(stdout.Bytes(), &item); err != nil {
		return nil, err
	}

	return item, nil
}

func (b *CLI) Create(item *op.Item) error {
	logrus.Infof("Creating new item: %s/%s", item.Vault.ID, item.Title)

	itemJSON, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("could not serialize op item into json: %w", err)
	}

	stdin := bytes.NewBuffer(itemJSON)

	_, err = invoke(b.DryRun, item.Vault.ID, stdin, "item", "create")
	if err != nil {
		return fmt.Errorf("could not create item: %w", err)
	}

	logrus.Infof("Item %s/%s created", item.Vault.ID, item.Title)
	return nil
}

func (b *CLI) Update(item *op.Item, remote *op.Item) error {
	res, err := invoke(false, "", nil, "--version")
	if err == nil {
		v := strings.TrimSpace(res.String())
		current, err := version.NewVersion(v)
		if err == nil {
			probedVersionModern = versionConstraint.Check(current)
		} else {
			logrus.Debugf("Failed parsing version <%s>: %s", v, err)
		}
	}

	if probedVersionModern {
		return b.UpdateModern(item, remote)
	}

	return b.DeprecatedUpdate(item, remote)
}

func (b *CLI) UpdateModern(updated *op.Item, original *op.Item) error {
	updated.ID = original.ID
	updated.Vault = original.Vault
	itemJSON, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf("could not serialize op item into json: %w", err)
	}

	stdin := bytes.NewBuffer(itemJSON)
	_, err = invoke(b.DryRun, updated.Vault.ID, stdin, "item", "edit", original.Title)
	if err != nil {
		return fmt.Errorf("could not create item: %w", err)
	}

	logrus.Infof("Item %s/%s updated", updated.Vault.ID, updated.Title)
	return nil
}

func (b *CLI) DeprecatedUpdate(updated *op.Item, original *op.Item) error {
	logrus.Warnf("Using op-cli < 2.23 means using a hack for updating items, please upgrade ASAP!")
	args := []string{"item", "edit", updated.Title, "--"}
	localKeys := map[string]int{}

	for _, field := range updated.Fields {
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

	for _, field := range original.Fields {
		key := keyForField(field)
		if _, exists := localKeys[key]; !exists {
			logrus.Debugf("Deleting remote key %s", key)
			args = append(args, key+"[delete]=")
		}
	}

	if b.DryRun {
		logrus.Warnf("dry-run: Would have invoked op %v", args)
		return nil
	}
	stdout, err := invoke(b.DryRun, updated.Vault.ID, nil, args...)
	if err != nil {
		logrus.Errorf("op stderr: %s", stdout.String())
		return err
	}
	logrus.Infof("Item %s/%s updated", updated.Vault.ID, updated.Title)
	return nil
}

func (b *CLI) List(vault, prefix string) ([]string, error) {
	return nil, nil
}
