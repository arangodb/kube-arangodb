//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"fmt"
	"path/filepath"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

// ArangodVersionCheckInitContainer creates a container configured to check version.
func ArangodVersionCheckInitContainer(name, executable, operatorImage string, version driver.Version, securityContext *core.SecurityContext) core.Container {
	versionFile := filepath.Join(shared.ArangodVolumeMountDir, "VERSION-1")
	var command = []string{
		executable,
		"init-containers",
		"version-check",
		"--path",
		versionFile,
	}

	if v := version.Major(); v > 0 {
		command = append(command,
			"--major",
			fmt.Sprintf("%d", v))

		if v := version.Minor(); v > 0 {
			command = append(command,
				"--minor",
				fmt.Sprintf("%d", v))
		}
	}

	volumes := []core.VolumeMount{
		ArangodVolumeMount(),
	}
	return operatorInitContainer(name, operatorImage, command, securityContext, volumes)
}
