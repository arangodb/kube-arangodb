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

package scheduler

import (
	core "k8s.io/api/core/v1"

	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
)

func SpecAsTemplate(in *pbSchedulerV1.Spec) *core.PodTemplateSpec {
	var ret core.PodTemplateSpec
	for _, c := range in.Containers {
		if c == nil {
			continue
		}

		var container core.Container

		if image := c.Image; image != nil {
			container.Image = *image
		}

		if len(c.Args) > 0 {
			container.Args = c.Args
		}

		for k, v := range c.EnvironmentVariables {
			container.Env = append(container.Env, core.EnvVar{
				Name:  k,
				Value: v,
			})
		}

		ret.Spec.Containers = append(ret.Spec.Containers, container)
	}

	if base := in.Base; base != nil {
		ret.ObjectMeta.Labels = base.Labels
	}

	ret.Spec.RestartPolicy = core.RestartPolicyNever

	return &ret
}
