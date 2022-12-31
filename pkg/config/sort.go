// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package config

type sortedMapEntry struct {
	key   *Entry
	value *Entry
}

type smec struct {
	smes []*sortedMapEntry
}

func (e smec) Len() int {
	return len(e.smes)
}

func (e smec) Less(i, j int) bool {
	return e.smes[i].value.Name() < e.smes[j].value.Name()
}

func (e smec) Swap(i, j int) {
	e.smes[i], e.smes[j] = e.smes[j], e.smes[i]
}
