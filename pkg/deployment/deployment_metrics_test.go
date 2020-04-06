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

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/arangodb/kube-arangodb/pkg/util"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	core "k8s.io/api/core/v1"
)

func TestEnsurePod_Metrics(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "DBserver Pod with metrics exporter and port override",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics: func() api.MetricsSpec {
						m := metricsSpec.DeepCopy()

						m.Port = util.NewUInt16(9999)

						return *m
					}(),
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
						testCreateExporterContainerWithPort(false, emptyResources, 9999),
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
			Name: "DBserver Pod with metrics exporter with mode",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics: func() api.MetricsSpec {
						m := metricsSpec.DeepCopy()

						m.Mode = api.MetricsModeExporter.New()

						return *m
					}(),
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
			Name: "DBserver Pod with metrics exporter with internal mode",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics: func() api.MetricsSpec {
						m := metricsSpec.DeepCopy()

						m.Mode = api.MetricsModeInternal.New()

						return *m
					}(),
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
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports: func()[]core.ContainerPort {
								ports := createTestPorts()

								ports = append(ports, core.ContainerPort{
									Name: "exporter",
									Protocol: core.ProtocolTCP,
									ContainerPort: k8sutil.ArangoPort,
								})

								return ports
							}(),
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
					Hostname:                      testDeploymentName + "-" + api.ServerGroupDBServersString + "-" + firstDBServerStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupDBServersString,
						false, ""),
				},
			},
		},
		{
			Name: "Agent Pod with metrics exporter with internal mode",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics: func() api.MetricsSpec {
						m := metricsSpec.DeepCopy()

						m.Mode = api.MetricsModeInternal.New()

						return *m
					}(),
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
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes
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
							Ports: func()[]core.ContainerPort {
								ports := createTestPorts()

								ports = append(ports, core.ContainerPort{
									Name: "exporter",
									Protocol: core.ProtocolTCP,
									ContainerPort: k8sutil.ArangoPort,
								})

								return ports
							}(),
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
			Name: "DBserver Pod with sidecar metrics exporter and port override",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics: func() api.MetricsSpec {
						m := metricsSpec.DeepCopy()

						m.Port = util.NewUInt16(9999)

						m.Mode = api.MetricsModeSidecar.New()

						return *m
					}(),
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
						func() core.Container {
									z := testCreateExporterContainerWithPort(false, emptyResources, 9999)
							z.Command = append(z.Command, "--mode=passthru")
							return z
						}(),
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
			Name: "Agency Pod with sidecar metrics exporter and port override",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					Metrics: func() api.MetricsSpec {
						m := metricsSpec.DeepCopy()

						m.Port = util.NewUInt16(9999)

						m.Mode = api.MetricsModeSidecar.New()

						return *m
					}(),
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
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes
			},
			ExpectedEvent: "member agent is created",
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
						func() core.Container {
							z := testCreateExporterContainerWithPort(false, emptyResources, 9999)
							z.Command = append(z.Command, "--mode=passthru")
							return z
						}(),
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
	}

	runTestCases(t, testCases...)
}
