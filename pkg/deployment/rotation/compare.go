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
	"encoding/json"

	jd "github.com/josephburnett/jd/lib"
	"github.com/rs/zerolog/log"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
)

type compareFuncGen[T any] func(deploymentSpec api.DeploymentSpec, group api.ServerGroup, spec, status *T) compareFunc
type compareFunc func(builder api.ActionBuilder) (mode Mode, plan api.Plan, err error)

// works for any type like: core.PodSpec, core.Container, core.InitContainer, core.DeploymentSpec, core.StatefulSetSpec
func genericFuncGenerator[T any](deploymentSpec api.DeploymentSpec, group api.ServerGroup, spec, status *T) func(c compareFuncGen[T]) compareFunc {
	return func(c compareFuncGen[T]) compareFunc {
		return c(deploymentSpec, group, spec, status)
	}
}

func compareGeneric(builder api.ActionBuilder, f ...compareFunc) (mode Mode, plan api.Plan, err error) {
	for _, q := range f {
		if m, p, err := q(builder); err != nil {
			return 0, nil, err
		} else {
			mode = mode.And(m)
			plan = append(plan, p...)
		}
	}

	return
}

func compare(deploymentSpec api.DeploymentSpec, member api.MemberStatus, group api.ServerGroup,
	spec, status *api.ArangoMemberPodTemplate) (mode Mode, plan api.Plan, err error) {

	if spec.Checksum == status.Checksum {
		return SkippedRotation, nil, nil
	}

	// If checksums are different and rotation is not needed and there are no changes between containers
	// then silent rotation must be applied to adjust status checksum.
	mode = SilentRotation

	podStatus := status.PodSpec.DeepCopy()

	// Try to fill fields
	b := actions.NewActionBuilderWrap(group, member)

	g := genericFuncGenerator(deploymentSpec, group, &spec.PodSpec.Spec, &podStatus.Spec)

	if m, p, err := compareGeneric(b, g(podCompare), g(affinityCompare), g(comparePodVolumes), g(containersCompare), g(initContainersCompare), g(comparePodTolerations)); err != nil {
		log.Err(err).Msg("Error while getting pod diff")
		return SkippedRotation, nil, err
	} else {
		mode = mode.And(m)
		plan = append(plan, p...)
	}

	checksum, err := resources.ChecksumArangoPod(deploymentSpec.GetServerGroupSpec(group), resources.CreatePodFromTemplate(podStatus))
	if err != nil {
		log.Err(err).Msg("Error while getting pod checksum")
		return SkippedRotation, nil, err
	}

	newStatus, err := api.GetArangoMemberPodTemplate(podStatus, checksum)
	if err != nil {
		log.Err(err).Msg("Error while getting template")
		return SkippedRotation, nil, err
	}

	if spec.RotationNeeded(newStatus) {
		line := logger.Str("id", member.ID)

		specBytes, errA := json.Marshal(spec.PodSpec)
		if errA == nil {
			line = line.Str("spec", string(specBytes))
		}

		statusBytes, errB := json.Marshal(newStatus.PodSpec)
		if errB == nil {
			line = line.Str("status", string(statusBytes))
		}

		if errA == nil && errB == nil {
			if specData, err := jd.ReadJsonString(string(specBytes)); err == nil && specData != nil {
				if statusData, err := jd.ReadJsonString(string(statusBytes)); err == nil && statusData != nil {
					line = line.Str("diff", specData.Diff(statusData).Render())
				}
			}
		}

		line.Info("Pod needs rotation - templates does not match")

		return GracefulRotation, nil, nil
	}

	return
}
