// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"

	"git.rob.mx/nidito/chinampa"
	"git.rob.mx/nidito/chinampa/pkg/runtime"
	"git.rob.mx/nidito/joao/cmd"
	"git.rob.mx/nidito/joao/pkg/version"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
		ForceColors:            runtime.ColorEnabled(),
	})

	if runtime.DebugEnabled() {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Debug("Debugging enabled")
	}

	chinampa.Register(
		cmd.Get,
		cmd.Set,
		cmd.Diff,
		cmd.Fetch,
		cmd.Flush,
		cmd.Plugin,
	)
	chinampa.Register(cmd.GitFilters...)

	if err := chinampa.Execute(chinampa.Config{
		Name:    "joao",
		Summary: "A very WIP configuration manager",
		Description: `﹅joao﹅ makes yaml, json, 1Password and Hashicorp Vault play along nicely.

Keeps config entries encoded as YAML in the filesystem, backs it up to 1Password, and syncs scrubbed copies to git. Robots consume entries via 1Password Connect + Vault.

Schema for configuration and non-secret values live along the code, and are pushed to remote origins. Secrets can optionally and temporally be flushed to disk for editing or other sorts of operations. Git filters are available to prevent secrets from being pushed to remotes. Secrets are grouped into files, and every file gets its own 1Password item.

Secret values are specified using the ﹅!!secret﹅ YAML tag.
`,
		Version: version.Version,
	}); err != nil {
		logrus.Errorf("total failure: %s", err)
		os.Exit(2)
	}
}
