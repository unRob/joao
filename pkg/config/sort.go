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
