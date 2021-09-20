//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

	v1 "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func Test_ArangoDContainers_SidecarImages(t *testing.T) {
	testCases := []TestCase{
		{
			name:   "Sidecar Image Update",
			spec:   buildPodSpec(addContainer(k8sutil.ServerContainerName, nil), addSidecarWithImage("sidecar", "local:1.0")),
			status: buildPodSpec(addContainer(k8sutil.ServerContainerName, nil), addSidecarWithImage("sidecar", "local:2.0")),

			expectedMode: InPlaceRotation,
			expectedPlan: api.Plan{
				api.NewAction(api.ActionTypeRuntimeContainerImageUpdate, 0, ""),
			},
		},
		{
			name:   "Sidecar Image Update with more than one sidecar",
			spec:   buildPodSpec(addSidecarWithImage("sidecar1", "local:1.0"), addSidecarWithImage("sidecar", "local:1.0")),
			status: buildPodSpec(addSidecarWithImage("sidecar1", "local:1.0"), addSidecarWithImage("sidecar", "local:2.0")),

			expectedMode: InPlaceRotation,
			expectedPlan: api.Plan{
				api.NewAction(api.ActionTypeRuntimeContainerImageUpdate, 0, ""),
			},
		},
	}

	runTestCases(t)(testCases...)
}

func Test_InitContainers(t *testing.T) {
	t.Run("Ignore", func(t *testing.T) {
		testCases := []TestCase{
			{
				name: "Same containers",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, nil), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, nil), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:1.0"
				})),

				expectedMode: SkippedRotation,

				deploymentSpec: buildDeployment(func(depl *api.DeploymentSpec) {
					depl.Agents.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerIgnoreMode.New(),
					}
				}),
			},
			{
				name: "Containers with different image",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, nil), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, nil), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:2.0"
				})),

				expectedMode: SilentRotation,

				deploymentSpec: buildDeployment(func(depl *api.DeploymentSpec) {
					depl.Agents.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerIgnoreMode.New(),
					}
				}),
			},
		}

		runTestCases(t)(testCases...)
	})

	t.Run("update", func(t *testing.T) {
		testCases := []TestCase{
			{
				name: "Containers with different image but init rotation enforced",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, nil), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, nil), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:2.0"
				})),

				expectedMode: GracefulRotation,

				deploymentSpec: buildDeployment(func(depl *api.DeploymentSpec) {
					depl.Agents.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerUpdateMode.New(),
					}
				}),
			},
			{
				name: "Core Containers with different image but init rotation enforced",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *v1.Container) {
					c.Image = "local:1.0"
				}), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *v1.Container) {
					c.Image = "local:2.0"
				}), addInitContainer("sidecar", func(c *v1.Container) {
					c.Image = "local:1.0"
				})),

				expectedMode: SilentRotation,

				deploymentSpec: buildDeployment(func(depl *api.DeploymentSpec) {
					depl.Agents.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerUpdateMode.New(),
					}
				}),
			},
		}

		runTestCases(t)(testCases...)
	})
}

func Test_Container_Args(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Only log level arguments of the ArangoDB server have been changed",
			spec: buildPodSpec(addContainerWithCommand(k8sutil.ServerContainerName,
				[]string{"--log.level=INFO", "--log.level=requests=error"})),
			status:       buildPodSpec(addContainerWithCommand(k8sutil.ServerContainerName, []string{"--log.level=INFO"})),
			expectedMode: InPlaceRotation,
			expectedPlan: api.Plan{
				api.NewAction(api.ActionTypeRuntimeContainerArgsLogLevelUpdate, 0, ""),
			},
		},
		{
			name: "Only log level arguments of the Sidecar have been changed",
			spec: buildPodSpec(addContainerWithCommand("sidecar",
				[]string{"--log.level=INFO", "--log.level=requests=error"})),
			status:       buildPodSpec(addContainerWithCommand("sidecar", []string{"--log.level=INFO"})),
			expectedMode: GracefulRotation,
		},
		{
			name:   "ArangoDB server arguments have not been changed",
			spec:   buildPodSpec(addContainerWithCommand(k8sutil.ServerContainerName, []string{"--log.level=INFO"})),
			status: buildPodSpec(addContainerWithCommand(k8sutil.ServerContainerName, []string{"--log.level=INFO"})),
		},
		{
			name: "Not only log level arguments of the ArangoDB server have been changed",
			spec: buildPodSpec(addContainerWithCommand(k8sutil.ServerContainerName, []string{"--log.level=INFO",
				"--server.endpoint=localhost"})),
			status:       buildPodSpec(addContainerWithCommand(k8sutil.ServerContainerName, []string{"--log.level=INFO"})),
			expectedMode: GracefulRotation,
		},
	}

	runTestCases(t)(testCases...)
}
