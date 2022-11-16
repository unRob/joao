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
	"strconv"
	"strings"

	op "github.com/1Password/connect-sdk-go/onepassword"
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

type Entry struct {
	Key        string
	Path       []string
	Value      interface{}
	Children   map[string]*Entry
	isSecret   bool
	isSequence bool
	node       *yaml.Node
}

func NewEntry(name string) *Entry {
	return &Entry{Key: name, Children: map[string]*Entry{}}
}

func (e *Entry) SetKey(key string, parent []string) {
	e.Path = append(parent, key)
	e.Key = key
	for k, child := range e.Children {
		child.SetKey(k, e.Path)
	}
}

func (e *Entry) UnmarshalYAML(node *yaml.Node) error {
	e.node = node
	switch node.Kind {
	case yaml.DocumentNode, yaml.MappingNode:
		if e.Children == nil {
			e.Children = map[string]*Entry{}
		}
		err := node.Decode(&e.Children)
		if err != nil {
			return err
		}

	case yaml.SequenceNode:
		list := []*Entry{}
		err := node.Decode(&list)
		if err != nil {
			return err
		}
		if e.Children == nil {
			e.Children = map[string]*Entry{}
		}
		for idx, child := range list {
			child.Key = fmt.Sprintf("%d", idx)
			e.Children[child.Key] = child
		}
		e.isSequence = true
	case yaml.ScalarNode:
		var val interface{}
		err := node.Decode(&val)
		if err != nil {
			return err
		}
		e.Value = val
		e.isSecret = node.Tag == "!!secret"
	default:
		return fmt.Errorf("unknown yaml type: %v", node.Kind)
	}
	return nil
}

func (e *Entry) MarshalYAML() (interface{}, error) {
	if len(e.Children) == 0 {
		if redactOutput && e.isSecret {
			n := e.node
			return &yaml.Node{
				Kind:        n.Kind,
				Style:       yaml.TaggedStyle,
				Tag:         n.Tag,
				Value:       "",
				HeadComment: n.HeadComment,
				LineComment: n.LineComment,
				FootComment: n.FootComment,
				Line:        n.Line,
				Column:      n.Column,
			}, nil
		}
		return e.node, nil
	}

	if e.isSequence {
		ret := make([]*Entry, len(e.Children))
		for k, child := range e.Children {
			idx, _ := strconv.Atoi(k)
			ret[idx] = child
		}
		return ret, nil
	}

	ret := map[string]*Entry{}
	for k, child := range e.Children {
		ret[k] = child
	}
	return ret, nil
}

func (e *Entry) FromOP(fields []*op.ItemField) error {
	annotations := map[string]string{}
	data := map[string]string{}
	labels := []string{}

	for i := 0; i < len(fields); i++ {
		field := fields[i]
		label := field.Label
		if field.Section != nil {
			if field.Section.Label == "~annotations" {
				annotations[label] = field.Value
				continue
			} else {
				label = field.Section.Label + "." + label
			}
		}
		labels = append(labels, label)
		data[label] = field.Value
	}

	for _, label := range labels {
		var value interface{} = data[label]

		if typeString, ok := annotations[label]; ok {
			switch typeString {
			case "bool":
				value = value == "true"
			case "int":
				var err error
				value, err = strconv.ParseInt(value.(string), 10, 64)
				if err != nil {
					return err
				}
			}
		}

		path := strings.Split(label, ".")
		container := e

		for idx, key := range path {
			if idx == len(path)-1 {
				if !isNumeric(key) {
					isSecretLabel := annotations[label]
					container.Children[key] = &Entry{
						Key:      key,
						Path:     path,
						Value:    value,
						isSecret: isSecretLabel == "!!secret",
					}
					break
				}

				holderI := container.Value
				if container.Value == nil {
					holderI = []interface{}{}
				}

				holder := holderI.([]interface{})
				container.Value = append(holder, value)
				break
			}

			subContainer, exists := container.Children[key]
			if exists {
				container = subContainer
			} else {
				container.Children[key] = NewEntry(key)
				container = container.Children[key]
			}
		}
	}

	return nil
}

func (e *Entry) ToOP(annotationsSection *op.ItemSection) []*op.ItemField {
	ret := []*op.ItemField{}
	var section *op.ItemSection
	name := e.Key
	if len(e.Path) > 1 {
		section = &op.ItemSection{ID: e.Path[0]}
		name = strings.Join(e.Path[1:], ".")
	}

	if e.isSecret {
		ret = append(ret, &op.ItemField{
			ID:      "~annotations." + strings.Join(e.Path, "."),
			Section: annotationsSection,
			Label:   name,
			Type:    "STRING",
			Value:   "secret",
		})
	} else if _, ok := e.Value.(bool); ok {
		ret = append(ret, &op.ItemField{
			ID:      "~annotations." + strings.Join(e.Path, "."),
			Section: annotationsSection,
			Label:   name,
			Type:    "STRING",
			Value:   "bool",
		})
	} else if _, ok := e.Value.(int); ok {
		ret = append(ret, &op.ItemField{
			ID:      "~annotations." + strings.Join(e.Path, "."),
			Section: annotationsSection,
			Label:   name,
			Type:    "STRING",
			Value:   "int",
		})
	} else if _, ok := e.Value.(float64); ok {
		ret = append(ret, &op.ItemField{
			ID:      "~annotations." + strings.Join(e.Path, "."),
			Section: annotationsSection,
			Label:   name,
			Type:    "STRING",
			Value:   "float",
		})
	}

	if len(e.Children) == 0 {
		ret = append(ret, &op.ItemField{
			ID:      strings.Join(e.Path, "."),
			Section: section,
			Label:   name,
			Type:    "STRING",
			Value:   fmt.Sprintf("%s", e.Value),
		})
	} else {
		for _, child := range e.Children {
			ret = append(ret, child.ToOP(annotationsSection)...)
		}
	}

	return ret
}

func (e *Entry) AsMap() interface{} {
	if len(e.Children) == 0 {
		if redactOutput && e.isSecret {
			return ""
		}
		return e.Value
	}

	if e.isSequence {
		ret := make([]interface{}, len(e.Children))
		for key, child := range e.Children {
			idx, _ := strconv.Atoi(key)
			ret[idx] = child.AsMap()
		}
		return ret
	}

	ret := map[string]interface{}{}
	for key, child := range e.Children {
		ret[key] = child.AsMap()
	}
	return ret
}
