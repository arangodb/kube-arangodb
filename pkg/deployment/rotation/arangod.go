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

package rotation

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func podCompare(_ api.DeploymentSpec, _ api.ServerGroup, spec, status *core.PodSpec) comparePodFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error) {
		if spec.SchedulerName != status.SchedulerName {
			status.SchedulerName = spec.SchedulerName
			mode = mode.And(SilentRotation)
		}

		if !util.CompareInt64p(spec.TerminationGracePeriodSeconds, status.TerminationGracePeriodSeconds) {
			status.TerminationGracePeriodSeconds = spec.TerminationGracePeriodSeconds
			mode = mode.And(SilentRotation)
		}

		return
	}
}

func affinityCompare(_ api.DeploymentSpec, _ api.ServerGroup, spec, status *core.PodSpec) comparePodFunc {
	return func(builder api.ActionBuilder) (mode Mode, plan api.Plan, e error) {
		if specC, err := util.SHA256FromJSON(spec.Affinity); err != nil {
			e = err
			return
		} else {
			if statusC, err := util.SHA256FromJSON(status.Affinity); err != nil {
				e = err
				return
			} else if specC != statusC {
				mode = mode.And(SilentRotation)
				status.Affinity = spec.Affinity.DeepCopy()
				return
			} else {
				return
			}
		}
	}
}
