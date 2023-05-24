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
	"testing"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func TestEnsurePod_ArangoDB_Features(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "DBserver POD with disabled foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
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
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", false)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupDBServers),
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
		{
			Name: "DBserver POD with enabled foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
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

				deployment.currentObject.Spec.Features = &api.DeploymentFeatures{
					FoxxQueues: util.NewType[bool](false),
				}

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
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", false)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupDBServers),
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
		{
			Name: "Coordinator POD with undefined foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Coordinators: api.MemberStatusList{
							firstCoordinatorStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.Coordinators[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForCoordinator(firstCoordinatorStatus.ID, false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", true)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupDBServers),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ReadinessProbe:  createTestReadinessProbe(httpProbe, false, ""),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" +
						firstCoordinatorStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						false, ""),
				},
			},
		},
		{
			Name: "Coordinator POD with disabled foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Coordinators: api.MemberStatusList{
							firstCoordinatorStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.Coordinators[0].IsInitialized = true

				deployment.currentObject.Spec.Features = &api.DeploymentFeatures{
					FoxxQueues: util.NewType[bool](false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForCoordinator(firstCoordinatorStatus.ID, false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", false)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupCoordinators),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ReadinessProbe:  createTestReadinessProbe(httpProbe, false, ""),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" +
						firstCoordinatorStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						false, ""),
				},
			},
		},
		{
			Name: "Coordinator POD with enabled foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Coordinators: api.MemberStatusList{
							firstCoordinatorStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.Coordinators[0].IsInitialized = true

				deployment.currentObject.Spec.Features = &api.DeploymentFeatures{
					FoxxQueues: util.NewType[bool](true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForCoordinator(firstCoordinatorStatus.ID, false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", true)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupCoordinators),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ReadinessProbe:  createTestReadinessProbe(httpProbe, false, ""),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" +
						firstCoordinatorStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						false, ""),
				},
			},
		},
		{
			Name: "Single POD with undefined foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Single: api.MemberStatusList{
							singleStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.Single[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupSingle, singleStatus)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName)
				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(httpProbe, false, "")
			},
			ExpectedEvent: "member single is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForSingleMode(false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", true)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupSingle),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSingleTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupSingleString + "-" +
						singleStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSingleString,
						false, ""),
				},
			},
		},
		{
			Name: "Single POD with disabled foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Single: api.MemberStatusList{
							singleStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.Single[0].IsInitialized = true

				deployment.currentObject.Spec.Features = &api.DeploymentFeatures{
					FoxxQueues: util.NewType[bool](false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSingle, singleStatus)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName)
				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(httpProbe, false, "")
			},
			ExpectedEvent: "member single is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForSingleMode(false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", false)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupSingle),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSingleTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupSingleString + "-" +
						singleStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSingleString,
						false, ""),
				},
			},
		},
		{
			Name: "Single POD with enabled foxx services",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Single: api.MemberStatusList{
							singleStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.currentObjectStatus.Members.Single[0].IsInitialized = true

				deployment.currentObject.Spec.Features = &api.DeploymentFeatures{
					FoxxQueues: util.NewType[bool](true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSingle, singleStatus)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName)
				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(httpProbe, false, "")
			},
			ExpectedEvent: "member single is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: createTestCommandForSingleMode(false, false, func() k8sutil.OptionPairs {
								args := k8sutil.NewOptionPair()

								args.Add("--foxx.queues", true)

								return args
							}),
							Ports:     createTestPorts(api.ServerGroupSingle),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSingleTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupSingleString + "-" +
						singleStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSingleString,
						false, ""),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}
