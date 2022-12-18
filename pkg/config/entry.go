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
package config

import (
	"fmt"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func isNumeric(s string) bool {
	for _, v := range s {
		if v < '0' || v > '9' {
			return false
		}
	}
	return true
}

// Entry is a configuration entry.
// Basically a copy of a yaml.Node with extra methods.
type Entry struct {
	Value       string
	Kind        yaml.Kind
	Tag         string
	Path        []string
	Content     []*Entry
	Style       yaml.Style
	FootComment string
	LineComment string
	HeadComment string
	Line        int
	Column      int
	// The ShortTag
	Type string
}

func NewEntry(name string, kind yaml.Kind) *Entry {
	return &Entry{
		Content: []*Entry{},
		Value:   name,
		Kind:    kind,
	}
}

func (e *Entry) copyFromNode(n *yaml.Node) {
	if e.Content == nil {
		e.Content = []*Entry{}
	}

	e.Kind = n.Kind
	e.Value = n.Value
	e.Tag = n.Tag
	e.Style = n.Style
	e.HeadComment = n.HeadComment
	e.LineComment = n.LineComment
	e.FootComment = n.FootComment
	e.Line = n.Line
	e.Column = n.Column
	e.Type = n.ShortTag()
}

func (e *Entry) String() string {
	return e.Value
}

func (e *Entry) ChildNamed(name string) *Entry {
	for _, child := range e.Content {
		if child.Name() == name {
			return child
		}
	}
	return nil
}

func (e *Entry) SetPath(parent []string, current string) {
	e.Path = append(parent, current) // nolint: gocritic
	switch e.Kind {
	case yaml.MappingNode, yaml.DocumentNode:
		for idx := 0; idx < len(e.Content); idx += 2 {
			key := e.Content[idx]
			child := e.Content[idx+1]
			child.SetPath(e.Path, key.Value)
		}
	case yaml.SequenceNode:
		for idx, child := range e.Content {
			child.Path = append(e.Path, fmt.Sprintf("%d", idx)) // nolint: gocritic
		}
	}
}

func (e *Entry) UnmarshalYAML(node *yaml.Node) error {
	e.copyFromNode(node)

	switch node.Kind {
	case yaml.SequenceNode, yaml.ScalarNode:
		for _, n := range node.Content {
			sub := &Entry{}
			sub.copyFromNode(n)
			if err := n.Decode(&sub); err != nil {
				return err
			}
			sub.SetPath(e.Path, n.Value)
			e.Content = append(e.Content, sub)
		}
	case yaml.DocumentNode, yaml.MappingNode:
		for idx := 0; idx < len(node.Content); idx += 2 {
			keyNode := node.Content[idx]
			valueNode := node.Content[idx+1]
			key := NewEntry("", keyNode.Kind)
			value := NewEntry(keyNode.Value, keyNode.Kind)
			if err := keyNode.Decode(key); err != nil {
				logrus.Errorf("decode map key: %s", keyNode.Value)
				return err
			}

			if err := valueNode.Decode(value); err != nil {
				logrus.Errorf("decode map key: %s", keyNode.Value)
				return err
			}
			if valueNode.Tag == YAMLTypeMetaConfig {
				key.Type = YAMLTypeMetaConfig
			}
			value.SetPath(e.Path, key.Value)
			e.Content = append(e.Content, key, value)
		}
	default:
		return fmt.Errorf("unknown yaml type: %v", node.Kind)
	}
	return nil
}

func (e *Entry) IsScalar() bool {
	return e.Kind != yaml.DocumentNode && e.Kind != yaml.MappingNode && e.Kind != yaml.SequenceNode
}

func (e *Entry) IsSecret() bool {
	return e.Tag == YAMLTypeSecret
}

func (e *Entry) TypeStr() string {
	if e.IsSecret() {
		return "secret"
	}

	switch e.Type {
	case "!!bool":
		return "bool"
	case "!!int":
		return "int"
	case "!!float":
		return "float"
	}

	return ""
}

func (e *Entry) asNode() *yaml.Node {
	return &yaml.Node{
		Kind:        e.Kind,
		Style:       e.Style,
		Tag:         e.Tag,
		Value:       e.Value,
		HeadComment: e.HeadComment,
		LineComment: e.LineComment,
		FootComment: e.FootComment,
		Line:        e.Line,
		Column:      e.Column,
		Content:     []*yaml.Node{},
	}
}

func (e *Entry) MarshalYAML() (*yaml.Node, error) {
	n := e.asNode()
	if n.Kind == yaml.ScalarNode {
		if redactOutput && e.IsSecret() {
			return &yaml.Node{
				Kind:        n.Kind,
				Style:       yaml.TaggedStyle & n.Style,
				Tag:         n.Tag,
				Value:       "",
				HeadComment: n.HeadComment,
				LineComment: n.LineComment,
				FootComment: n.FootComment,
				Line:        n.Line,
				Column:      n.Column,
			}, nil
		}
		return n, nil
	}

	for _, v := range e.Content {
		node, err := v.MarshalYAML()
		if err != nil {
			return nil, err
		}
		n.Content = append(n.Content, node)
	}
	return n, nil
}

func (e *Entry) FromOP(fields []*op.ItemField) error {
	annotations := map[string]string{}
	data := map[string]string{}

	for _, field := range fields {
		label := field.Label
		if field.Section != nil {
			if field.Section.Label == "~annotations" {
				annotations[label] = field.Value
				continue
			} else {
				label = field.Section.Label + "." + label
			}
		}
		if label == "password" || label == "notesPlain" {
			continue
		}
		data[label] = field.Value
	}

	for label, valueStr := range data {
		var style yaml.Style
		var tag string

		if annotations[label] == "secret" {
			style = yaml.TaggedStyle
			tag = YAMLTypeSecret
		}

		path := strings.Split(label, ".")
		container := e

		for idx, key := range path {
			if idx == len(path)-1 {
				container.Content = append(container.Content, NewEntry(key, yaml.ScalarNode), &Entry{
					Path:  path,
					Kind:  yaml.ScalarNode,
					Value: valueStr,
					Style: style,
					Tag:   tag,
				})
				break
			}

			subContainer := container.ChildNamed(key)
			if subContainer != nil {
				container = subContainer
			} else {
				kind := yaml.MappingNode
				if isNumeric(key) {
					kind = yaml.SequenceNode
				}
				child := NewEntry(key, kind)
				child.Path = append(container.Path, key) // nolint: gocritic
				if isNumeric(key) {
					container.Content = append(container.Content, child)
				} else {
					container.Content = append(container.Content, NewEntry(child.Name(), child.Kind), child)
				}
				container = child
			}
		}
	}

	return nil
}

func (e *Entry) ToOP() []*op.ItemField {
	ret := []*op.ItemField{}
	var section *op.ItemSection

	if e.Kind == yaml.ScalarNode {
		name := e.Path[len(e.Path)-1]
		fullPath := strings.Join(e.Path, ".")
		if len(e.Path) > 1 {
			section = &op.ItemSection{ID: e.Path[0]}
			name = strings.Join(e.Path[1:], ".")
		}

		fieldType := "STRING"
		if e.IsSecret() {
			fieldType = "CONCEALED"
		}

		if annotationType := e.TypeStr(); annotationType != "" {
			ret = append(ret, &op.ItemField{
				ID:      "~annotations." + fullPath,
				Section: annotationsSection,
				Label:   fullPath,
				Type:    "STRING",
				Value:   annotationType,
			})
		}

		ret = append(ret, &op.ItemField{
			ID:      fullPath,
			Section: section,
			Label:   name,
			Type:    fieldType,
			Value:   e.Value,
		})
		return ret
	}

	if e.Kind == yaml.SequenceNode {
		ret := []*op.ItemField{}
		for _, child := range e.Content {
			ret = append(ret, child.ToOP()...)
		}
		return ret
	}

	for i := 0; i < len(e.Content); i += 2 {
		child := e.Content[i+1]
		if child.Type == YAMLTypeMetaConfig {
			continue
		}
		ret = append(ret, child.ToOP()...)
	}
	return ret
}

func (e *Entry) Name() string {
	if e.Path == nil || len(e.Path) == 0 {
		return ""
	}
	return e.Path[len(e.Path)-1]
}

func (e *Entry) AsMap() any {
	if len(e.Content) == 0 {
		if redactOutput && e.IsSecret() {
			return ""
		}
		return e.Value
	}

	if e.Kind == yaml.SequenceNode {
		ret := []any{}
		for _, child := range e.Content {
			ret = append(ret, child.AsMap())
		}
		return ret
	}

	ret := map[string]any{}
	for idx, child := range e.Content {
		if idx%2 == 0 {
			continue
		}
		ret[child.Name()] = child.AsMap()
	}
	return ret
}

func (e *Entry) Merge(other *Entry) error {
	if e.IsScalar() && other.IsScalar() {
		e.Value = other.Value
		e.Tag = other.Tag
		e.Kind = other.Kind
		e.Type = other.Type
		return nil
	}

	if e.Kind == yaml.MappingNode || e.Kind == yaml.DocumentNode {
		for i := 0; i < len(other.Content); i += 2 {
			remote := other.Content[i+1]
			local := e.ChildNamed(remote.Name())
			if local != nil {
				if err := local.Merge(remote); err != nil {
					return err
				}
			} else {
				e.Content = append(e.Content, NewEntry(remote.Name(), remote.Kind), remote)
			}
		}
		return nil
	}

	for _, remote := range other.Content {
		local := other.ChildNamed(remote.Name())
		if local != nil {
			if err := local.Merge(remote); err != nil {
				return err
			}
		} else {
			logrus.Debugf("adding new collection value at %s", remote.Path)
			local.Content = append(local.Content, remote)
		}
	}

	return nil
}
