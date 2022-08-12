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

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

func Test_ArangoD_Volumes(t *testing.T) {
	testCases := []TestCase{
		{
			name:   "Empty volumes",
			spec:   buildPodSpec(),
			status: buildPodSpec(),

			deploymentSpec: buildDeployment(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
		{
			name:   "Same volumes",
			spec:   buildPodSpec(addVolume("data", addVolumeConfigMapSource(&core.ConfigMapVolumeSource{}))),
			status: buildPodSpec(addVolume("data", addVolumeConfigMapSource(&core.ConfigMapVolumeSource{}))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
		{
			name: "Different volumes",
			spec: buildPodSpec(addVolume("data", addVolumeConfigMapSource(&core.ConfigMapVolumeSource{}))),
			status: buildPodSpec(addVolume("data", addVolumeConfigMapSource(&core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: "test",
				},
			}))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name:   "Missing volumes",
			spec:   buildPodSpec(addVolume("data", addVolumeConfigMapSource(&core.ConfigMapVolumeSource{}))),
			status: buildPodSpec(),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name: "Added volumes",
			spec: buildPodSpec(),
			status: buildPodSpec(addVolume("data", addVolumeConfigMapSource(&core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: "test",
				},
			}))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name: "Timezone: Different volumes",
			spec: buildPodSpec(addVolume(shared.ArangoDTimezoneVolumeName, addVolumeConfigMapSource(&core.ConfigMapVolumeSource{}))),
			status: buildPodSpec(addVolume(shared.ArangoDTimezoneVolumeName, addVolumeConfigMapSource(&core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: "test",
				},
			}))),

			overrides: map[api.DeploymentMode]map[api.ServerGroup]TestCaseOverride{
				api.DeploymentModeSingle: {
					api.ServerGroupSingle: {
						expectedMode: GracefulRotation,
					},
				},
				api.DeploymentModeActiveFailover: {
					api.ServerGroupSingle: {
						expectedMode: GracefulRotation,
					},
				},
				api.DeploymentModeCluster: {
					api.ServerGroupCoordinators: {
						expectedMode: GracefulRotation,
					},
				},
			},

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name:   "Timezone: Missing volumes",
			spec:   buildPodSpec(addVolume(shared.ArangoDTimezoneVolumeName, addVolumeConfigMapSource(&core.ConfigMapVolumeSource{}))),
			status: buildPodSpec(),

			overrides: map[api.DeploymentMode]map[api.ServerGroup]TestCaseOverride{
				api.DeploymentModeSingle: {
					api.ServerGroupSingle: {
						expectedMode: GracefulRotation,
					},
				},
				api.DeploymentModeActiveFailover: {
					api.ServerGroupSingle: {
						expectedMode: GracefulRotation,
					},
				},
				api.DeploymentModeCluster: {
					api.ServerGroupCoordinators: {
						expectedMode: GracefulRotation,
					},
				},
			},

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
		{
			name: "Timezone: Added volumes",
			spec: buildPodSpec(),
			status: buildPodSpec(addVolume(shared.ArangoDTimezoneVolumeName, addVolumeConfigMapSource(&core.ConfigMapVolumeSource{
				LocalObjectReference: core.LocalObjectReference{
					Name: "test",
				},
			}))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}

func Test_ArangoD_VolumeMounts(t *testing.T) {
	testCases := []TestCase{
		{
			name:   "Empty volume mounts",
			spec:   buildPodSpec(addContainer("server")),
			status: buildPodSpec(addContainer("server")),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
		{
			name: "Same volumes",
			spec: buildPodSpec(addContainer("server", addVolumeMount("mount", func(in *core.VolumeMount) {

			}))),
			status: buildPodSpec(addContainer("server", addVolumeMount("mount", func(in *core.VolumeMount) {

			}))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: SkippedRotation,
			},
		},
		{
			name:   "Different volumes",
			spec:   buildPodSpec(addContainer("server", addVolumeMount("mount"))),
			status: buildPodSpec(addContainer("server", addVolumeMount("mount2"))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name:   "Missing volumes",
			spec:   buildPodSpec(addContainer("server", addVolumeMount("mount"))),
			status: buildPodSpec(addContainer("server")),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name:   "Added volumes",
			spec:   buildPodSpec(addContainer("server")),
			status: buildPodSpec(addContainer("server", addVolumeMount("mount"))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name:   "Timezone: Different volumes",
			spec:   buildPodSpec(addContainer("server", addVolumeMount(shared.ArangoDTimezoneVolumeName))),
			status: buildPodSpec(addContainer("server", addVolumeMount("mount"))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name:   "Timezone: Missing volumes",
			spec:   buildPodSpec(addContainer("server")),
			status: buildPodSpec(addContainer("server", addVolumeMount(shared.ArangoDTimezoneVolumeName))),

			TestCaseOverride: TestCaseOverride{
				expectedMode: GracefulRotation,
			},
		},
		{
			name:   "Timezone: Added volumes",
			spec:   buildPodSpec(addContainer("server", addVolumeMount(shared.ArangoDTimezoneVolumeName))),
			status: buildPodSpec(addContainer("server")),

			overrides: map[api.DeploymentMode]map[api.ServerGroup]TestCaseOverride{
				api.DeploymentModeSingle: {
					api.ServerGroupSingle: {
						expectedMode: GracefulRotation,
					},
				},
				api.DeploymentModeActiveFailover: {
					api.ServerGroupSingle: {
						expectedMode: GracefulRotation,
					},
				},
				api.DeploymentModeCluster: {
					api.ServerGroupCoordinators: {
						expectedMode: GracefulRotation,
					},
				},
			},

			TestCaseOverride: TestCaseOverride{
				expectedMode: SilentRotation,
			},
		},
	}

	runTestCases(t)(testCases...)
}
