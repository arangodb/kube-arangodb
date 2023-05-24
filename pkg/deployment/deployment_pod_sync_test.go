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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func TestEnsurePod_Sync_Error(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Sync Pod does not work for enterprise image",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testImage),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
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
					Image: util.NewType[string](testImage),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
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
					Image: util.NewType[string](testImage),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, meta.DeleteOptions{})
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
					Image: util.NewType[string](testImage),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Sync.TLS.GetCASecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, meta.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Failed to create TLS keyfile secret: secrets \"" +
				testDeploymentName + "-sync-ca\" not found"),
		},
		{
			Name: "Sync Master Pod cannot get cluster JWT secret",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Authentication.GetJWTSecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, meta.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Cluster JWT secret validation failed: secrets \"" +
				testJWTSecretName + "\" not found"),
		},
		{
			Name: "Sync Master Pod cannot get authentication CA certificate",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				secretName := testCase.ArangoDeployment.Spec.Sync.Authentication.GetClientCASecretName()
				err := deployment.SecretsModInterface().Delete(context.Background(), secretName, meta.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Client authentication CA certificate secret validation failed: " +
				"secrets \"" + testDeploymentName + "-sync-client-auth-ca\" not found"),
		},
		{
			DropInit: true,
			Name: "Sync Master Pod with authentication, monitoring, tls, service account, node selector, " +
				"liveness probe, priority class name, resource requirements",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: authenticationSpec,
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
					SyncMasters: api.ServerGroupSpec{
						ServiceAccountName: util.NewType[string](testServiceAccountName),
						NodeSelector:       nodeSelectorTest,
						PriorityClassName:  testPriorityClassName,
						Resources:          resourcesUnfiltered,
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)

				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName, "test-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName, "test-sync-jwt"),
						k8sutil.CreateVolumeWithSecret(shared.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, true, true),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
							Env: []core.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
									testDeploymentName+"-sync-mt", constants.SecretKeyToken),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
							Resources:       resourcesUnfiltered,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.LifecycleVolumeMount(),
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
			DropInit: true,
			Name:     "Sync Master Pod with lifecycle, license, monitoring without authentication and alpine",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - existing service, ClusterIP and valid name",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
						ExternalAccess: api.SyncExternalAccessSpec{
							MasterEndpoint: []string{
								"https://arangodb.xyz:8629",
							},
						},
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				svc := &core.Service{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-sync",
					},
					Spec: core.ServiceSpec{
						ClusterIP: "1.2.3.4",
					},
				}

				deployment.GetCachedStatus().ServicesModInterface().V1().Create(context.Background(), svc, meta.CreateOptions{})

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true, "https://arangodb.xyz:8629"),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
					HostAliases: []core.HostAlias{
						{
							IP: "1.2.3.4",
							Hostnames: []string{
								"arangodb.xyz",
							},
						},
					},
				},
			},
		},
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - missing service and valid name",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
						ExternalAccess: api.SyncExternalAccessSpec{
							MasterEndpoint: []string{
								"https://arangodb.xyz:8629",
							},
						},
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true, "https://arangodb.xyz:8629"),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - existing service, missing ClusterIP and valid name",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
						ExternalAccess: api.SyncExternalAccessSpec{
							MasterEndpoint: []string{
								"https://arangodb.xyz:8629",
							},
						},
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				svc := &core.Service{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-sync",
					},
				}

				deployment.GetCachedStatus().ServicesModInterface().V1().Create(context.Background(), svc, meta.CreateOptions{})

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true, "https://arangodb.xyz:8629"),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - existing service, ClusterIP and invalid name",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
						ExternalAccess: api.SyncExternalAccessSpec{
							MasterEndpoint: []string{
								"https://127.0.0.1:8629",
							},
						},
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				svc := &core.Service{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-sync",
					},
					Spec: core.ServiceSpec{
						ClusterIP: "1.2.3.4",
					},
				}

				deployment.GetCachedStatus().ServicesModInterface().V1().Create(context.Background(), svc, meta.CreateOptions{})

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true, "https://127.0.0.1:8629"),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - existing service, ClusterIP and missing name",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				svc := &core.Service{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-sync",
					},
					Spec: core.ServiceSpec{
						ClusterIP: "1.2.3.4",
					},
				}

				deployment.GetCachedStatus().ServicesModInterface().V1().Create(context.Background(), svc, meta.CreateOptions{})

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - existing service, ClusterIP and valid names",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
						ExternalAccess: api.SyncExternalAccessSpec{
							MasterEndpoint: []string{
								"https://arangodb.xyz:8629",
								"https://arangodb1.xyz:8629",
								"https://arangodb2.xyz:8629",
							},
						},
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				svc := &core.Service{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-sync",
					},
					Spec: core.ServiceSpec{
						ClusterIP: "1.2.3.4",
					},
				}

				deployment.GetCachedStatus().ServicesModInterface().V1().Create(context.Background(), svc, meta.CreateOptions{})

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true, "https://arangodb.xyz:8629", "https://arangodb1.xyz:8629", "https://arangodb2.xyz:8629"),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
					HostAliases: []core.HostAlias{
						{
							IP: "1.2.3.4",
							Hostnames: []string{
								"arangodb.xyz",
								"arangodb1.xyz",
								"arangodb2.xyz",
							},
						},
					},
				},
			},
		},
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - existing service, ClusterIP and valid names with different ports",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
						ExternalAccess: api.SyncExternalAccessSpec{
							MasterEndpoint: []string{
								"https://arangodb.xyz:8629",
								"https://arangodb.xyz:8639",
								"https://arangodb.xyz:8649",
							},
						},
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				svc := &core.Service{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-sync",
					},
					Spec: core.ServiceSpec{
						ClusterIP: "1.2.3.4",
					},
				}

				deployment.GetCachedStatus().ServicesModInterface().V1().Create(context.Background(), svc, meta.CreateOptions{})

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true, "https://arangodb.xyz:8629", "https://arangodb.xyz:8639", "https://arangodb.xyz:8649"),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
					HostAliases: []core.HostAlias{
						{
							IP: "1.2.3.4",
							Hostnames: []string{
								"arangodb.xyz",
							},
						},
					},
				},
			},
		},
		{
			DropInit: true,
			Name:     "Sync Master Pod alias - existing service, ClusterIP and mixed names",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Environment:    api.NewEnvironment(api.EnvironmentProduction),
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
						ExternalAccess: api.SyncExternalAccessSpec{
							MasterEndpoint: []string{
								"https://arangodb2.xyz:8629",
								"https://arangodb.xyz:8629",
								"https://127.0.0.1:8629",
							},
						},
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncMasters: api.MemberStatusList{
							firstSyncMaster,
						},
					},
					Images: createTestImages(true),
				}

				svc := &core.Service{
					ObjectMeta: meta.ObjectMeta{
						Name: "test-sync",
					},
					Spec: core.ServiceSpec{
						ClusterIP: "1.2.3.4",
					},
				}

				deployment.GetCachedStatus().ServicesModInterface().V1().Create(context.Background(), svc, meta.CreateOptions{})

				testCase.createTestPodData(deployment, api.ServerGroupSyncMasters, firstSyncMaster)
				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(shared.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true, "https://127.0.0.1:8629", "https://arangodb.xyz:8629", "https://arangodb2.xyz:8629"),
							Ports:   createTestPorts(api.ServerGroupSyncMasters),
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
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
					HostAliases: []core.HostAlias{
						{
							IP: "1.2.3.4",
							Hostnames: []string{
								"arangodb.xyz",
								"arangodb2.xyz",
							},
						},
					},
				},
			},
		},
	}

	runTestCases(t, testCases...)
}

