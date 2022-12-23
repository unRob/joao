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

	if err := chinampa.Execute(version.Version); err != nil {
		logrus.Errorf("total failure: %s", err)
		os.Exit(2)
	}
}
