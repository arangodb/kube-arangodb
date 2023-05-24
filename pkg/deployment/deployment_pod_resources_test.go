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

package deployment

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type envFunc func() []core.EnvVar

func withEnvs(f ...envFunc) []core.EnvVar {
	var e []core.EnvVar

	for _, c := range f {
		e = append(e, c()...)
	}

	return e
}

func withDefaultEnvs(t *testing.T, requirements core.ResourceRequirements) []core.EnvVar {
	var q []envFunc

	q = append(q, resourceLimitAsEnv(t, requirements))

	return withEnvs(q...)
}

func resourceLimitAsEnv(t *testing.T, requirements core.ResourceRequirements) envFunc {
	return func() []core.EnvVar {
		var e []core.EnvVar
		if _, ok := requirements.Limits[core.ResourceMemory]; ok {
			e = append(e, resourceMemoryLimitAsEnv(t, requirements)()...)
		}
		if _, ok := requirements.Limits[core.ResourceCPU]; ok {
			e = append(e, resourceCPULimitAsEnv(t, requirements)()...)
		}
		return e
	}
}

func resourceMemoryLimitAsEnv(t *testing.T, requirements core.ResourceRequirements) envFunc {
	value, ok := requirements.Limits[core.ResourceMemory]
	require.True(t, ok)
	return func() []core.EnvVar {
		return []core.EnvVar{{
			Name:  resources.ArangoDBOverrideDetectedTotalMemoryEnv,
			Value: fmt.Sprintf("%d", value.Value()),
		},
		}
	}
}

func resourceCPULimitAsEnv(t *testing.T, requirements core.ResourceRequirements) envFunc {
	value, ok := requirements.Limits[core.ResourceCPU]
	require.True(t, ok)

	return func() []core.EnvVar {
		return []core.EnvVar{{
			Name:  resources.ArangoDBOverrideDetectedNumberOfCoresEnv,
			Value: fmt.Sprintf("%d", value.Value()),
		},
		}
	}
}

func TestEnsurePod_ArangoDB_Resources(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "DBserver POD with resource requirements",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Resources: resourcesUnfiltered,
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      shared.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(api.ServerGroupAgents),
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							Env:             withDefaultEnvs(t, resourcesUnfiltered),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver POD with resource requirements, with override flag",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Resources:                   resourcesUnfiltered,
						OverrideDetectedTotalMemory: util.NewType[bool](false),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      shared.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(api.ServerGroupAgents),
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Env:             withEnvs(resourceCPULimitAsEnv(t, resourcesUnfiltered)),
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver POD without resource requirements, with override flag",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						OverrideDetectedTotalMemory: util.NewType[bool](true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      shared.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(api.ServerGroupAgents),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}
