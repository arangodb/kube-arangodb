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

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/arangodb/kube-arangodb/pkg/util"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	core "k8s.io/api/core/v1"
)

func createTestDiscoveredImages(image, version, id string) api.ImageInfoList {
	return api.ImageInfoList{
		{
			Image:           image,
			ArangoDBVersion: driver.Version(version),
			ImageID:         id,
			Enterprise:      false,
		},
	}
}

func TestEnsurePod_ArangoDB_ImagePropagation(t *testing.T) {
	image := "arangodb/test:0.0.0"
	version := "0.0.0"
	imageID := "arangodb/test@sha256:xxx"

	discoveredImages := createTestDiscoveredImages(image, version, imageID)

	testCases := []testCaseStruct{
		{
			Name: "Agent Pod with defined image",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:           util.NewString(image),
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
					Images: discoveredImages,
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
							Image:   imageID,
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
			Name: "Agent Pod with defined image and defined kubelet mode",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:              util.NewString(image),
					ImageDiscoveryMode: api.NewDeploymentImageDiscoveryModeSpec(api.DeploymentImageDiscoveryKubeletMode),
					Authentication:     noAuthentication,
					TLS:                noTLS,
					ImagePullPolicy:    util.NewPullPolicy(core.PullAlways),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: discoveredImages,
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
							Image:   imageID,
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
			Name: "Agent Pod with defined image and defined direct mode",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:              util.NewString(image),
					ImageDiscoveryMode: api.NewDeploymentImageDiscoveryModeSpec(api.DeploymentImageDiscoveryDirectMode),
					Authentication:     noAuthentication,
					TLS:                noTLS,
					ImagePullPolicy:    util.NewPullPolicy(core.PullAlways),
				},
			},
			Helper: func(t *testing.T, deployment *Deployment, testCase *testCaseStruct) {
				deployment.status.last = api.DeploymentStatus{
					Members: api.DeploymentStatusMembers{
						Agents: api.MemberStatusList{
							firstAgentStatus,
						},
					},
					Images: discoveredImages,
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
							Image:   image,
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
	}

	runTestCases(t, testCases...)
}
