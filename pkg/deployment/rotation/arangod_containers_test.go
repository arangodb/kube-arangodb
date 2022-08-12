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

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
)

func Test_ArangoDContainers_SidecarImages(t *testing.T) {
	testCases := []TestCase{
		{
			name:   "Sidecar Image Update",
			spec:   buildPodSpec(addContainer(shared.ServerContainerName), addSidecarWithImage("sidecar", "local:1.0")),
			status: buildPodSpec(addContainer(shared.ServerContainerName), addSidecarWithImage("sidecar", "local:2.0")),

			TestCaseOverride: TestCaseOverride{
				expectedMode: InPlaceRotation,
				expectedPlan: api.Plan{
					actions.NewClusterAction(api.ActionTypeRuntimeContainerImageUpdate),
				},
			},
		},
		{
			name:   "Sidecar Image Update with more than one sidecar",
			spec:   buildPodSpec(addSidecarWithImage("sidecar1", "local:1.0"), addSidecarWithImage("sidecar", "local:1.0")),
			status: buildPodSpec(addSidecarWithImage("sidecar1", "local:1.0"), addSidecarWithImage("sidecar", "local:2.0")),

			TestCaseOverride: TestCaseOverride{
				expectedMode: InPlaceRotation,
				expectedPlan: api.Plan{
					actions.NewClusterAction(api.ActionTypeRuntimeContainerImageUpdate),
				},
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
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),

				TestCaseOverride: TestCaseOverride{
					expectedMode: SkippedRotation,
				},

				groupSpec: buildGroupSpec(func(depl *api.ServerGroupSpec) {
					depl.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerIgnoreMode.New(),
					}
				}),
			},
			{
				name: "Containers with different image",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:2.0"
				})),

				TestCaseOverride: TestCaseOverride{
					expectedMode: SilentRotation,
				},

				groupSpec: buildGroupSpec(func(depl *api.ServerGroupSpec) {
					depl.InitContainers = &api.ServerGroupInitContainers{
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
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:2.0"
				})),

				TestCaseOverride: TestCaseOverride{
					expectedMode: GracefulRotation,
				},

				groupSpec: buildGroupSpec(func(depl *api.ServerGroupSpec) {
					depl.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerUpdateMode.New(),
					}
				}),
			},
			{
				name: "Core Containers with different image but init rotation enforced",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:1.0"
				}), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:2.0"
				}), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),

				TestCaseOverride: TestCaseOverride{
					expectedMode: SilentRotation,
				},

				groupSpec: buildGroupSpec(func(depl *api.ServerGroupSpec) {
					depl.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerUpdateMode.New(),
					}
				}),
			},
			{
				name: "Only core container change",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:1.0"
				}), addInitContainer(api.ServerGroupReservedInitContainerNameUpgrade, func(c *core.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:2.0"
				})),

				TestCaseOverride: TestCaseOverride{
					expectedMode: SilentRotation,
				},

				groupSpec: buildGroupSpec(func(depl *api.ServerGroupSpec) {
					depl.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerUpdateMode.New(),
					}
				}),
			},
			{
				name: "Only core container change with sidecar",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:1.0"
				}), addInitContainer(api.ServerGroupReservedInitContainerNameUpgrade, func(c *core.Container) {
					c.Image = "local:1.0"
				}), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:2.0"
				}), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),

				TestCaseOverride: TestCaseOverride{
					expectedMode: SilentRotation,
				},

				groupSpec: buildGroupSpec(func(depl *api.ServerGroupSpec) {
					depl.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerUpdateMode.New(),
					}
				}),
			},
			{
				name: "Only core container change with sidecar change",
				spec: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:1.0"
				}), addInitContainer(api.ServerGroupReservedInitContainerNameUpgrade, func(c *core.Container) {
					c.Image = "local:1.0"
				}), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:1.0"
				})),
				status: buildPodSpec(addInitContainer(api.ServerGroupReservedInitContainerNameUUID, func(c *core.Container) {
					c.Image = "local:2.0"
				}), addInitContainer("sidecar", func(c *core.Container) {
					c.Image = "local:2.0"
				})),

				TestCaseOverride: TestCaseOverride{
					expectedMode: GracefulRotation,
				},

				groupSpec: buildGroupSpec(func(depl *api.ServerGroupSpec) {
					depl.InitContainers = &api.ServerGroupInitContainers{
						Mode: api.ServerGroupInitContainerUpdateMode.New(),
					}
				}),
			},
		}

		runTestCases(t)(testCases...)
	})
}

