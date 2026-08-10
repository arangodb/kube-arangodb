//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"fmt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
)

// hardenArangodArgs returns the hardening arguments appended to the arangod containers when the harden
// feature is enabled.
func hardenArangodArgs(input pod.Input) []string {
	args := []string{
		"--server.harden",
		"--javascript.harden",
		"--javascript.startup-options-denylist=.*",
		"--javascript.environment-variables-allowlist=^HOSTNAME$",
		"--javascript.environment-variables-allowlist=^PATH$",
		"--log.api-enabled=jwt",
		"--backup.api-enabled=jwt",
	}

	// In cluster mode also constrain replication based on the number of DBServers (only once there are at
	// least 2): a default replication factor of min(DBServers, 3), a minimum replication factor of 2, and
	// a write concern of 2 when the default replication factor is 3.
	if input.Deployment.GetMode() == api.DeploymentModeCluster {
		if count := input.Deployment.DBServers.GetCount(); count >= 2 {
			rf := count
			if rf > 3 {
				rf = 3
			}

			args = append(args,
				fmt.Sprintf("--cluster.default-replication-factor=%d", rf),
				"--cluster.min-replication-factor=2",
			)

			// Write concern is set only when the default replication factor is 3.
			if rf >= 3 {
				args = append(args, "--cluster.write-concern=2")
			}
		}
	}

	return args
}
