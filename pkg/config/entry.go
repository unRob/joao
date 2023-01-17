// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

import (
	"fmt"
	"sort"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

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
	if e.IsSecret() && yamlOutput.Has(OutputModeRedacted) {
		return ""
	}
	return e.Value
}

func (e *Entry) Contents() []*Entry {
	entries := []*Entry{}

	if (e.Kind == yaml.MappingNode || e.Kind == yaml.DocumentNode) && yamlOutput.Has(OutputModeSorted) {
		smes := []*sortedMapEntry{}
		for i := 0; i < len(e.Content); i += 2 {
			smes = append(smes, &sortedMapEntry{key: e.Content[i], value: e.Content[i+1]})
		}
		sortedEntries := &smec{smes}
		sort.Sort(sortedEntries)
		for _, v := range sortedEntries.smes {
			entries = append(entries, v.key, v.value)
		}
	} else {
		entries = e.Content
	}

	return entries
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
	if current != "." {
		e.Path = append([]string{}, parent...)
		e.Path = append(e.Path, current)
	}
	if e.IsScalar() {
		return
	}

	if e.Kind == yaml.SequenceNode {
		for idx, child := range e.Content {
			child.SetPath(e.Path, fmt.Sprintf("%d", idx))
		}
		return
	}

	for idx := 0; idx < len(e.Content); idx += 2 {
		key := e.Content[idx]
		child := e.Content[idx+1]
		child.SetPath(e.Path, key.Value)
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
	n := &yaml.Node{
		Kind:    e.Kind,
		Style:   e.Style,
		Tag:     e.Tag,
		Value:   e.String(),
		Line:    e.Line,
		Column:  e.Column,
		Content: []*yaml.Node{},
	}

	if !yamlOutput.Has(OutputModeNoComments) {
		n.HeadComment = e.HeadComment
		n.LineComment = e.LineComment
		n.FootComment = e.FootComment
	}

	if yamlOutput.Has(OutputModeStandardYAML) {
		if e.IsScalar() {
			if len(e.Path) > 0 {
				if !strings.Contains(n.Value, "\n") {
					n.Style &= yaml.FoldedStyle
				} else {
					n.Style &= yaml.FlowStyle
				}
			}
		} else {
			n.Style = yaml.FoldedStyle | yaml.LiteralStyle
			for _, v := range n.Content {
				v.Style = yaml.FlowStyle
			}
		}
	}

	return n
}

func (e *Entry) MarshalYAML() (*yaml.Node, error) {
	n := e.asNode()

	if e.Kind == yaml.SequenceNode {
		for _, v := range e.Content {
			node, err := v.MarshalYAML()
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, node)
		}
	} else if e.Kind == yaml.MappingNode || e.Kind == yaml.DocumentNode {
		entries := e.Contents()
		if len(entries)%2 != 0 {
			return nil, fmt.Errorf("cannot decode odd numbered contents list: %s", e.Path)
		}

		for i := 0; i < len(entries); i += 2 {
			key := entries[i]
			value := entries[i+1]
			if yamlOutput.Has(OutputModeNoConfig) && value.Type == YAMLTypeMetaConfig {
				continue
			}

			if key.Type == "" {
				key.Kind = yaml.ScalarNode
				key.Type = "!!map"
			}

			keyNode, err := key.MarshalYAML()
			if err != nil {
				return nil, err
			}

			node, err := value.MarshalYAML()
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, keyNode, node)
		}
	}

	return n, nil
}

func (e *Entry) FromOP(fields []*op.ItemField) error {
	annotations := map[string]string{}
	data := map[string]string{}
	entryKeys := []string{}

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
		entryKeys = append(entryKeys, label)
		data[label] = field.Value
	}

	for _, label := range entryKeys {
		valueStr := data[label]
		var style yaml.Style
		var tag string
		kind := ""

		if annotations[label] == "secret" {
			style = yaml.TaggedStyle
			tag = YAMLTypeSecret
		} else if k, ok := annotations[label]; ok {
			kind = "!!" + k
		}

		path := strings.Split(label, ".")
		container := e

		for idx, key := range path {
			if idx == len(path)-1 {
				if existing := container.ChildNamed(key); existing != nil {
					existing.Value = valueStr
					existing.Style = style
					existing.Tag = tag
					existing.Kind = yaml.ScalarNode
					existing.Path = path
					existing.Type = kind
					break
				}

				newEntry := &Entry{
					Path:  path,
					Kind:  yaml.ScalarNode,
					Value: valueStr,
					Style: style,
					Tag:   tag,
					Type:  kind,
				}
				if isNumeric(key) {
					logrus.Debugf("hydrating sequence value at %s", path)
					container.Kind = yaml.SequenceNode
					container.Content = append(container.Content, newEntry)
				} else {
					logrus.Debugf("hydrating map value at %s", path)
					keyEntry := NewEntry(key, yaml.ScalarNode)
					keyEntry.Value = key
					container.Content = append(container.Content, keyEntry, newEntry)
				}
				break
			}

			subContainer := container.ChildNamed(key)
			if subContainer != nil {
				container = subContainer
				continue
			}

			kind := yaml.MappingNode
			if idx+1 == len(path)-1 && isNumeric(path[idx+1]) {
				logrus.Debugf("creating sequence container for key %s at %s", key, path)
				kind = yaml.SequenceNode
			}
			child := NewEntry(key, kind)
			child.Path = append([]string{}, container.Path...)
			child.Path = append(child.Path, key)

			keyEntry := NewEntry(child.Name(), yaml.ScalarNode)
			keyEntry.Value = key
			container.Content = append(container.Content, keyEntry, child)
			container = child
		}
	}

	return nil
}

func (e *Entry) ToOP() []*op.ItemField {
	ret := []*op.ItemField{}
	var section *op.ItemSection

	if e.IsScalar() {
		name := e.Path[len(e.Path)-1]
		fullPath := strings.Join(e.Path, ".")
		if len(e.Path) > 1 {
			section = &op.ItemSection{ID: e.Path[0], Label: e.Path[0]}
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
	if e.IsScalar() {
		switch e.TypeStr() {
		case "bool":
			var boolVal bool
			err := e.asNode().Decode(&boolVal)
			if err != nil {
				panic(fmt.Sprintf("could not encode boolean at %s, %s", e.Path, err))
			}
			return boolVal
		case "int":
			var intVal uint64
			err := e.asNode().Decode(&intVal)
			if err != nil {
				panic(fmt.Sprintf("could not encode int at %s, %s", e.Path, err))
			}
			return intVal
		case "float":
			var floatVal float64
			err := e.asNode().Decode(&floatVal)
			if err != nil {
				panic(fmt.Sprintf("could not encode float at %s, %s", e.Path, err))
			}
			return floatVal
		}
		return e.String()
	}

	if e.Kind == yaml.SequenceNode {
		ret := []any{}
		for _, sub := range e.Content {
			ret = append(ret, sub.AsMap())
		}
		return ret
	}

	ret := map[string]any{}
	for idx := 1; idx < len(e.Content); idx += 2 {
		child := e.Content[idx]
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
		}
		local.Content = append(local.Content, remote)
	}

	return nil
}
