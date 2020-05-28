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
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package deployment

import (
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	core "k8s.io/api/core/v1"
)

func TestEnsurePod_ArangoDB_Core(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Agent Pod with image pull policy",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:           util.NewString(testImage),
					Authentication:  noAuthentication,
					TLS:             noTLS,
					ImagePullPolicy: util.NewPullPolicy(core.PullAlways),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:           util.NewString(testImage),
					Authentication:  noAuthentication,
					TLS:             noTLS,
					ImagePullPolicy: util.NewPullPolicy(core.PullAlways),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:          util.NewString(testImage),
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
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:            util.NewString(testImage),
					Authentication:   noAuthentication,
					TLS:              noTLS,
					ImagePullSecrets: []string{"docker-registry", "other-registry"},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			config: Config{
				OperatorUUIDInitImage: testImageOperatorUUIDInit,
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					InitContainers: []core.Container{
						createTestAlpineContainer(firstAgentStatus.ID, false),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Resources: resourcesUnfiltered,
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
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
			Name: "DBserver POD with resource requirements and memory override",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Resources: resourcesUnfiltered,
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
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
			Name: "DBserver POD with resource requirements and persistent volume claim",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					DBServers: api.ServerGroupSpec{
						Resources: resourcesUnfiltered,
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

				deployment.status.last.Members.DBServers[0].PersistentVolumeClaimName = testPersistentVolumeClaimName
				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeWithPersitantVolumeClaim(k8sutil.ArangodVolumeName,
							testPersistentVolumeClaimName),
					},
					Containers: []core.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
			Name: "Initialized DBserver POD with alpine init container",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			config: Config{
				OperatorUUIDInitImage: testImageOperatorUUIDInit,
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
					InitContainers: []core.Container{
						createTestAlpineContainer(firstDBServerStatus.ID, true),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				agentWithPersistentVolumeClaim := firstAgentStatus
				agentWithPersistentVolumeClaim.PersistentVolumeClaimName = testPersistentVolumeClaimName

				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeWithPersitantVolumeClaim(k8sutil.ArangodVolumeName,
							testPersistentVolumeClaimName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            tlsSpec,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupAgentsString, firstAgentStatus.ID),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, true, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(true, "", k8sutil.ArangoPort),
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
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					TLS:            noTLS,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(false,
					authorization, k8sutil.ArangoPort)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, true, false),
							Ports:   createTestPorts(),
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
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					TLS: api.TLSSpec{
						CASecretName: util.NewString(testCASecretName),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(true,
					authorization, k8sutil.ArangoPort)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupAgentsString, firstAgentStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            k8sutil.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForAgent(firstAgentStatus.ID, true, true, false),
							Ports:           createTestPorts(),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					RocksDB:        rocksDBSpec,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImages(false),
				}

				testCase.createTestPodData(deployment, api.ServerGroupAgents, firstAgentStatus)

				secrets := deployment.GetKubeCli().CoreV1().Secrets(testNamespace)
				key := make([]byte, 32)
				k8sutil.CreateEncryptionKeySecret(secrets, testRocksDBEncryptionKey, key)
			},
			ExpectedEvent: "member agent is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, testRocksDBEncryptionKey),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, true),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.RocksdbEncryptionVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
			Name: "Agent Pod can not have metrics exporter",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics:        metricsSpec,
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
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
			Name: "DBserver Pod with metrics exporter",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics:        metricsSpec,
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, emptyResources),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics: api.MetricsSpec{
						Enabled: util.NewBool(true),
						Image:   util.NewString(testExporterImage),
						Authentication: api.MetricsAuthenticationSpec{
							JWTTokenSecretName: util.NewString(testExporterToken),
						},
						Resources: resourcesUnfiltered,
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered)),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics:        metricsSpec,
					Lifecycle: api.LifecycleSpec{
						Resources: resourcesUnfiltered,
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes
			},
			config: Config{
				LifecycleImage: testImageLifecycle,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered)),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Env: []core.EnvVar{
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							Ports: createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
							},
							Resources:       emptyResources,
							Lifecycle:       createTestLifecycle(),
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, emptyResources),
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
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics:        metricsSpec,
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes
			},
			config: Config{
				LifecycleImage:        testImageLifecycle,
				OperatorUUIDInitImage: testImageOperatorUUIDInit,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
						createTestAlpineContainer(firstDBServerStatus.ID, false),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Env: []core.EnvVar{
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							Ports: createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
							},
							Resources:       emptyResources,
							Lifecycle:       createTestLifecycle(),
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, emptyResources),
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
			Name: "DBserver Pod with metrics exporter, lifecycle, tls, authentication, license, rocksDB encryption, secured liveness",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					TLS:            tlsSpec,
					Metrics:        metricsSpec,
					RocksDB:        rocksDBSpec,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					License: api.LicenseSpec{
						SecretName: util.NewString(testLicense),
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

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes

				secrets := deployment.GetKubeCli().CoreV1().Secrets(testNamespace)
				key := make([]byte, 32)
				k8sutil.CreateEncryptionKeySecret(secrets, testRocksDBEncryptionKey, key)

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(true,
					authorization, k8sutil.ArangoPort)
			},
			config: Config{
				LifecycleImage: testImageLifecycle,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupDBServersString, firstDBServerStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, testRocksDBEncryptionKey),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, true, true, true),
							Env: []core.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
									testLicense, constants.SecretKeyToken),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							Ports:           createTestPorts(),
							Lifecycle:       createTestLifecycle(),
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.RocksdbEncryptionVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							Resources: emptyResources,
						},
						func() core.Container {
							c := testCreateExporterContainer(true, emptyResources)
							c.VolumeMounts = append(c.VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
							return c
						}(),
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultDBServerTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupDBServersString + "-" + firstDBServerStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						true, ""),
				},
			},
		},
		{
			Name: "Coordinator Pod with TLS and authentication and readiness and liveness",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					TLS: api.TLSSpec{
						CASecretName: util.NewString(testCASecretName),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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

				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(true, auth)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            k8sutil.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForCoordinator(firstCoordinatorStatus.ID, true, true, false),
							Ports:           createTestPorts(),
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
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					TLS: api.TLSSpec{
						CASecretName: util.NewString(testCASecretName),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
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

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(true,
					authLiveness, 0)
				testCase.ExpectedPod.Spec.Containers[0].ReadinessProbe = createTestReadinessProbe(true, authReadiness)
			},
			ExpectedEvent: "member single is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupSingleString, singleStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:            k8sutil.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForSingleMode(singleStatus.ID, true, true, false),
							Ports:           createTestPorts(),
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
	}

	runTestCases(t, testCases...)
}
