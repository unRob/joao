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

import "fmt"

func (c *Config) Lookup(query []string) (*Entry, error) {
	if len(query) == 0 || len(query) == 1 && query[0] == "." {
		return c.Tree, nil
	}

	entry := c.Tree
	for _, part := range query {
		entry = entry.ChildNamed(part)
		if entry == nil {
			return nil, fmt.Errorf("value not found at %s of %s", part, query)
		}
	}

	return entry, nil
}
