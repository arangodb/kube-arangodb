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
	"fmt"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	core "k8s.io/api/core/v1"
)

func TestEnsurePod_ArangoDB_Encryption(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Agent CE 3.7.0 Pod with encrypted rocksdb",
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
					Images: createTestImagesWithVersion(false, "3.7.0"),
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
							Name:  k8sutil.ServerContainerName,
							Image: testImage,
							Command: BuildTestAgentArgs(t, firstAgentStatus.ID,
								AgentArgsWithTLS(firstAgentStatus.ID, false),
								ArgsWithAuth(false),
								ArgsWithEncryptionKey()),
							Ports: createTestPorts(),
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
			Name: "DBserver CE 3.7.0 Pod with metrics exporter, lifecycle, tls, authentication, license, rocksDB encryption, secured liveness",
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
					Images: createTestImagesWithVersion(false, "3.7.0"),
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
			Name: "Agent EE 3.7.0 Pod with encrypted rocksdb",
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
					Images: createTestImagesWithVersion(true, "3.7.0"),
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
						k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, fmt.Sprintf("%s-encryption-folder", testDeploymentName)),
					},
					Containers: []core.Container{
						{
							Name:  k8sutil.ServerContainerName,
							Image: testImage,
							Command: BuildTestAgentArgs(t, firstAgentStatus.ID,
								AgentArgsWithTLS(firstAgentStatus.ID, false),
								ArgsWithAuth(false),
								ArgsWithEncryptionFolder()),
							Ports: createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.RocksdbEncryptionReadOnlyVolumeMount(),
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
	}

	runTestCases(t, testCases...)
}
