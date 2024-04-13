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

package rotation

import (
	"testing"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/compare"
)

func Test_ArangoD_PodSecurityContext(t *testing.T) {
	testCases := []TestCase{
		{
			name:   "With deployment",
			spec:   buildPodSpec(),
			status: buildPodSpec(),

			deploymentSpec: buildDeployment(func(depl *api.DeploymentSpec) {

			}),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.SkippedRotation,
			},
		},
		{
			name:   "Nil to Nil SecurityContext",
			spec:   buildPodSpec(),
			status: buildPodSpec(),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.SkippedRotation,
			},
		},
		{
			name:   "Nil to Empty SecurityContext",
			spec:   buildPodSpec(addPodSecurityContext(nil)),
			status: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{})),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.SilentRotation,
			},
		},
		{
			name:   "Empty to nil SecurityContext",
			spec:   buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{})),
			status: buildPodSpec(addPodSecurityContext(nil)),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.SilentRotation,
			},
		},
		{
			name:   "Empty to Empty SecurityContext",
			spec:   buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{})),
			status: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{})),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.SkippedRotation,
			},
		},
		{
			name: "Empty to NonEmpty SecurityContext",
			spec: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{})),
			status: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1000),
			})),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.GracefulRotation,
			},
		},
		{
			name: "Nil to NonEmpty SecurityContext",
			spec: buildPodSpec(addPodSecurityContext(nil)),
			status: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1000),
			})),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.GracefulRotation,
			},
		},
		{
			name: "NonEmpty to Nil SecurityContext",
			spec: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1000),
			})),
			status: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{})),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.GracefulRotation,
			},
		},
		{
			name: "NonEmpty to Nil SecurityContext",
			spec: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1000),
			})),
			status: buildPodSpec(addPodSecurityContext(nil)),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.GracefulRotation,
			},
		},
		{
			name: "NonEmpty to NonEmpty SecurityContext",
			spec: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1000),
			})),
			status: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1000),
			})),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.SkippedRotation,
			},
		},
		{
			name: "NonEmpty to NonEmpty Changed SecurityContext",
			spec: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1000),
			})),
			status: buildPodSpec(addPodSecurityContext(&core.PodSecurityContext{
				RunAsGroup: util.NewType[int64](1001),
			})),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: compare.GracefulRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}
