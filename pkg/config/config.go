// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var annotationsSection = &op.ItemSection{
	ID:    "~annotations",
	Label: "~annotations",
}
var defaultItemFields = []*op.ItemField{
	{
		ID:      "password",
		Type:    "CONCEALED",
		Purpose: "PASSWORD",
		Label:   "password",
		Value:   "hash",
	}, {
		ID:      "notesPlain",
		Type:    "STRING",
		Purpose: "NOTES",
		Label:   "notesPlain",
		Value:   "flushed by joao",
	},
}

type Config struct {
	Vault string
	Name  string
	Tree  *Entry
}

// Delete a value at path.
func (cfg *Config) Delete(path []string) error {
	parent := cfg.Tree

	for idx, key := range path {
		if len(path)-1 == idx {
			newContents := []*Entry{}
			found := false
			for idx, child := range parent.Content {
				if child.Name() == key {
					found = true
					logrus.Debugf("Deleting %s", strings.Join(path, "."))
					if parent.Kind == yaml.DocumentNode || parent.Kind == yaml.MappingNode {
						newContents = newContents[0 : idx-1]
					}
					continue
				}
				newContents = append(newContents, child)
			}

			if !found {
				return fmt.Errorf("no value found at %s", key)
			}

			parent.Content = newContents
			break
		}

		parent = parent.ChildNamed(key)
		if parent == nil {
			return fmt.Errorf("no value found at %s", key)
		}
	}

	return nil
}

// Set a new value, optionally parsing the supplied bytes as a secret or a JSON-encoded value.
func (cfg *Config) Set(path []string, data []byte, isSecret, parseEntry bool) error {
	newEntry := NewEntry(path[len(path)-1], yaml.ScalarNode)
	newEntry.Path = path
	valueStr := string(data)
	newEntry.Value = valueStr

	if parseEntry {
		if err := yaml.Unmarshal(data, newEntry); err != nil {
			return err
		}
		if newEntry.Kind == yaml.MappingNode || newEntry.Kind == yaml.SequenceNode {
			newEntry.Style = yaml.FoldedStyle | yaml.LiteralStyle
			for _, v := range newEntry.Content {
				v.Style = yaml.FlowStyle
			}
		}
	} else {
		valueStr = strings.Trim(valueStr, "\n")
		if isSecret {
			newEntry.Style = yaml.TaggedStyle
			newEntry.Tag = YAMLTypeSecret
		}
		newEntry.Kind = yaml.ScalarNode
		newEntry.Value = valueStr

		if !strings.Contains(valueStr, "\n") {
			newEntry.Style &= yaml.LiteralStyle
		} else {
			newEntry.Style &= yaml.FlowStyle
		}
	}

	entry := cfg.Tree
	for idx, key := range path {
		if len(path)-1 == idx {
			if dst := entry.ChildNamed(key); dst == nil {
				key := NewEntry(key, yaml.ScalarNode)
				if entry.Kind == yaml.MappingNode {
					logrus.Infof("setting new map key %v", newEntry.Path)
					entry.Content = append(entry.Content, key, newEntry)
				} else {
					logrus.Infof("setting new list key %v", newEntry.Path)
					entry.Kind = yaml.SequenceNode
					entry.Content = append(entry.Content, newEntry)
				}
			} else {
				logrus.Infof("setting %v", newEntry.Path)
				dst.Value = newEntry.Value
				dst.Tag = newEntry.Tag
				dst.Style = newEntry.Style
			}
			break
		}

		if child := entry.ChildNamed(key); child != nil {
			entry = child
			continue
		}

		kind := yaml.MappingNode
		if idx+1 == len(path)-1 && isNumeric(path[idx+1]) {
			kind = yaml.SequenceNode
		}
		sub := NewEntry(key, kind)
		sub.Path = append(entry.Path, key) // nolint: gocritic

		keyEntry := NewEntry(sub.Name(), yaml.ScalarNode)
		keyEntry.Value = key
		entry.Content = append(entry.Content, keyEntry, sub)
		entry = sub
	}

	return nil
}

func (cfg *Config) Merge(other *Config) error {
	return cfg.Tree.Merge(other.Tree)
}

func (cfg *Config) DiffRemote(path string, stdout io.Writer, stderr io.Writer) error {
	remote, err := Load(path, true)
	if err != nil {
		return err
	}

	localBytes, err := cfg.AsYAML(OutputModeNoComments, OutputModeSorted, OutputModeNoConfig)
	if err != nil {
		return err
	}

	file1, cleanupLocalDiff, err := tempfile(localBytes)
	if err != nil {
		return err
	}
	defer cleanupLocalDiff()

	remoteBytes, err := remote.AsYAML(OutputModeNoComments, OutputModeSorted)
	if err != nil {
		return err
	}
	file2, cleanupRemoteDiff, err := tempfile(remoteBytes)
	if err != nil {
		return err
	}
	defer cleanupRemoteDiff()

	opPath := fmt.Sprintf("op://%s/%s", cfg.Vault, remote.Name)
	diff := exec.Command("diff", "-u", "-L", path, file1, "-L", opPath, file2)
	diff.Env = os.Environ()

	diff.Stdout = stdout
	diff.Stderr = stderr

	if err := diff.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			if diff.ProcessState.ExitCode() == 1 {
				return nil
			}
		}
		return fmt.Errorf("diff could not run: %w", err)
	}

	if diff.ProcessState.ExitCode() > 2 {
		return fmt.Errorf("diff exited with exit code %d", diff.ProcessState.ExitCode())
	}

	return nil
}

func tempfile(data []byte) (string, func(), error) {
	f, err := ioutil.TempFile("", "joao-diff")
	if err != nil {
		return "", nil, err
	}

	if _, err := f.Write(data); err != nil {
		return "", nil, err
	}

	if err := f.Close(); err != nil {
		return "", nil, err
	}

	deferFn := func() {
		if err := os.Remove(f.Name()); err != nil {
			logrus.Error(err)
		}
	}
	return f.Name(), deferFn, nil
}
