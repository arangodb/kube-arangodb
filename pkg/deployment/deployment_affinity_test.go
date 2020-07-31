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
// Author Adam Janikowski
//

package deployment

import (
	"testing"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

func modifyAffinity(name, group string, required bool, role string, mods ...func(a *core.Affinity)) *core.Affinity {
	affinity := k8sutil.CreateAffinity(name, group,
		required, role)

	for _, mod := range mods {
		mod(affinity)
	}

	return affinity
}

func TestEnsurePod_ArangoDB_AntiAffinity(t *testing.T) {
	testAffinity := core.PodAffinityTerm{
		TopologyKey: "myTopologyKey",
	}

	weight := core.WeightedPodAffinityTerm{
		Weight:          6,
		PodAffinityTerm: testAffinity,
	}

	testCases := []testCaseStruct{
		{
			Name: "DBserver POD with antiAffinity required",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						AntiAffinity: &core.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
								testAffinity,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, testAffinity)
						}),
				},
			},
		},
		{
			Name: "DBserver POD with antiAffinity prefered",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						AntiAffinity: &core.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
								weight,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, weight)
						}),
				},
			},
		},
		{
			Name: "DBserver POD with antiAffinity both",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						AntiAffinity: &core.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
								weight,
							},
							RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
								testAffinity,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, weight)
							a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, testAffinity)
						}),
				},
			},
		},
		{
			Name: "DBserver POD with antiAffinity mixed",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						AntiAffinity: &core.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
								weight,
								weight,
								weight,
								weight,
							},
							RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
								testAffinity,
								testAffinity,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, weight, weight, weight, weight)
							a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, testAffinity, testAffinity)
						}),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}

func TestEnsurePod_ArangoDB_Affinity(t *testing.T) {
	testAffinity := core.PodAffinityTerm{
		TopologyKey: "myTopologyKey",
	}

	weight := core.WeightedPodAffinityTerm{
		Weight:          6,
		PodAffinityTerm: testAffinity,
	}

	testCases := []testCaseStruct{
		{
			Name: "DBserver POD with affinity required",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Affinity: &core.PodAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
								testAffinity,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							if a.PodAffinity == nil {
								a.PodAffinity = &core.PodAffinity{}
							}
							a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, testAffinity)
						}),
				},
			},
		},
		{
			Name: "DBserver POD with affinity prefered",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Affinity: &core.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
								weight,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							if a.PodAffinity == nil {
								a.PodAffinity = &core.PodAffinity{}
							}
							a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, weight)
						}),
				},
			},
		},
		{
			Name: "DBserver POD with affinity both",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Affinity: &core.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
								weight,
							},
							RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
								testAffinity,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							if a.PodAffinity == nil {
								a.PodAffinity = &core.PodAffinity{}
							}
							a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, weight)
							a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, testAffinity)
						}),
				},
			},
		},
		{
			Name: "DBserver POD with affinity mixed",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Affinity: &core.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []core.WeightedPodAffinityTerm{
								weight,
								weight,
								weight,
								weight,
							},
							RequiredDuringSchedulingIgnoredDuringExecution: []core.PodAffinityTerm{
								testAffinity,
								testAffinity,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							if a.PodAffinity == nil {
								a.PodAffinity = &core.PodAffinity{}
							}
							a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, weight, weight, weight, weight)
							a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(a.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, testAffinity, testAffinity)
						}),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}

func TestEnsurePod_ArangoDB_NodeAffinity(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "DBserver POD with nodeAffinity required",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						NodeAffinity: &core.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &core.NodeSelector{
								NodeSelectorTerms: []core.NodeSelectorTerm{
									{
										MatchFields: []core.NodeSelectorRequirement{
											{
												Key: "key",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImages(false),
				}
				deployment.status.last.Members.DBServers[0].IsInitialized = true

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: emptyResources,
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupDBServersString + "-" +
						firstDBServerStatus.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: modifyAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, "", func(a *core.Affinity) {
							f := a.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0]

							f.MatchFields = []core.NodeSelectorRequirement{
								{
									Key: "key",
								},
							}

							a.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0] = f
						}),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}
