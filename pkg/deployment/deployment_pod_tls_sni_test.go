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
	"context"
	"fmt"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/stretchr/testify/require"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
)

func createTLSSNISecret(t *testing.T, client kubernetes.Interface, name, namespace string) {
	secret := core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type: core.SecretTypeOpaque,
		Data: map[string][]byte{},
	}
	secret.Data[constants.SecretTLSKeyfile] = []byte("")

	_, err := client.CoreV1().Secrets(namespace).Create(context.Background(), &secret, meta.CreateOptions{})
	require.NoError(t, err)
}

func TestEnsurePod_ArangoDB_TLS_SNI(t *testing.T) {
	testCases := []testCaseStruct{
		{
			Name: "Pod SNI Mounts",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS: func() api.TLSSpec {
						s := tlsSpec.DeepCopy()

						s.SNI = &api.TLSSNISpec{
							Mapping: map[string][]string{
								"sni1": {
									"a",
									"b",
								},
								"sni2": {
									"c",
									"d",
								},
							},
						}

						return *s
					}(),
				},
			},
			Features: testCaseFeatures{
				TLSSNI: true,
			},
			Resources: func(t *testing.T, deployment *Deployment) {
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni1", deployment.Namespace())
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni2", deployment.Namespace())
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
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForCoordinator(firstCoordinatorStatus.ID, true, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
							},
							Resources:       emptyResources,
							ReadinessProbe:  createTestReadinessProbe(httpProbe, true, ""),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" + firstCoordinatorStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						false, ""),
				},
			},
		},
		{
			Name: "Pod SNI Mounts - Enterprise - 3.6.0",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS: func() api.TLSSpec {
						s := tlsSpec.DeepCopy()

						s.SNI = &api.TLSSNISpec{
							Mapping: map[string][]string{
								"sni1": {
									"a",
									"b",
								},
								"sni2": {
									"c",
									"d",
								},
							}}

						return *s
					}(),
				},
			},
			Features: testCaseFeatures{
				TLSSNI: true,
			},
			Resources: func(t *testing.T, deployment *Deployment) {
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni1", deployment.Namespace())
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni2", deployment.Namespace())
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Coordinators: api.MemberStatusList{
							firstCoordinatorStatus,
						},
					},
					Images: createTestImagesWithVersion(true, "3.6.0"),
				}
				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForCoordinator(firstCoordinatorStatus.ID, true, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
							},
							Resources:       emptyResources,
							ReadinessProbe:  createTestReadinessProbe(httpProbe, true, ""),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" + firstCoordinatorStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						false, ""),
				},
			},
		},
		{
			Name: "Pod SNI Mounts - 3.7.0",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS: func() api.TLSSpec {
						s := tlsSpec.DeepCopy()

						s.SNI = &api.TLSSNISpec{
							Mapping: map[string][]string{
								"sni1": {
									"a",
									"b",
								},
								"sni2": {
									"c",
									"d",
								},
							}}

						return *s
					}(),
				},
			},
			Features: testCaseFeatures{
				TLSSNI: true,
			},
			Resources: func(t *testing.T, deployment *Deployment) {
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni1", deployment.Namespace())
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni2", deployment.Namespace())
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Coordinators: api.MemberStatusList{
							firstCoordinatorStatus,
						},
					},
					Images: createTestImagesWithVersion(false, "3.7.0"),
				}
				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
					},
					Containers: []core.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForCoordinator(firstCoordinatorStatus.ID, true, false),
							Ports:   createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
							},
							Resources:       emptyResources,
							ReadinessProbe:  createTestReadinessProbe(httpProbe, true, ""),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" + firstCoordinatorStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						false, ""),
				},
			},
		},
		{
			Name: "Pod SNI Mounts - Enterprise- 3.7.0",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS: func() api.TLSSpec {
						s := tlsSpec.DeepCopy()

						s.SNI = &api.TLSSNISpec{
							Mapping: map[string][]string{
								"sni1": {
									"a",
									"b",
								},
								"sni2": {
									"c",
									"d",
								},
							}}

						return *s
					}(),
				},
			},
			Features: testCaseFeatures{
				TLSSNI: true,
			},
			Resources: func(t *testing.T, deployment *Deployment) {
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni1", deployment.Namespace())
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni2", deployment.Namespace())
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Coordinators: api.MemberStatusList{
							firstCoordinatorStatus,
						},
					},
					Images: createTestImagesWithVersion(true, "3.7.0"),
				}
				testCase.createTestPodData(deployment, api.ServerGroupCoordinators, firstCoordinatorStatus)
			},
			ExpectedEvent: "member coordinator is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
						{
							Name: "sni-1b43a8b9b6df3d38b4ef394346283cd5aeda46a9b61d52da",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: "sni1",
								},
							},
						},
						{
							Name: "sni-bbd5fc9d5151a1294ffb5de7b85ee74b7f4620021b5891e4",
							VolumeSource: core.VolumeSource{
								Secret: &core.SecretVolumeSource{
									SecretName: "sni2",
								},
							},
						},
					},
					Containers: []core.Container{
						{
							Name:  k8sutil.ServerContainerName,
							Image: testImage,
							Command: func() []string {
								args := createTestCommandForCoordinator(firstCoordinatorStatus.ID, true, false)
								args = append(args, fmt.Sprintf("--ssl.server-name-indication=a=%s/sni1/tls.keyfile", k8sutil.TLSSNIKeyfileVolumeMountDir),
									fmt.Sprintf("--ssl.server-name-indication=b=%s/sni1/tls.keyfile", k8sutil.TLSSNIKeyfileVolumeMountDir),
									fmt.Sprintf("--ssl.server-name-indication=c=%s/sni2/tls.keyfile", k8sutil.TLSSNIKeyfileVolumeMountDir),
									fmt.Sprintf("--ssl.server-name-indication=d=%s/sni2/tls.keyfile", k8sutil.TLSSNIKeyfileVolumeMountDir))
								return args
							}(),
							Ports: createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								{
									Name:      "sni-1b43a8b9b6df3d38b4ef394346283cd5aeda46a9b61d52da",
									MountPath: k8sutil.TLSSNIKeyfileVolumeMountDir + "/sni1",
									ReadOnly:  true,
								},
								{
									Name:      "sni-bbd5fc9d5151a1294ffb5de7b85ee74b7f4620021b5891e4",
									MountPath: k8sutil.TLSSNIKeyfileVolumeMountDir + "/sni2",
									ReadOnly:  true,
								},
							},
							Resources:       emptyResources,
							ReadinessProbe:  createTestReadinessProbe(httpProbe, true, ""),
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultCoordinatorTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupCoordinatorsString + "-" + firstCoordinatorStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupCoordinatorsString,
						false, ""),
				},
			},
		},
		{
			Name: "Pod SNI Mounts - Enterprise - 3.7.0 - DBServer",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:          util.NewString(testImage),
					Authentication: noAuthentication,
					TLS: func() api.TLSSpec {
						s := tlsSpec.DeepCopy()

						s.SNI = &api.TLSSNISpec{
							Mapping: map[string][]string{
								"sni1": {
									"a",
									"b",
								},
								"sni2": {
									"c",
									"d",
								},
							}}

						return *s
					}(),
				},
			},
			Features: testCaseFeatures{
				TLSSNI: true,
			},
			Resources: func(t *testing.T, deployment *Deployment) {
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni1", deployment.Namespace())
				createTLSSNISecret(t, deployment.GetKubeCli(), "sni2", deployment.Namespace())
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						DBServers: api.MemberStatusList{
							firstDBServerStatus,
						},
					},
					Images: createTestImagesWithVersion(true, "3.7.0"),
				}
				testCase.createTestPodData(deployment, api.ServerGroupDBServers, firstDBServerStatus)
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupDBServersString, firstDBServerStatus.ID),
					},
					Containers: []core.Container{
						{
							Name:  k8sutil.ServerContainerName,
							Image: testImage,
							Command: func() []string {
								args := createTestCommandForDBServer(firstDBServerStatus.ID, true, false, false)
								return args
							}(),
							Ports: createTestPorts(),
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
							},
							Resources:       emptyResources,
							LivenessProbe:   createTestLivenessProbe(httpProbe, true, "", k8sutil.ArangoPort),
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
	}

	runTestCases(t, testCases...)
}
