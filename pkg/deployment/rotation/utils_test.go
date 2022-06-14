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
	"testing"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
)

type TestCase struct {
	name         string
	spec, status *core.PodTemplateSpec

	deploymentSpec api.DeploymentSpec
	expectedMode   Mode
	expectedPlan   api.Plan
	expectedErr    string
}

func runTestCases(t *testing.T) func(tcs ...TestCase) {
	return func(tcs ...TestCase) {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {

				pspec := newTemplateFromSpec(t, tc.spec, api.ServerGroupAgents, tc.deploymentSpec)
				pstatus := newTemplateFromSpec(t, tc.status, api.ServerGroupAgents, tc.deploymentSpec)

				mode, plan, err := compare(tc.deploymentSpec, api.MemberStatus{ID: "id"}, api.ServerGroupAgents, pspec, pstatus)

				if tc.expectedErr != "" {
					require.Error(t, err)
					require.EqualError(t, err, tc.expectedErr)
				} else {
					require.Equal(t, tc.expectedMode, mode)

					switch mode {
					case InPlaceRotation:
						require.Len(t, plan, len(tc.expectedPlan))

						for i := range plan {
							require.Equal(t, tc.expectedPlan[i].Type, plan[i].Type)
						}
					}
				}
			})
		}
	}
}

func newTemplateFromSpec(t *testing.T, podSpec *core.PodTemplateSpec, group api.ServerGroup, deploymentSpec api.DeploymentSpec) *api.ArangoMemberPodTemplate {
	checksum, err := resources.ChecksumArangoPod(deploymentSpec.GetServerGroupSpec(group), resources.CreatePodFromTemplate(podSpec))
	require.NoError(t, err)

	newSpec, err := api.GetArangoMemberPodTemplate(podSpec, checksum)
	require.NoError(t, err)

	return newSpec
}

type podSpecBuilder func(pod *core.PodTemplateSpec)

func buildPodSpec(b ...podSpecBuilder) *core.PodTemplateSpec {
	p := &core.PodTemplateSpec{}

	for _, i := range b {
		i(p)
	}

	return p
}

func addContainer(name string, f func(c *core.Container)) podSpecBuilder {
	return func(pod *core.PodTemplateSpec) {
		var c core.Container

		c.Name = name

		if f != nil {
			f(&c)
		}

		pod.Spec.Containers = append(pod.Spec.Containers, c)
	}
}

func addInitContainer(name string, f func(c *core.Container)) podSpecBuilder {
	return func(pod *core.PodTemplateSpec) {
		var c core.Container

		c.Name = name

		if f != nil {
			f(&c)
		}

		pod.Spec.InitContainers = append(pod.Spec.InitContainers, c)
	}
}

func addSidecarWithImage(name, image string) podSpecBuilder {
	return addContainer(name, func(c *core.Container) {
		c.Image = image
	})
}

func addContainerWithCommand(name string, command []string) podSpecBuilder {
	return addContainer(name, func(c *core.Container) {
		c.Command = command
	})
}

type deploymentBuilder func(depl *api.DeploymentSpec)

func buildDeployment(b ...deploymentBuilder) api.DeploymentSpec {
	p := api.DeploymentSpec{}

	for _, i := range b {
		i(&p)
	}

	return p
}
