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

package utils

import (
	"time"

	core "k8s.io/api/core/v1"
)

func IsNodeSchedulableForPod(node *core.Node, pod *core.Pod) bool {
	return AreTaintsTolerated(pod.Spec.Tolerations, node.Spec.Taints)
}

func AreTaintsTolerated(tolerations []core.Toleration, taints []core.Taint) bool {
	for _, taint := range taints {
		if !IsTaintTolerated(tolerations, taint) {
			return false
		}
	}

	return true
}

func IsTaintTolerated(tolerations []core.Toleration, taint core.Taint) bool {
	for _, toleration := range tolerations {
		if toleration.Effect != "" && toleration.Effect != taint.Effect {
			// Not same effect
			continue
		}

		if toleration.Key != "" && toleration.Key != taint.Key {
			// Not same toleration key
			continue
		}

		switch toleration.Operator {
		case core.TolerationOpExists:
		// We accept all values
		case core.TolerationOpEqual:
			if toleration.Value != taint.Value {
				// If value does not match check next one
				continue
			}
		}

		if ts := toleration.TolerationSeconds; ts != nil {
			if taint.Effect == core.TaintEffectNoExecute {
				// NoExecute taint cant be tolerated for period of time
				continue
			}

			if s := taint.TimeAdded; s != nil {
				if start := s.Time; !start.IsZero() {
					since := time.Since(start)

					if since > time.Duration(*ts)*time.Second {
						// We tolerate particular duration for short period of time
						continue
					}
				}
			}
		}

		return true
	}

	return false
}
