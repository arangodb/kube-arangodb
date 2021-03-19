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
	"context"
	"crypto/sha1"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/stretchr/testify/require"
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
	ExpectedPod      v1.Pod
}

func TestEnsureImages(t *testing.T) {
	// Arange
	terminationGracePeriodSeconds := int64((time.Second * 30).Seconds())
	id := fmt.Sprintf("%0x", sha1.Sum([]byte(testNewImage)))[:6]
	hostname := testDeploymentName + "-" + k8sutil.ImageIDAndVersionRole + "-" + id

	var securityContext api.ServerGroupSpecSecurityContext

	testCases := []testCaseImageUpdate{
		{
			Name: "Image has not been changed",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testImage),
				},
			},
		},
		{
			Name: "Image has been changed",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testNewImage),
				},
			},
			RetrySoon: true,
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testNewImage,
							Command: createTestCommandForImageUpdatePod(),
							Ports:   createTestPorts(),
							Resources: v1.ResourceRequirements{
								Limits:   make(v1.ResourceList),
								Requests: make(v1.ResourceList),
							},
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
					Tolerations:                   getTestTolerations(),
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Hostname:                      hostname,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName,
						k8sutil.ImageIDAndVersionRole, false, ""),
				},
			},
		},
		{
			Name: "Image not been changed with license",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testNewImage),
					License: api.LicenseSpec{
						SecretName: util.NewString(testLicense),
					},
				},
			},
			RetrySoon: true,
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testNewImage,
							Command: createTestCommandForImageUpdatePod(),
							Ports:   createTestPorts(),
							Env: []v1.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
									testLicense, constants.SecretKeyToken),
							},
							Resources: v1.ResourceRequirements{
								Limits:   make(v1.ResourceList),
								Requests: make(v1.ResourceList),
							},
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
					Tolerations:                   getTestTolerations(),
					TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
					Hostname:                      hostname,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName,
						k8sutil.ImageIDAndVersionRole, false, ""),
				},
			},
		},
		{
			Name: "Image is being updated in failed phase",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testNewImage),
				},
			},
			Before: func(t *testing.T, deployment *Deployment) {
				pod := v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:              k8sutil.CreatePodName(testDeploymentName, k8sutil.ImageIDAndVersionRole, id, ""),
						CreationTimestamp: metav1.Now(),
					},
					Spec: v1.PodSpec{},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				}

				_, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).Create(context.Background(), &pod, metav1.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).List(context.Background(), metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pods.Items, 1)
			},
		},
		{
			Name: "Image is being updated too long in failed phase",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testNewImage),
				},
			},
			Before: func(t *testing.T, deployment *Deployment) {
				pod := v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, k8sutil.ImageIDAndVersionRole, id, ""),
					},
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				}
				_, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).Create(context.Background(), &pod, metav1.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).List(context.Background(), metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pods.Items, 0)
			},
		},
		{
			Name: "Image is being updated in not ready phase",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testNewImage),
				},
			},
			RetrySoon: true,
			Before: func(t *testing.T, deployment *Deployment) {
				pod := v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, k8sutil.ImageIDAndVersionRole, id, ""),
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type: v1.PodScheduled,
							},
						},
					},
				}
				_, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).Create(context.Background(), &pod, metav1.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).List(context.Background(), metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pods.Items, 1)
			},
		},
		{
			Name: "Image is being updated in ready phase with empty statuses list",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testNewImage),
				},
			},
			RetrySoon: true,
			Before: func(t *testing.T, deployment *Deployment) {
				pod := v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, k8sutil.ImageIDAndVersionRole, id, ""),
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   v1.PodReady,
								Status: v1.ConditionTrue,
							},
						},
					},
				}
				_, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).Create(context.Background(), &pod, metav1.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).List(context.Background(), metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pods.Items, 1)
			},
		},
		{
			Name: "Can not get API version of arnagod",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image: util.NewString(testNewImage),
				},
			},
			RetrySoon: true,
			Before: func(t *testing.T, deployment *Deployment) {
				pod := v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: k8sutil.CreatePodName(testDeploymentName, k8sutil.ImageIDAndVersionRole, id, ""),
					},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   v1.PodReady,
								Status: v1.ConditionTrue,
							},
						},
						ContainerStatuses: []v1.ContainerStatus{
							{},
						},
					},
				}
				_, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).Create(context.Background(), &pod, metav1.CreateOptions{})
				require.NoError(t, err)
			},
			After: func(t *testing.T, deployment *Deployment) {
				pods, err := deployment.GetKubeCli().CoreV1().Pods(testNamespace).List(context.Background(), metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pods.Items, 1)
			},
		},
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			d, _ := createTestDeployment(Config{}, testCase.ArangoDeployment)

			d.status.last = api.DeploymentStatus{
				Images: createTestImages(false),
			}

			if testCase.Before != nil {
				testCase.Before(t, d)
			}

			// Create custom resource in the fake kubernetes API
			_, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(testNamespace).Create(context.Background(), d.apiObject, metav1.CreateOptions{})
			require.NoError(t, err)

			// Act
			retrySoon, _, err := d.ensureImages(d.apiObject)

			// Assert
			assert.EqualValues(t, testCase.RetrySoon, retrySoon)
			if testCase.ExpectedError != nil {
				assert.EqualError(t, err, testCase.ExpectedError.Error())
				return
			}

			require.NoError(t, err)

			if len(testCase.ExpectedPod.Spec.Containers) > 0 {
				pods, err := d.deps.KubeCli.CoreV1().Pods(testNamespace).List(context.Background(), metav1.ListOptions{})
				require.NoError(t, err)
				require.Len(t, pods.Items, 1)
				require.Equal(t, testCase.ExpectedPod.Spec, pods.Items[0].Spec)

				ownerRef := pods.Items[0].GetOwnerReferences()
				require.Len(t, ownerRef, 1)
				require.Equal(t, ownerRef[0], testCase.ArangoDeployment.AsOwner())
			}

			if testCase.After != nil {
				testCase.After(t, d)
			}
		})
	}
}

func createTestCommandForImageUpdatePod() []string {
	return []string{resources.ArangoDExecutor,
		"--server.authentication=false",
		fmt.Sprintf("--server.endpoint=tcp://[::]:%d", k8sutil.ArangoPort),
		"--database.directory=" + k8sutil.ArangodVolumeMountDir,
		"--log.output=+",
	}
}

func getTestTolerations() []v1.Toleration {

	shortDur := k8sutil.TolerationDuration{
		Forever:  false,
		TimeSpan: time.Second * 5,
	}

	return []v1.Toleration{
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeNotReady, shortDur),
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeUnreachable, shortDur),
		k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeAlphaUnreachable, shortDur),
	}
}
