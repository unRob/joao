// Copyright © 2022 Roberto Hidalgo <joao@un.rob.mx>
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"

	"git.rob.mx/nidito/chinampa"
	"git.rob.mx/nidito/chinampa/pkg/runtime"
	_ "git.rob.mx/nidito/joao/cmd"
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

	if err := chinampa.Execute(chinampa.Config{
		Name:        "joao",
		Summary:     "Helps organize config for roberto",
		Description: `﹅joao﹅ makes yaml, json, 1password and vault play along nicely.`,
		Version:     version.Version,
	}); err != nil {
		logrus.Errorf("total failure: %s", err)
		os.Exit(2)
	}
}
