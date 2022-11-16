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
package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	opClient "git.rob.mx/nidito/joao/internal/op-client"
	"git.rob.mx/nidito/joao/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	getCommand.Flags().StringP("output", "o", "raw", "the format to output in")
	getCommand.Flags().Bool("remote", false, "query 1password instead of the filesystem")
	getCommand.Flags().Bool("redacted", false, "do not print secrets")
}

var getCommand = &cobra.Command{
	Use:  "get CONFIG [--output|-o=(raw|json|yaml)] [--remote] [--redacted] [jq expr]",
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]
		query := ""
		if len(args) > 1 {
			query = args[1]
		}
		var cfg *config.Config
		remote, _ := cmd.Flags().GetBool("remote")

		isYaml := strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")
		if !remote && isYaml {
			buf, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("could not read file %s", path)
			}

			if len(buf) == 0 {
				buf = []byte("{}")
			}

			cfg, err = config.ConfigFromYAML(buf)
			if err != nil {
				return err
			}
		} else {
			name := path
			if isYaml {
				comps := strings.Split(path, "config/")
				name = strings.ReplaceAll(strings.Replace(comps[len(comps)-1], ".yaml", "", 1), "/", ":")
			}

			item, err := opClient.Get("nidito-admin", name)
			if err != nil {
				return err
			}

			cfg, err = config.ConfigFromOP(item)
			if err != nil {
				return err
			}
		}

		format, _ := cmd.Flags().GetString("output")
		redacted, _ := cmd.Flags().GetBool("redacted")

		if query == "" {
			switch format {
			case "yaml", "raw":
				bytes, err := cfg.AsYAML(redacted)
				if err != nil {
					return err
				}
				_, err = cmd.OutOrStdout().Write(bytes)
				return err
			case "json", "json-op":
				bytes, err := cfg.AsJSON(redacted, format == "json-op")
				if err != nil {
					return err
				}
				_, err = cmd.OutOrStdout().Write(bytes)
				return err
			}
			return fmt.Errorf("unknown format %s", format)
		}

		parts := strings.Split(query, ".")

		entry := cfg.Tree
		for _, part := range parts {
			entry = entry.Children[part]
			if entry == nil {
				return fmt.Errorf("value not found at %s of %s", part, query)
			}
		}

		var bytes []byte
		var err error
		if len(entry.Children) > 0 {
			val := entry.AsMap()
			if format == "yaml" {
				enc := yaml.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent(2)
				return enc.Encode(val)
			}

			bytes, err = json.Marshal(val)
			if err != nil {
				return err
			}
		} else {
			if valString, ok := entry.Value.(string); ok {
				bytes = []byte(valString)
			} else {
				bytes, err = json.Marshal(entry.Value)
				if err != nil {
					return err
				}
			}
		}

		_, err = cmd.OutOrStdout().Write(bytes)
		return err
	},
}
