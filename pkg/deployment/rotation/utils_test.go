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

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
)

type TestCaseOverride struct {
	expectedMode Mode
	expectedPlan api.Plan
	expectedErr  string
}

type TestCase struct {
	name         string
	spec, status *core.PodTemplateSpec

	deploymentSpec api.DeploymentSpec
	groupSpec      api.ServerGroupSpec

	TestCaseOverride

	overrides map[api.DeploymentMode]map[api.ServerGroup]TestCaseOverride
}

func runTestCases(t *testing.T) func(tcs ...TestCase) {

	return func(tcs ...TestCase) {
		for _, tc := range tcs {
			t.Run(tc.name, func(t *testing.T) {
				runTestCasesForMode(t, api.DeploymentModeSingle, tc)
				runTestCasesForMode(t, api.DeploymentModeActiveFailover, tc)
				runTestCasesForMode(t, api.DeploymentModeCluster, tc)
			})
		}
	}
}

func runTestCasesForMode(t *testing.T, m api.DeploymentMode, tc TestCase) {
	t.Run(m.String(), func(t *testing.T) {
		switch m {
		case api.DeploymentModeSingle:
			runTestCasesForModeAndGroup(t, m, api.ServerGroupSingle, tc)
		case api.DeploymentModeCluster:
			runTestCasesForModeAndGroup(t, m, api.ServerGroupAgents, tc)
			runTestCasesForModeAndGroup(t, m, api.ServerGroupDBServers, tc)
			runTestCasesForModeAndGroup(t, m, api.ServerGroupCoordinators, tc)
		case api.DeploymentModeActiveFailover:
			runTestCasesForModeAndGroup(t, m, api.ServerGroupAgents, tc)
			runTestCasesForModeAndGroup(t, m, api.ServerGroupSingle, tc)
		}
	})
}

func runTestCasesForModeAndGroup(t *testing.T, m api.DeploymentMode, g api.ServerGroup, tc TestCase) {
	t.Run(g.AsRole(), func(t *testing.T) {
		ds := tc.deploymentSpec.DeepCopy()
		if ds == nil {
			ds = &api.DeploymentSpec{}
		}

		ds.Mode = m.New()

		ds.UpdateServerGroupSpec(g, tc.groupSpec)

		if tc.spec == nil {
			tc.spec = buildPodSpec()
		}
		if tc.status == nil {
			tc.status = buildPodSpec()
		}

		pspec := newTemplateFromSpec(t, tc.spec, g, *ds)
		pstatus := newTemplateFromSpec(t, tc.status, g, *ds)

		mode, plan, err := compare(*ds, api.MemberStatus{ID: "id"}, g, pspec, pstatus)

		q := tc.TestCaseOverride

		if v, ok := tc.overrides[m][g]; ok {
			q = v
		}

		if tc.expectedErr != "" {
			require.Error(t, err)
			require.EqualError(t, err, q.expectedErr)
		} else {
			require.Equal(t, q.expectedMode, mode)

			switch mode {
			case InPlaceRotation:
				require.Len(t, plan, len(q.expectedPlan))

				for i := range plan {
					require.Equal(t, q.expectedPlan[i].Type, plan[i].Type)
				}
			}
		}
	})
}

func newTemplateFromSpec(t *testing.T, podSpec *core.PodTemplateSpec, group api.ServerGroup, deploymentSpec api.DeploymentSpec) *api.ArangoMemberPodTemplate {
	checksum, err := resources.ChecksumArangoPod(deploymentSpec.GetServerGroupSpec(group), resources.CreatePodFromTemplate(podSpec))
	require.NoError(t, err)

	newSpec, err := api.GetArangoMemberPodTemplate(podSpec, checksum)
	require.NoError(t, err)

	return newSpec
}

type podSpecBuilder func(pod *core.PodTemplateSpec)

type podContainerBuilder func(c *core.Container)

func buildPodSpec(b ...podSpecBuilder) *core.PodTemplateSpec {
	p := &core.PodTemplateSpec{}

	for _, i := range b {
		i(p)
	}

	return p
}

func addContainer(name string, f ...podContainerBuilder) podSpecBuilder {
	return func(pod *core.PodTemplateSpec) {
		var c core.Container

		c.Name = name

		for _, q := range f {
			q(&c)
		}

		pod.Spec.Containers = append(pod.Spec.Containers, c)
	}
}

func addInitContainer(name string, f ...podContainerBuilder) podSpecBuilder {
	return func(pod *core.PodTemplateSpec) {
		var c core.Container

		c.Name = name

		for _, q := range f {
			q(&c)
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

type groupSpecBuilder func(depl *api.ServerGroupSpec)

func buildGroupSpec(b ...groupSpecBuilder) api.ServerGroupSpec {
	p := api.ServerGroupSpec{}

	for _, i := range b {
		i(&p)
	}

	return p
}
