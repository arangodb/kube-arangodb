//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package cli

import (
	"github.com/spf13/cobra"

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
)

func Register(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVarP(&input.Namespace, "namespace", "n", utilConstants.NamespaceWithDefault("default"), "Kubernetes namespace")
	f.BoolVar(&input.HideSensitiveData, "hide-sensitive-data", true, "Hide sensitive data")
	f.BoolVar(&input.PodLogs, "pod-logs", true, "Collect pod logs")
	f.BoolVar(&input.DebugPackageFiles, "debug-package-files", false, "Collect Debug files from Storage")
}

var input Input

func GetInput() Input {
	return input
}

type Input struct {
	Namespace         string
	HideSensitiveData bool
	PodLogs           bool
	DebugPackageFiles bool
}
