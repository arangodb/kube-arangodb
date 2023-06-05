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
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func TestEnsurePod_ArangoDB_Encryption(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Agent CE 3.7.0 Pod with encrypted rocksdb",
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
					Images: createTestImagesWithVersion(false, testVersion),
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
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: BuildTestAgentArgs(t, firstAgentStatus.ID,
								AgentArgsWithTLS(firstAgentStatus.ID, false),
								ArgsWithAuth(false),
								ArgsWithEncryptionKey()),
							Ports: createTestPorts(api.ServerGroupAgents),
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
			Name: "DBserver CE 3.7.0 Pod with metrics exporter, lifecycle, tls, authentication, license, rocksDB encryption, secured liveness",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					TLS:            tlsSpec,
					Metrics:        metricsSpec,
					RocksDB:        rocksDBSpec,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
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
					Images: createTestImagesWithVersion(false, testVersion),
				}

				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
				testCase.ExpectedPod.ObjectMeta.Labels[k8sutil.LabelKeyArangoExporter] = testYes

				key := make([]byte, 32)
				k8sutil.CreateEncryptionKeySecret(deployment.SecretsModInterface(), testRocksDBEncryptionKey, key)

				authorization, err := createTestToken(deployment, testCase, []string{"/_api/version"})
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(httpProbe, true, authorization, shared.ServerPortName)
			},
			config: Config{
				OperatorImage: testImageOperator,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupDBServersString, firstDBServerStatus.ID),
						k8sutil.CreateVolumeWithSecret(shared.RocksdbEncryptionVolumeName, testRocksDBEncryptionKey),
						k8sutil.CreateVolumeWithSecret(shared.ExporterJWTVolumeName, testExporterToken),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:            shared.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForDBServer(firstDBServerStatus.ID, true, true, true),
							Ports:           createTestPorts(api.ServerGroupDBServers),
							Lifecycle:       createTestLifecycle(api.ServerGroupAgents),
							LivenessProbe:   createTestLivenessProbe(httpProbe, false, "", shared.ServerPortName),
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
							c := testArangodbInternalExporterContainer(true, true, emptyResources)
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
			Name: "Agent EE 3.7.0 Pod with encrypted rocksdb, disabled feature",
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
					Images: createTestImagesWithVersion(true, testVersion),
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
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: BuildTestAgentArgs(t, firstAgentStatus.ID,
								AgentArgsWithTLS(firstAgentStatus.ID, false),
								ArgsWithAuth(false),
								ArgsWithEncryptionKey()),
							Ports: createTestPorts(api.ServerGroupAgents),
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
			Name: "Agent EE 3.7.0 Pod with encrypted rocksdb, enabled feature",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					TLS:            noTLS,
					RocksDB:        rocksDBSpec,
				},
			},
			Features: testCaseFeatures{
				EncryptionRotation: true,
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: createTestImagesWithVersion(true, testVersion),
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
						k8sutil.CreateVolumeWithSecret(shared.RocksdbEncryptionVolumeName, fmt.Sprintf("%s-encryption-folder", testDeploymentName)),
					},
					Containers: []core.Container{
						{
							Name:  shared.ServerContainerName,
							Image: testImage,
							Command: BuildTestAgentArgs(t, firstAgentStatus.ID,
								AgentArgsWithTLS(firstAgentStatus.ID, false),
								ArgsWithAuth(false),
								ArgsWithEncryptionFolder(), func(t *testing.T) map[string]string {
									return map[string]string{
										"rocksdb.encryption-key-rotation": "true",
									}
								}),
							Ports: createTestPorts(api.ServerGroupAgents),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.RocksdbEncryptionReadOnlyVolumeMount(),
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
	}

	runTestCases(t, testCases...)
}
