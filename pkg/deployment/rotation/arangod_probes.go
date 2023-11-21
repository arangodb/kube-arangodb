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

package rotation

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
)

func compareServerContainerProbes(ds api.DeploymentSpec, g api.ServerGroup, spec, status *core.Container) compare.Func {
	return func(builder api.ActionBuilder) (mode compare.Mode, plan api.Plan, err error) {
		if !areProbesEqual(spec.StartupProbe, status.StartupProbe) {
			status.StartupProbe = spec.StartupProbe
			mode = mode.And(compare.SilentRotation)
		}

		if !areProbesEqual(spec.ReadinessProbe, status.ReadinessProbe) {
			if isManagedProbe(spec.ReadinessProbe, status.ReadinessProbe) {
				q := status.ReadinessProbe.DeepCopy()

				q.Exec = spec.ReadinessProbe.Exec.DeepCopy()

				if equality.Semantic.DeepDerivative(spec.ReadinessProbe, q) {
					status.ReadinessProbe = spec.ReadinessProbe
					mode = mode.And(compare.SilentRotation)
				}
			}
		}

		if !areProbesEqual(spec.LivenessProbe, status.LivenessProbe) {
			if isManagedProbe(spec.LivenessProbe, status.LivenessProbe) {
				if spec.LivenessProbe.FailureThreshold != status.LivenessProbe.FailureThreshold {
					status.LivenessProbe.FailureThreshold = spec.LivenessProbe.FailureThreshold
					mode = mode.And(compare.SilentRotation)
				}

				q := status.LivenessProbe.DeepCopy()

				q.Exec = spec.LivenessProbe.Exec.DeepCopy()

				if equality.Semantic.DeepDerivative(spec.LivenessProbe, q) {
					status.LivenessProbe = spec.LivenessProbe
					mode = mode.And(compare.SilentRotation)
				}
			}
		}

		return
	}
}
