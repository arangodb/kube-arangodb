//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"time"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tolerations"
)

// CreatePodTolerations creates a list of tolerations for a pod created for the given group.
func CreatePodTolerations(mode api.DeploymentMode, group api.ServerGroup) []core.Toleration {
	notReadyDur := tolerations.TolerationDuration{Forever: false, TimeSpan: time.Minute}
	unreachableDur := tolerations.TolerationDuration{Forever: false, TimeSpan: time.Minute}
	switch group {
	case api.ServerGroupAgents:
		notReadyDur.Forever = true
		unreachableDur.Forever = true
	case api.ServerGroupCoordinators:
		notReadyDur.TimeSpan = 15 * time.Second
		unreachableDur.TimeSpan = 15 * time.Second
	case api.ServerGroupDBServers:
		notReadyDur.TimeSpan = 5 * time.Minute
		unreachableDur.TimeSpan = 5 * time.Minute
	case api.ServerGroupSingle:
		if mode == api.DeploymentModeSingle {
			notReadyDur.Forever = true
			unreachableDur.Forever = true
		} else {
			notReadyDur.TimeSpan = 5 * time.Minute
			unreachableDur.TimeSpan = 5 * time.Minute
		}
	case api.ServerGroupSyncMasters:
		notReadyDur.TimeSpan = 15 * time.Second
		unreachableDur.TimeSpan = 15 * time.Second
	case api.ServerGroupSyncWorkers:
		notReadyDur.TimeSpan = 1 * time.Minute
		unreachableDur.TimeSpan = 1 * time.Minute
	case api.ServerGroupGateways:
		notReadyDur.TimeSpan = 15 * time.Second
		unreachableDur.TimeSpan = 15 * time.Second
	}
	return []core.Toleration{tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeNotReady, notReadyDur),
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeUnreachable, unreachableDur),
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeAlphaUnreachable, unreachableDur),
	}
}