func TestEnsurePod_Sync_Worker(t *testing.T) {
	testCases := []testCaseStruct{
		{
			DropInit: true,
			Name: "Sync Worker Pod with monitoring, service account, node selector, lifecycle, license " +
				"liveness probe, priority class name, resource requirements without alpine",
			config: Config{
				OperatorImage: testImageOperator,
			},
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewType[string](testImage),
					Authentication: noAuthentication,
					Sync: api.SyncSpec{
						Enabled: util.NewType[bool](true),
					},
					SyncWorkers: api.ServerGroupSpec{
						ServiceAccountName: util.NewType[string](testServiceAccountName),
						NodeSelector:       nodeSelectorTest,
						PriorityClassName:  testPriorityClassName,
						Resources:          resourcesUnfiltered,
					},
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.currentObjectStatus = &api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						SyncWorkers: api.MemberStatusList{
							firstSyncWorker,
						},
					},
					Images: createTestImages(true),
				}

				testCase.createTestPodData(deployment, api.ServerGroupSyncWorkers, firstSyncWorker)

				name := testCase.ArangoDeployment.Spec.Sync.Monitoring.GetTokenSecretName()
				auth, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().Secret().V1().Read(), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe("", true, "bearer "+auth, shared.ServerPortName)
			},
			ExpectedEvent: "member syncworker is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.LifecycleVolume(),
						k8sutil.CreateVolumeWithSecret(shared.MasterJWTSecretVolumeName, testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []core.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncWorker(firstSyncWorker.ID, true, true),
							Ports:   createTestPorts(api.ServerGroupSyncWorkers),
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
							ImagePullPolicy: core.PullIfNotPresent,
							Lifecycle:       createTestLifecycle(api.ServerGroupSyncMasters),
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