func logLevelTestCaseGen(name string, mode Mode, spec, status []string) TestCase {
	var c TestCase
	c.name = name
	c.expectedMode = mode
	if c.expectedMode == InPlaceRotation {
		c.expectedPlan = api.Plan{
			actions.NewClusterAction(api.ActionTypeRuntimeContainerArgsLogLevelUpdate),
		}
	}
	c.spec = buildPodSpec(addContainerWithCommand(shared.ServerContainerName, spec))
	c.status = buildPodSpec(addContainerWithCommand(shared.ServerContainerName, status))

	return c
}

func Test_Container_LogArgs(t *testing.T) {
	testCases := []TestCase{
		logLevelTestCaseGen("Only log level arguments of the ArangoDB server have been changed",
			InPlaceRotation,
			[]string{"--log.level=INFO", "--log.level=requests=error"},
			[]string{"--log.level=INFO"}),
		logLevelTestCaseGen("ArangoDB server arguments have not been changed",
			SkippedRotation,
			[]string{"--log.level=INFO"},
			[]string{"--log.level=INFO"}),
		logLevelTestCaseGen("Multi ArangoDB server arguments have not been changed",
			SkippedRotation,
			[]string{"--log.level=INFO", "--log.level=requests=debug"},
			[]string{"--log.level=INFO", "--log.level=requests=debug"}),
		logLevelTestCaseGen("Not only log argument changed",
			GracefulRotation,
			[]string{"--log.level=INFO", "--server.endpoint=localhost"},
			[]string{"--log.level=INFO"}),
		logLevelTestCaseGen("Change of order with existing arg & switch to DEBUG",
			InPlaceRotation,
			[]string{"--log.level=DEBUG", "--foo"},
			[]string{"--foo", "--log.level=INFO"}),
		logLevelTestCaseGen("Removal of arg",
			InPlaceRotation,
			[]string{"--foo", "--log.level=INFO"},
			[]string{"--foo"}),
	}

	runTestCases(t)(testCases...)
}

func Test_Container_Args(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Only log level arguments of the Sidecar have been changed",
			spec: buildPodSpec(addContainerWithCommand("sidecar",
				[]string{"--log.level=INFO", "--log.level=requests=error"})),
			status: buildPodSpec(addContainerWithCommand("sidecar", []string{"--log.level=INFO"})),
			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}

func Test_Container_Ports(t *testing.T) {
	testCases := []TestCase{
		{
			name: "Ports of server pod changed",
			spec: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Ports = []core.ContainerPort{
					{
						Name:          "server",
						ContainerPort: 8526,
						Protocol:      "tcp",
					},
				}
			})),
			status: buildPodSpec(addContainer("server", func(c *core.Container) {
				c.Ports = []core.ContainerPort{
					{
						Name:          "server",
						ContainerPort: 8626,
						Protocol:      "tcp",
					},
				}
			})),
			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Ports of sidecar pod changed",
			spec: buildPodSpec(addContainer("sidecar", func(c *core.Container) {
				c.Ports = []core.ContainerPort{
					{
						Name:          "server",
						ContainerPort: 8526,
						Protocol:      "tcp",
					},
				}
			})),
			status: buildPodSpec(addContainer("sidecar", func(c *core.Container) {
				c.Ports = []core.ContainerPort{
					{
						Name:          "server",
						ContainerPort: 8626,
						Protocol:      "tcp",
					},
				}
			})),
			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}

func TestIsOnlyLogLevelChanged(t *testing.T) {
	type args struct {
		specArgs   []string
		statusArgs []string
	}
	tests := map[string]struct {
		args args
		want bool
	}{
		"log level not changed": {
			args: args{
				specArgs:   []string{"--log.level=INFO"},
				statusArgs: []string{"--log.level=INFO"},
			},
		},
		"log level changed": {
			args: args{
				specArgs:   []string{"--log.level=INFO", "--log.level=requests=DEBUG"},
				statusArgs: []string{"--log.level=INFO"},
			},
			want: true,
		},
		"log level and server endpoint changed": {
			args: args{
				specArgs:   []string{"--log.level=INFO", "--log.level=requests=DEBUG", "--server.endpoint=localhost"},
				statusArgs: []string{"--log.level=INFO"},
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got := isOnlyLogLevelChanged(testCase.args.specArgs, testCase.args.statusArgs)

			assert.Equal(t, testCase.want, got)
		})
	}
}
