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
	_ "git.rob.mx/nidito/joao/cmd"
	"git.rob.mx/nidito/joao/internal/registry"
	"git.rob.mx/nidito/joao/internal/runtime"
	"github.com/sirupsen/logrus"
)

var version = "dev"

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		DisableTimestamp:       true,
		ForceColors:            runtime.ColorEnabled(),
	})

	logrus.SetLevel(logrus.DebugLevel)

	err := registry.Execute(version)

	if err != nil {
		logrus.Fatal(err)
	}
}
