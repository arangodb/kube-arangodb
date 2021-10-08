//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEnsurePod_Sync_Error(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Sync Pod does not work for enterprise image",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testImage),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(false),
				}
			},
			ExpectedError: errors.New("Image '" + testImage + "' does not contain an Enterprise version of ArangoDB"),
		},
		{
			Name: "Sync Pod cannot get master JWT secret",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testImage),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}
			},
			ExpectedError: errors.New("Master JWT secret validation failed: secrets \"" +
				testDeploymentName + "-sync-jwt\" not found"),
		},
		{
			Name: "Sync Pod cannot get monitoring token secret",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testImage),
					Sync: api.SyncSpec{
						Enabled: util.NewBool(true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, metav1.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Monitoring token secret validation failed: secrets \"" +
				testDeploymentName + "-sync-mt\" not found"),
		},
	}

	runTestCases(t, testCases...)
}

func TestEnsurePod_Sync_Master(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Sync Master Pod cannot create TLS keyfile secret",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testImage),
					Sync: api.SyncSpec{
						Enabled: util.NewBool(true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Sync.TLS.GetCASecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, metav1.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Failed to create TLS keyfile secret: secrets \"" +
				testDeploymentName + "-sync-ca\" not found"),
		},
		{
			Name: "Sync Master Pod cannot get cluster JWT secret",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					Sync: api.SyncSpec{
						Enabled: util.NewBool(true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Authentication.GetJWTSecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, metav1.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Cluster JWT secret validation failed: secrets \"" +
				testJWTSecretName + "\" not found"),
		},
		{
			Name: "Sync Master Pod cannot get authentication CA certificate",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					Sync: api.SyncSpec{
						Enabled: util.NewBool(true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Sync.Authentication.GetClientCASecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, metav1.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Client authentication CA certificate secret validation failed: " +
				"secrets \"" + testDeploymentName + "-sync-client-auth-ca\" not found"),
		},
		{
			Name: "Sync Master Pod with authentication, monitoring, tls, service account, node selector, " +
				"liveness probe, priority class name, resource requirements",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: authenticationSpec,
					Sync: api.SyncSpec{
						Enabled: util.NewBool(true),
					},
					SyncMasters: api.ServerGroupSpec{
						ServiceAccountName: util.NewString(testServiceAccountName),
						NodeSelector:       nodeSelectorTest,
						PriorityClassName:  testPriorityClassName,
						Resources:          resourcesUnfiltered,
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)

				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().SecretReadInterface(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(
					"", true, "bearer "+auth, k8sutil.ArangoSyncMasterPort)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClientAuthCAVolumeName, "test-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName, "test-sync-jwt"),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, true, true),
							Ports:   createTestPorts(),
							Env: []core.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
									testDeploymentName+"-sync-mt", constants.SecretKeyToken),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							Resources:       resourcesUnfiltered,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClientAuthCACertificateVolumeMount(),
								k8sutil.MasterJWTVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
					},
					PriorityClassName:             testPriorityClassName,
					RestartPolicy:                 core.RestartPolicyNever,
					ServiceAccountName:            testServiceAccountName,
					NodeSelector:                  nodeSelectorTest,
					TerminationGracePeriodSeconds: &defaultSyncMasterTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupSyncMastersString + "-" +
						firstSyncMaster.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSyncMastersString,
						false, ""),
				},
			},
		},
		{
			Name: "Sync Master Pod with lifecycle, license, monitoring without authentication and alpine",
			config: Config{
				LifecycleImage: testImageLifecycle,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewBool(true),
					},
					License: api.LicenseSpec{
						SecretName: util.NewString(testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().SecretReadInterface(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(
					"", true, "bearer "+auth, k8sutil.ArangoSyncMasterPort)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true),
							Ports:   createTestPorts(),
							Env: []core.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
									testDeploymentName+"-sync-mt", constants.SecretKeyToken),
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
									testLicense, constants.SecretKeyToken),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							Resources:       emptyResources,
							ImagePullPolicy: core.PullIfNotPresent,
							Lifecycle:       createTestLifecycle(),
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.LifecycleVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClientAuthCACertificateVolumeMount(),
								k8sutil.MasterJWTVolumeMount(),
							},
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSyncMasterTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupSyncMastersString + "-" +
						firstSyncMaster.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSyncMastersString,
						true, ""),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}

func TestEnsurePod_Sync_Worker(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Sync Worker Pod with monitoring, service account, node selector, lifecycle, license " +
				"liveness probe, priority class name, resource requirements without alpine",
			config: Config{
				LifecycleImage: testImageLifecycle,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					Sync: api.SyncSpec{
						Enabled: util.NewBool(true),
					},
					SyncWorkers: api.ServerGroupSpec{
						ServiceAccountName: util.NewString(testServiceAccountName),
						NodeSelector:       nodeSelectorTest,
						PriorityClassName:  testPriorityClassName,
						Resources:          resourcesUnfiltered,
					},
					License: api.LicenseSpec{
						SecretName: util.NewString(testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncWorkers: api.MemberStatusList{
							firstSyncWorker,
						},
					},
					Images: createTestImages(true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSyncWorkers, firstSyncWorker)

				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().SecretReadInterface(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(
					"", true, "bearer "+auth, k8sutil.ArangoSyncWorkerPort)
			},
			ExpectedEvent: "member syncworker is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName, testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncWorker(firstSyncWorker.ID, true, true),
							Ports:   createTestPorts(),
							Env: []core.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
									testDeploymentName+"-sync-mt", constants.SecretKeyToken),
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
									testLicense, constants.SecretKeyToken),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							Lifecycle:       createTestLifecycle(),
							ImagePullPolicy: core.PullIfNotPresent,
							Resources:       k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.LifecycleVolumeMount(),
								k8sutil.MasterJWTVolumeMount(),
							},
						},
					},
					PriorityClassName:             testPriorityClassName,
					RestartPolicy:                 core.RestartPolicyNever,
					ServiceAccountName:            testServiceAccountName,
					NodeSelector:                  nodeSelectorTest,
					TerminationGracePeriodSeconds: &defaultSyncWorkerTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupSyncWorkersString + "-" +
						firstSyncWorker.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSyncWorkersString,
						false, api.ServerGroupDBServersString),
				},
			},
		},
	}

	runTestCases(t, testCases...)
}
