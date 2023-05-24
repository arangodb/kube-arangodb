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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func TestEnsurePod_ArangoDB_Core(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Agent Pod with image pull policy",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:           util.NewType[string](testImage),
					Authentication:  noAuthentication,
					TLS:             noTLS,
					ImagePullPolicy: util.NewType[core.PullPolicy](core.PullAlways),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}
				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullAlways,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with image pull policy",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:           util.NewType[string](testImage),
					Authentication:  noAuthentication,
					TLS:             noTLS,
					ImagePullPolicy: util.NewType[core.PullPolicy](core.PullAlways),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}
				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullAlways,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with sidecar",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Agents: api.ServerGroupSpec{
						Sidecars: []core.Container{
							{
								Name: sidecarName1,
							},
							{
								Name: sidecarName2,
							},
						},
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}
				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						{
							Name: sidecarName1,
						},
						{
							Name: sidecarName2,
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with image pull secrets",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:            util.NewType[string](testImage),
					Authentication:   noAuthentication,
					TLS:              noTLS,
					ImagePullSecrets: []string{"docker-registry", "other-registry"},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}
				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					ImagePullSecrets: []core.LocalObjectReference{
						{
							Name: "docker-registry",
						},
						{
							Name: "other-registry",
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with alpine init container",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			config: Config{
				OperatorImage: testImageOperator,
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					InitContainers: []core.Container{
						createTestAlpineContainer(firstAgentStatus.ID, false),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
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
							Ports:     createTestPorts(api.ServerGroupDBServers),
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
			Name: "DBserver POD with resource requirements and memory override",
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
							Ports:     createTestPorts(api.ServerGroupDBServers),
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
			Name: "DBserver POD with resource requirements and persistent volume claim",
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

				deployment.currentObjectStatus.Members.DBServers[0].PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: testPersistentVolumeClaimName,
				}
				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeWithPersitantVolumeClaim(shared.ArangodVolumeName,
							testPersistentVolumeClaimName),
					},
					Containers: []core.Container{
						{
							Name:      shared.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(api.ServerGroupDBServers),
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
			Name: "Initialized DBserver POD with alpine init container",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			config: Config{
				OperatorImage: testImageOperator,
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
					InitContainers: []core.Container{
						createTestAlpineContainer(firstDBServerStatus.ID, true),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupDBServers),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
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
			Name: "Agent Pod without TLS, authentication, persistent volume claim, metrics, rocksDB encryption, license",
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
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with persistent volume claim",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				agentWithPersistentVolumeClaim := firstAgentStatus
				agentWithPersistentVolumeClaim.PersistentVolumeClaim = &api.MemberPersistentVolumeClaimStatus{
					Name: testPersistentVolumeClaimName,
				}

				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							agentWithPersistentVolumeClaim,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeWithPersitantVolumeClaim(shared.ArangodVolumeName,
							testPersistentVolumeClaimName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with TLS",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            tlsSpec,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupAgentsString, firstAgentStatus.ID),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, true, false, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, true, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with authentication and unsecured liveness",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS:            noTLS,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, authorization, shared.ServerPortName)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, true, false),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources:       emptyResources,
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						true, ""),
				},
			},
		},
		{
			Name: "Agent Pod with TLS and authentication and secured liveness probe",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS: api.TLSSpec{
						CASecretName: util.NewType[string](testCASecretName),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, true, authorization, shared.ServerPortName)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupAgentsString, firstAgentStatus.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            shared.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForAgent(firstAgentStatus.ID, true, true, false),
							Ports:           createTestPorts(api.ServerGroupAgents),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources: emptyResources,
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with encrypted rocksdb",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					RocksDB:        rocksDBSpec,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)

				key := make([]byte, 32)
				k8sutil.CreateEncryptionKeySecret(deployment.SecretsModInterface(), testRocksDBEncryptionKey, key)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(shared.RocksdbEncryptionVolumeName, testRocksDBEncryptionKey),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, true),
							Ports:   createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.RocksdbEncryptionVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultAgentTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupAgentsString + "-" + firstAgentStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupAgentsString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver Pod with internal metrics exporter",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS:            noTLS,
					Metrics:        metricsSpec,
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, authorization, shared.ServerPortName)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(shared.ExporterJWTVolumeName, testExporterToken),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, true, false),
							Ports:   createTestPorts(api.ServerGroupDBServers),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources:       emptyResources,
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testArangodbInternalExporterContainer(false, true, emptyResources),
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupDBServersString + "-" + firstDBServerStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver Pod with metrics exporter which contains resource requirements",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS:            noTLS,
					Metrics: api.MetricsSpec{
						Enabled: util.NewType[bool](true),
						Image:   util.NewType[string](testImage),
						Authentication: api.MetricsAuthenticationSpec{
							JWTTokenSecretName: util.NewType[string](testExporterToken),
						},
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, authorization, shared.ServerPortName)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(shared.ExporterJWTVolumeName, testExporterToken),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, true, false),
							Ports:   createTestPorts(api.ServerGroupDBServers),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources:       emptyResources,
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testArangodbInternalExporterContainer(false, true, k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered)),
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupDBServersString + "-" + firstDBServerStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver Pod with lifecycle init container which contains resource requirements",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS:            noTLS,
					Metrics:        metricsSpec,
					Lifecycle: api.LifecycleSpec{
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, authorization, shared.ServerPortName)
			},
			config: Config{
				OperatorImage: testImageOperator,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(shared.ExporterJWTVolumeName, testExporterToken),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered)),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, true, false),
							Ports:   createTestPorts(api.ServerGroupDBServers),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources:       emptyResources,
							Lifecycle:       createTestLifecycle(api.ServerGroupAgents),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testArangodbInternalExporterContainer(false, true, emptyResources),
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupDBServersString + "-" + firstDBServerStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "DBserver Pod with metrics exporter and lifecycle init container and alpine init container",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS:            noTLS,
					Metrics:        metricsSpec,
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, false, authorization, shared.ServerPortName)
			},
			config: Config{
				OperatorImage: testImageOperator,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(shared.ExporterJWTVolumeName, testExporterToken),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
						createTestAlpineContainer(firstDBServerStatus.ID, false),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, true, false),
							Ports:   createTestPorts(api.ServerGroupDBServers),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources:       emptyResources,
							Lifecycle:       createTestLifecycle(api.ServerGroupAgents),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testArangodbInternalExporterContainer(false, true, emptyResources),
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupDBServersString + "-" + firstDBServerStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "Coordinator Pod with TLS and authentication and readiness and liveness",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					TLS: api.TLSSpec{
						CASecretName: util.NewType[string](testCASecretName),
					},
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

				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)

				auth, err := createTestToken(deployment, testCase, []string{"/_admin/server/availability"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(httpProbe, true, auth)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            shared.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForCoordinator(firstCoordinatorStatus.ID, true, true),
							Ports:           createTestPorts(api.ServerGroupCoordinators),
							ImagePullPolicy: core.PullIfNotPresent,
							Resources:       emptyResources,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" + firstCoordinatorStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						true, ""),
				},
			},
		},
		{
			Name: "Single Pod with TLS and authentication and readiness and readiness",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS: api.TLSSpec{
						CASecretName: util.NewType[string](testCASecretName),
					},
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

				testCase.createTestPodData(deployment, api.ServerGroupSingle, singleStatus)

				authLiveness, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				authReadiness, err := createTestToken(deployment, testCase, []string{"/_admin/server/availability"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, true, authLiveness, shared.ServerPortName)
				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(httpProbe, true, authReadiness)
			},
			ExpectedEvent: "member single is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupSingleString, singleStatus.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            shared.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForSingleMode(true, true),
							Ports:           createTestPorts(api.ServerGroupSingle),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources: emptyResources,
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSingleTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupSingleString + "-" + singleStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSingleString,
						false, ""),
				},
			},
		},
		{
			Name: "Single Pod with TLS and authentication and readiness and port override",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS: api.TLSSpec{
						CASecretName: util.NewType[string](testCASecretName),
					},
					Single: api.ServerGroupSpec{
						Port: util.NewType[uint16](18529),
					},
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

				testCase.createTestPodData(deployment, api.ServerGroupSingle, singleStatus)

				authLiveness, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				authReadiness, err := createTestToken(deployment, testCase, []string{"/_admin/server/availability"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, true, authLiveness, shared.ServerPortName)
				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(httpProbe, true, authReadiness)
			},
			ExpectedEvent: "member single is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupSingleString, singleStatus.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            shared.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForSingleModeWithPortOverride(true, true, 18529),
							Ports:           createTestPorts(api.ServerGroupSingle, 18529),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources: emptyResources,
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSingleTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupSingleString + "-" + singleStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSingleString,
						false, ""),
				},
			},
		},
		{
			Name: "Coordinator Pod with TLS and authentication and readiness and liveness and custom port",
			Features: testCaseFeatures{
				JWTRotation: true,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					TLS: api.TLSSpec{
						CASecretName: util.NewType[string](testCASecretName),
					},
					Coordinators: api.ServerGroupSpec{
						Port: util.NewType[uint16](18529),
					},
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

				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)

				auth, err := createTestToken(deployment, testCase, []string{"/_admin/server/availability"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(cmdProbe, true, auth)
				testCase.ExpectedPod.Spec.Containers[0].StartupProbe = createTestStartupProbe(cmdProbe, true, auth, 720)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            shared.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForCoordinatorWithPort(firstCoordinatorStatus.ID, true, true, 18529),
							Ports:           createTestPorts(api.ServerGroupCoordinators, 18529),
							ImagePullPolicy: core.PullIfNotPresent,
							Resources:       emptyResources,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" + firstCoordinatorStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						true, ""),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}

func testArangodbInternalExporterContainer(secure, auth bool, resources core.ResourceRequirements, ports ...int32) core.Container {

	var port int32 = shared.ArangoExporterPort
	if len(ports) > 0 {
		port = ports[0]
	}

	c := core.Container{
		Name:    shared.ExporterContainerName,
		Image:   testImage,
		Command: createTestInternalExporterCommand(secure, auth, port),
		Ports: []core.ContainerPort{
			{
				Name:          string(api.MetricsModeExporter),
				ContainerPort: port,
				Protocol:      core.ProtocolTCP,
			},
		},
		LivenessProbe:   createTestExporterLivenessProbe(secure),
		Resources:       resources,
		ImagePullPolicy: core.PullIfNotPresent,
		SecurityContext: &core.SecurityContext{
			Capabilities: &core.Capabilities{
				Drop: []core.Capability{
					"ALL",
				},
			},
		},
		VolumeMounts: []core.VolumeMount{
			k8sutil.LifecycleVolumeMount(),
		},
	}

	if auth {
		c.VolumeMounts = append(c.VolumeMounts, k8sutil.ExporterJWTVolumeMount())
	}

	return c
}

func createTestInternalExporterCommand(secure, auth bool, port int32) []string {
	binaryPath, err := os.Executable()
	if err != nil {
		return []string{}
	}
	exePath := filepath.Join(k8sutil.LifecycleVolumeMountDir, filepath.Base(binaryPath))

	args := []string{exePath, "exporter"}
	if secure {
		args = append(args, "--arangodb.endpoint=https://localhost:8529/_admin/metrics")
	} else {
		args = append(args, "--arangodb.endpoint=http://localhost:8529/_admin/metrics")
	}

	if auth {
		args = append(args, "--arangodb.jwt-file=/secrets/exporter/jwt/token")
	}

	if port != shared.ArangoExporterPort {
		args = append(args, fmt.Sprintf("--server.address=:%d", port))
	}

	if secure {
		args = append(args, "--ssl.keyfile=/secrets/tls/tls.keyfile")
	}

	return args
}
