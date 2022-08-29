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

package k8sutil

import (
	"time"

	core "k8s.io/api/core/v1"
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

// AddTolerationIfNotFound adds the given tolerations, if no such toleration has been set in the given source.
func AddTolerationIfNotFound(source []core.Toleration, toAdd core.Toleration) []core.Toleration {
	if len(source) == 0 {
		return []core.Toleration{
			toAdd,
		}
	}

	for _, t := range source {
		if (t.Key == toAdd.Key || len(t.Key) == 0) && (t.Effect == toAdd.Effect || len(t.Effect) == 0) {
			// Toleration alread exists, do not add
			return source
		}
	}
	return append(source, toAdd)
}
