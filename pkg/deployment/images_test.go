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
	"crypto/sha1"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tolerations"
)

const (
	testNewImage = testImage + "2"
)

type testCaseImageUpdate struct {
	Name             string
	ArangoDeployment *api.ArangoDeployment
	Before           func(*testing.T, *Deployment)
	After            func(*testing.T, *Deployment)
	ExpectedError    error
	RetrySoon        bool
	ExpectedPod      core.Pod
}

func TestEnsureImages(t *testing.T) {
	// Arange
	terminationGracePeriodSeconds := int64((time.Second * 30).Seconds())
	id := fmt.Sprintf("%0x", sha1.Sum([]byte(testNewImage)))[:6]
	hostname := testDeploymentName + "-" + api.ServerGroupImageDiscovery.AsRole() + "-" + id

	var securityContext api.ServerGroupSpecSecurityContext

	testCases := []testCaseImageUpdate{
		{
			Name: "Image has not been changed",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testImage),
				},
			},
		},
		{
			Name: "Image has been changed",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
				},
			},
			RetrySoon: true,
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testNewImage,
							Command: createTestCommandForImageUpdatePod(),
							Ports:   createTestPorts(api.ServerGroupAgents),
							Resources: core.ResourceRequirements{
								Limits:   make(core.ResourceList),
								Requests: make(core.ResourceList),
							},
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					Tolerations:                   getTestTolerations(),
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Hostname:                      hostname,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName,
						api.ServerGroupImageDiscovery.AsRole(), false, ""),
				},
			},
		},
		{
			Before: func(t *testing.T, deployment *Deployment) {
				c := deployment.acs.CurrentClusterCache()

				s := &core.Secret{
					ObjectMeta: meta.ObjectMeta{
						Name:      testLicense,
						Namespace: testNamespace,
					},
					Data: map[string][]byte{
						constants.SecretKeyToken: []byte("data"),
					},
				}

				_, err := c.SecretsModInterface().V1().Create(context.Background(), s, meta.CreateOptions{})
				require.NoError(t, err)
			},
			Name: "Image not been changed with license (proper one)",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			RetrySoon: true,
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testNewImage,
							Command: createTestCommandForImageUpdatePod(),
							Ports:   createTestPorts(api.ServerGroupAgents),
							Env: []core.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
									testLicense, constants.SecretKeyToken),
							},
							Resources: core.ResourceRequirements{
								Limits:   make(core.ResourceList),
								Requests: make(core.ResourceList),
							},
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					Tolerations:                   getTestTolerations(),
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Hostname:                      hostname,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName,
						api.ServerGroupImageDiscovery.AsRole(), false, ""),
				},
			},
		},
		{
			Name: "Image not been changed with license (missing one)",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			RetrySoon: true,
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testNewImage,
							Command: createTestCommandForImageUpdatePod(),
							Ports:   createTestPorts(api.ServerGroupAgents),
							Resources: core.ResourceRequirements{
								Limits:   make(core.ResourceList),
								Requests: make(core.ResourceList),
							},
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					Tolerations:                   getTestTolerations(),
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Hostname:                      hostname,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName,
						api.ServerGroupImageDiscovery.AsRole(), false, ""),
				},
			},
		},
		{
			Before: func(t *testing.T, deployment *Deployment) {
				c := deployment.acs.CurrentClusterCache()

				s := &core.Secret{
					ObjectMeta: meta.ObjectMeta{
						Name:      testLicense,
						Namespace: testNamespace,
					},
				}

				_, err := c.SecretsModInterface().V1().Create(context.Background(), s, meta.CreateOptions{})
				require.NoError(t, err)
			},
			Name: "Image not been changed with license (missing key)",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
					License: api.LicenseSpec{
						SecretName: util.NewType[string](testLicense),
					},
				},
			},
			RetrySoon: true,
			ExpectedPod: core.Pod{
				Spec: core.PodSpec{
					Volumes: []core.Volume{
						k8sutil.CreateVolumeEmptyDir(shared.ArangodVolumeName),
					},
					Containers: []core.Container{
						{
							Name:    shared.ServerContainerName,
							Image:   testNewImage,
							Command: createTestCommandForImageUpdatePod(),
							Ports:   createTestPorts(api.ServerGroupAgents),
							Resources: core.ResourceRequirements{
								Limits:   make(core.ResourceList),
								Requests: make(core.ResourceList),
							},
							VolumeMounts: []core.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: core.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 core.RestartPolicyNever,
					Tolerations:                   getTestTolerations(),
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Hostname:                      hostname,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName,
						api.ServerGroupImageDiscovery.AsRole(), false, ""),
				},
			},
		},
		{
			Name: "Image is being updated in failed phase",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
				},
			},
			Before: func(t *testing.T, deployment *Deployment) {
				pod := core.Pod{
					ObjectMeta: meta.ObjectMeta{
						Name:              k8sutil.CreatePodName(testDeploymentName, api.ServerGroupImageDiscovery.AsRole(), id, ""),
						CreationTimestamp: meta.Now(),
					},
					Spec: core.PodSpec{},
					Status: core.PodStatus{
						Phase: core.PodFailed,
					},
				}

				_, err := deployment.PodsModInterface().Create(context.Background(), &pod, meta.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods := deployment.GetCachedStatus().Pod().V1().ListSimple()
				require.Len(t, pods, 1)
			},
		},
		{
			Name: "Image is being updated too long in failed phase",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
				},
			},
			Before: func(t *testing.T, deployment *Deployment) {
				pod := core.Pod{
					ObjectMeta: meta.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, api.ServerGroupImageDiscovery.AsRole(), id, ""),
					},
					Status: core.PodStatus{
						Phase: core.PodFailed,
					},
				}
				_, err := deployment.PodsModInterface().Create(context.Background(), &pod, meta.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods := deployment.GetCachedStatus().Pod().V1().ListSimple()
				require.Len(t, pods, 0)
			},
		},
		{
			Name: "Image is being updated in not ready phase",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
				},
			},
			RetrySoon: true,
			Before: func(t *testing.T, deployment *Deployment) {
				pod := core.Pod{
					ObjectMeta: meta.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, api.ServerGroupImageDiscovery.AsRole(), id, ""),
					},
					Status: core.PodStatus{
						Conditions: []core.PodCondition{
							{
								Type: core.PodScheduled,
							},
						},
					},
				}
				_, err := deployment.PodsModInterface().Create(context.Background(), &pod, meta.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods := deployment.GetCachedStatus().Pod().V1().ListSimple()
				require.Len(t, pods, 1)
			},
		},
		{
			Name: "Image is being updated in ready phase with empty statuses list",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
				},
			},
			RetrySoon: true,
			Before: func(t *testing.T, deployment *Deployment) {
				pod := core.Pod{
					ObjectMeta: meta.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, api.ServerGroupImageDiscovery.AsRole(), id, ""),
					},
					Status: core.PodStatus{
						Conditions: []core.PodCondition{
							{
								Type:   core.PodReady,
								Status: core.ConditionTrue,
							},
						},
					},
				}
				_, err := deployment.PodsModInterface().Create(context.Background(), &pod, meta.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods := deployment.GetCachedStatus().Pod().V1().ListSimple()
				require.Len(t, pods, 1)
			},
		},
		{
			Name: "Can not get API version of arnagod",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewType[string](testNewImage),
				},
			},
			RetrySoon: true,
			Before: func(t *testing.T, deployment *Deployment) {
				pod := core.Pod{
					ObjectMeta: meta.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, api.ServerGroupImageDiscovery.AsRole(), id, ""),
					},
					Status: core.PodStatus{
						Conditions: []core.PodCondition{
							{
								Type:   core.PodReady,
								Status: core.ConditionTrue,
							},
						},
						ContainerStatuses: []core.ContainerStatus{
							{},
						},
					},
				}
				_, err := deployment.PodsModInterface().Create(context.Background(), &pod, meta.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods := deployment.GetCachedStatus().Pod().V1().ListSimple()
				require.Len(t, pods, 1)
			},
		},
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			d, _ := createTestDeployment(t, Config{}, testCase.ArangoDeployment)

			d.currentObjectStatus = &api.DeploymentStatus{
				Images: createTestImages(false),
			}

			if testCase.Before != nil {
				testCase.Before(t, d)
				require.NoError(t, d.GetCachedStatus().Refresh(context.Background()))
			}

			// Create custom resource in the fake kubernetes API
			_, err := d.deps.Client.Arango().DatabaseV1().ArangoDeployments(testNamespace).Create(context.Background(), d.currentObject, meta.CreateOptions{})
			require.NoError(t, err)

			require.NoError(t, d.acs.CurrentClusterCache().Refresh(context.Background()))

			// Act
			retrySoon, _, err := d.ensureImages(context.Background(), d.currentObject, d.GetCachedStatus())

			// Assert
			assert.EqualValues(t, testCase.RetrySoon, retrySoon)
			if testCase.ExpectedError != nil {
				assert.EqualError(t, err, testCase.ExpectedError.Error())
				return
			}

			require.NoError(t, err)

			if len(testCase.ExpectedPod.Spec.Containers) > 0 {
				pods, err := d.deps.Client.Kubernetes().CoreV1().Pods(testNamespace).List(context.Background(), meta.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pods.Items, 1)
				require.Equal(t, testCase.ExpectedPod.Spec, pods.Items[0].Spec)

				ownerRef := pods.Items[0].GetOwnerReferences()
				require.Len(t, ownerRef, 1)
				require.Equal(t, ownerRef[0], testCase.ArangoDeployment.AsOwner())
			}

			require.NoError(t, d.GetCachedStatus().Refresh(context.Background()))

			if testCase.After != nil {
				testCase.After(t, d)
			}
		})
	}
}

func createTestCommandForImageUpdatePod() []string {
	return []string{resources.ArangoDExecutor,
		"--database.directory=" + shared.ArangodVolumeMountDir,
		"--log.output=+",
		"--server.authentication=false",
		fmt.Sprintf("--server.endpoint=tcp://[::]:%d", shared.ArangoPort),
	}
}

func getTestTolerations() []core.Toleration {

	shortDur := tolerations.TolerationDuration{
		Forever:  false,
		TimeSpan: time.Second * 5,
	}

	return []core.Toleration{
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeNotReady, shortDur),
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeUnreachable, shortDur),
		tolerations.NewNoExecuteToleration(tolerations.TolerationKeyNodeAlphaUnreachable, shortDur),
	}
}
