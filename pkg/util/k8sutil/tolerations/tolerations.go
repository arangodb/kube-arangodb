//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package tolerations

import (
	"time"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

const (
	TolerationKeyNodeNotReady         = "node.kubernetes.io/not-ready"
	TolerationKeyNodeAlphaUnreachable = "node.alpha.kubernetes.io/unreachable"
	TolerationKeyNodeUnreachable      = "node.kubernetes.io/unreachable"
)

// TolerationDuration is a duration spec for tolerations.
type TolerationDuration struct {
	Forever  bool
	TimeSpan time.Duration
}

// NewNoExecuteToleration is a helper to create a Toleration with
// Key=key, Operator='Exists' Effect='NoExecute', TolerationSeconds=tolerationDuration.Seconds().
func NewNoExecuteToleration(key string, duration TolerationDuration) core.Toleration {
	t := core.Toleration{
		Key:      key,
		Operator: "Exists",
		Effect:   "NoExecute",
	}
	if !duration.Forever {
		tolerationSeconds := int64(duration.TimeSpan.Seconds())
		t.TolerationSeconds = &tolerationSeconds
	}
	return t
}

func CopyTolerations(source []core.Toleration) []core.Toleration {
	out := make([]core.Toleration, len(source))

	for id := range out {
		source[id].DeepCopyInto(&out[id])
	}

	return out
}

// MergeTolerationsIfNotFound merge the given tolerations lists, if no such toleration has been set in the given source.
func MergeTolerationsIfNotFound(source []core.Toleration, toAdd ...[]core.Toleration) []core.Toleration {
	for _, toleration := range toAdd {
		source = AddTolerationsIfNotFound(source, toleration...)
	}

	return source
}

// AddTolerationsIfNotFound add the given tolerations, if no such toleration has been set in the given source.
func AddTolerationsIfNotFound(source []core.Toleration, toAdd ...core.Toleration) []core.Toleration {
	for _, toleration := range toAdd {
		source = AddTolerationIfNotFound(source, toleration)
	}

	return source
}

// AddTolerationIfNotFound adds the given tolerations, if no such toleration has been set in the given source.
func AddTolerationIfNotFound(source []core.Toleration, toAdd core.Toleration) []core.Toleration {
	if len(source) == 0 {
		return []core.Toleration{
			toAdd,
		}
	}

	// Ensure we are working on the copy
	source = CopyTolerations(source)

	for id, t := range source {
		if t.Key == toAdd.Key && t.Effect == toAdd.Effect && t.Operator == toAdd.Operator && t.Value == toAdd.Value {
			// We are on same toleration, only value needs to be modified
			toAdd.DeepCopyInto(&source[id])

			return source
		}
	}

	return append(source, toAdd)
}

// CreatePodTolerations creates a list of tolerations for a pod created for the given group.
func CreatePodTolerations(mode api.DeploymentMode, group api.ServerGroup) []core.Toleration {
	notReadyDur := TolerationDuration{Forever: false, TimeSpan: time.Minute}
	unreachableDur := TolerationDuration{Forever: false, TimeSpan: time.Minute}
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
	}
	return []core.Toleration{NewNoExecuteToleration(TolerationKeyNodeNotReady, notReadyDur),
		NewNoExecuteToleration(TolerationKeyNodeUnreachable, unreachableDur),
		NewNoExecuteToleration(TolerationKeyNodeAlphaUnreachable, unreachableDur),
	}
}
