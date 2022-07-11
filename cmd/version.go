//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/version"
)

func init() {
	cmdMain.AddCommand(cmdVersion)
}

var cmdVersion = &cobra.Command{
	Use: "version",
	Run: versionRun,
}

func versionRun(cmd *cobra.Command, args []string) {
	v := version.GetVersionV1()
	println(fmt.Sprintf("Version: %s %s, Build: %s, Go: %s, Build Date: %s", v.Edition.Title(), v.Version, v.Build, v.GoVersion, v.BuildDate))
}
