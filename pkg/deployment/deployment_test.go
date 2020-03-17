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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	"github.com/pkg/errors"

	"github.com/arangodb/go-driver/jwt"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"

	"github.com/rs/zerolog"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	recordfake "k8s.io/client-go/tools/record"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	arangofake "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
)

const (
	testNamespace                 = "default"
	testDeploymentName            = "test"
	testVersion                   = "3.5.2"
	testImage                     = "arangodb/arangodb:" + testVersion
	testCASecretName              = "testCA"
	testJWTSecretName             = "testJWT"
	testExporterToken             = "testExporterToken"
	testRocksDBEncryptionKey      = "testRocksDB"
	testPersistentVolumeClaimName = "testClaim"
	testLicense                   = "testLicense"
	testServiceAccountName        = "testServiceAccountName"
	testPriorityClassName         = "testPriority"
	testImageLifecycle            = "arangodb/kube-arangodb:0.3.16"
	testExporterImage             = "arangodb/arangodb-exporter:0.1.6"
	testImageAlpine               = "alpine:3.7"

	testYes = "yes"
)

type testCaseStruct struct {
	Name             string
	ArangoDeployment *api.ArangoDeployment
	Helper           func(*testing.T, *Deployment, *testCaseStruct)
	config           Config
	ExpectedError    error
	ExpectedEvent    string
	ExpectedPod      v1.Pod
}

func TestEnsurePods(t *testing.T) {
	// Arange
	defaultAgentTerminationTimeout := int64(api.ServerGroupAgents.DefaultTerminationGracePeriod().Seconds())
	defaultDBServerTerminationTimeout := int64(api.ServerGroupDBServers.DefaultTerminationGracePeriod().Seconds())
	defaultCoordinatorTerminationTimeout := int64(api.ServerGroupCoordinators.DefaultTerminationGracePeriod().Seconds())
	defaultSingleTerminationTimeout := int64(api.ServerGroupSingle.DefaultTerminationGracePeriod().Seconds())
	defaultSyncMasterTerminationTimeout := int64(api.ServerGroupSyncMasters.DefaultTerminationGracePeriod().Seconds())
	defaultSyncWorkerTerminationTimeout := int64(api.ServerGroupSyncWorkers.DefaultTerminationGracePeriod().Seconds())

	var securityContext api.ServerGroupSpecSecurityContext

	nodeSelectorTest := map[string]string{
		"test": "test",
	}

	firstAgentStatus := api.MemberStatus{
		ID:    "agent1",
		Phase: api.MemberPhaseNone,
	}

	firstCoordinatorStatus := api.MemberStatus{
		ID:    "coordinator1",
		Phase: api.MemberPhaseNone,
	}

	singleStatus := api.MemberStatus{
		ID:    "single1",
		Phase: api.MemberPhaseNone,
	}

	firstSyncMaster := api.MemberStatus{
		ID:    "syncMaster1",
		Phase: api.MemberPhaseNone,
	}

	firstSyncWorker := api.MemberStatus{
		ID:    "syncWorker1",
		Phase: api.MemberPhaseNone,
	}

	firstDBServerStatus := api.MemberStatus{
		ID:    "DBserver1",
		Phase: api.MemberPhaseNone,
	}

	noAuthentication := api.AuthenticationSpec{
		JWTSecretName: util.NewString(api.JWTSecretNameDisabled),
	}

	noTLS := api.TLSSpec{
		CASecretName: util.NewString(api.CASecretNameDisabled),
	}

	authenticationSpec := api.AuthenticationSpec{
		JWTSecretName: util.NewString(testJWTSecretName),
	}
	tlsSpec := api.TLSSpec{
		CASecretName: util.NewString(testCASecretName),
	}

	rocksDBSpec := api.RocksDBSpec{
		Encryption: api.RocksDBEncryptionSpec{
			KeySecretName: util.NewString(testRocksDBEncryptionKey),
		},
	}

	metricsSpec := api.MetricsSpec{
		Enabled: util.NewBool(true),
		Image:   util.NewString(testExporterImage),
		Authentication: api.MetricsAuthenticationSpec{
			JWTTokenSecretName: util.NewString(testExporterToken),
		},
	}

	resourcesUnfiltered := v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:     resource.MustParse("500m"),
			v1.ResourceMemory:  resource.MustParse("2Gi"),
			v1.ResourceStorage: resource.MustParse("8Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:     resource.MustParse("100m"),
			v1.ResourceMemory:  resource.MustParse("1Gi"),
			v1.ResourceStorage: resource.MustParse("2Gi"),
		},
	}

	emptyResources := v1.ResourceRequirements{
		Limits:   make(v1.ResourceList),
		Requests: make(v1.ResourceList),
	}

	sidecarName1 := "sidecar1"
	sidecarName2 := "sidecar2"

	testCases := []testCaseStruct{
		{
			Name: "Agent Pod with image pull policy",
			ArangoDeployment: &api.ArangoDeployment{
				Spec: api.DeploymentSpec{
					Image:           util.NewString(testImage),
					Authentication:  noAuthentication,
					TLS:             noTLS,
					ImagePullPolicy: util.NewPullPolicy(v1.PullAlways),
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullAlways,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
						Sidecars: []v1.Container{
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						{
							Name: sidecarName1,
						},
						{
							Name: sidecarName2,
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					ImagePullSecrets: []v1.LocalObjectReference{
						{
							Name: "docker-registry",
						},
						{
							Name: "other-registry",
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
				AlpineImage: testImageAlpine,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					InitContainers: []v1.Container{
						createTestAlpineContainer(firstAgentStatus.ID, false),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeWithPersitantVolumeClaim(k8sutil.ArangodVolumeName,
							testPersistentVolumeClaimName),
					},
					Containers: []v1.Container{
						{
							Name:      k8sutil.ServerContainerName,
							Image:     testImage,
							Command:   createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:     createTestPorts(),
							Resources: k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
				AlpineImage: testImageAlpine,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					InitContainers: []v1.Container{
						createTestAlpineContainer(firstDBServerStatus.ID, true),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeWithPersitantVolumeClaim(k8sutil.ArangodVolumeName,
							testPersistentVolumeClaimName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupAgentsString, firstAgentStatus.ID),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, true, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(true, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, true, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupAgentsString, firstAgentStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []v1.Container{
						{
							Name:            k8sutil.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForAgent(firstAgentStatus.ID, true, true, false),
							Ports:           createTestPorts(),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, testRocksDBEncryptionKey),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, true),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.RocksdbEncryptionVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForAgent(firstAgentStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, emptyResources),
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Ports:   createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
							},
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered)),
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []v1.Container{
						createTestLifecycleContainer(k8sutil.ExtractPodResourceRequirement(resourcesUnfiltered)),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Env: []v1.EnvVar{
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							Ports: createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
							},
							Lifecycle:       createTestLifecycle(),
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, emptyResources),
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
				LifecycleImage: testImageLifecycle,
				AlpineImage:    testImageAlpine,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []v1.Container{
						createTestLifecycleContainer(emptyResources),
						createTestAlpineContainer(firstDBServerStatus.ID, false),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, false, false, false),
							Env: []v1.EnvVar{
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							Ports: createTestPorts(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
							},
							Lifecycle:       createTestLifecycle(),
							LivenessProbe:   createTestLivenessProbe(false, "", k8sutil.ArangoPort),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
						},
						testCreateExporterContainer(false, emptyResources),
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
				testCase.ExpectedPod.Spec.Containers[1].VolumeMounts = append(
					testCase.ExpectedPod.Spec.Containers[1].VolumeMounts, k8sutil.TlsKeyfileVolumeMount())
			},
			config: Config{
				LifecycleImage: testImageLifecycle,
			},
			ExpectedEvent: "member dbserver is created",
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupDBServersString, firstDBServerStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.RocksdbEncryptionVolumeName, testRocksDBEncryptionKey),
						k8sutil.CreateVolumeWithSecret(k8sutil.ExporterJWTVolumeName, testExporterToken),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
						k8sutil.LifecycleVolume(),
					},
					InitContainers: []v1.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForDBServer(firstDBServerStatus.ID, true, true, true),
							Env: []v1.EnvVar{
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
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.LifecycleVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.RocksdbEncryptionVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
						testCreateExporterContainer(true, emptyResources),
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupCoordinatorsString, firstCoordinatorStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []v1.Container{
						{
							Name:            k8sutil.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForCoordinator(firstCoordinatorStatus.ID, true, true, false),
							Ports:           createTestPorts(),
							ImagePullPolicy: v1.PullIfNotPresent,
							Resources:       emptyResources,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
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
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.CreateVolumeEmptyDir(k8sutil.ArangodVolumeName),
						createTestTLSVolume(api.ServerGroupSingleString, singleStatus.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []v1.Container{
						{
							Name:            k8sutil.ServerContainerName,
							Image:           testImage,
							Command:         createTestCommandForSingleMode(singleStatus.ID, true, true, false),
							Ports:           createTestPorts(),
							ImagePullPolicy: v1.PullIfNotPresent,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.ArangodVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSingleTerminationTimeout,
					Hostname:                      testDeploymentName + "-" + api.ServerGroupSingleString + "-" + singleStatus.ID,
					Subdomain:                     testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSingleString,
						false, ""),
				},
			},
		},
		//ArangoD container - end

		// Arango sync master container - start
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
				err := deployment.GetKubeCli().CoreV1().Secrets(testNamespace).Delete(secretName, &metav1.DeleteOptions{})
				require.NoError(t, err)
			},
			ExpectedError: errors.New("Monitoring token secret validation failed: secrets \"" +
				testDeploymentName + "-sync-mt\" not found"),
		},
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
				err := deployment.GetKubeCli().CoreV1().Secrets(testNamespace).Delete(secretName, &metav1.DeleteOptions{})
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
				err := deployment.GetKubeCli().CoreV1().Secrets(testNamespace).Delete(secretName, &metav1.DeleteOptions{})
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
				err := deployment.GetKubeCli().CoreV1().Secrets(testNamespace).Delete(secretName, &metav1.DeleteOptions{})
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
				auth, err := k8sutil.GetTokenSecret(deployment.GetKubeCli().CoreV1().Secrets(testNamespace), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(
					true, "bearer "+auth, k8sutil.ArangoSyncMasterPort)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClientAuthCAVolumeName, "test-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName, "test-sync-jwt"),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClusterJWTSecretVolumeName, testJWTSecretName),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, true, true),
							Ports:   createTestPorts(),
							Env: []v1.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
									testDeploymentName+"-sync-mt", constants.SecretKeyToken),
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							Resources:       resourcesUnfiltered,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClientAuthCACertificateVolumeMount(),
								k8sutil.MasterJWTVolumeMount(),
								k8sutil.ClusterJWTVolumeMount(),
							},
						},
					},
					PriorityClassName:             testPriorityClassName,
					RestartPolicy:                 v1.RestartPolicyNever,
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
				AlpineImage:    testImageAlpine,
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
				auth, err := k8sutil.GetTokenSecret(deployment.GetKubeCli().CoreV1().Secrets(testNamespace), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(
					true, "bearer "+auth, k8sutil.ArangoSyncMasterPort)
			},
			ExpectedEvent: "member syncmaster is created",
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.LifecycleVolume(),
						createTestTLSVolume(api.ServerGroupSyncMastersString, firstSyncMaster.ID),
						k8sutil.CreateVolumeWithSecret(k8sutil.ClientAuthCAVolumeName,
							testDeploymentName+"-sync-client-auth-ca"),
						k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName,
							testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []v1.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncMaster(firstSyncMaster.ID, true, false, true),
							Ports:   createTestPorts(),
							Env: []v1.EnvVar{
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoSyncMonitoringToken,
									testDeploymentName+"-sync-mt", constants.SecretKeyToken),
								k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
									testLicense, constants.SecretKeyToken),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
								k8sutil.CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
							},
							ImagePullPolicy: v1.PullIfNotPresent,
							Lifecycle:       createTestLifecycle(),
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.LifecycleVolumeMount(),
								k8sutil.TlsKeyfileVolumeMount(),
								k8sutil.ClientAuthCACertificateVolumeMount(),
								k8sutil.MasterJWTVolumeMount(),
							},
						},
					},
					RestartPolicy:                 v1.RestartPolicyNever,
					TerminationGracePeriodSeconds: &defaultSyncMasterTerminationTimeout,
					Hostname: testDeploymentName + "-" + api.ServerGroupSyncMastersString + "-" +
						firstSyncMaster.ID,
					Subdomain: testDeploymentName + "-int",
					Affinity: k8sutil.CreateAffinity(testDeploymentName, api.ServerGroupSyncMastersString,
						true, ""),
				},
			},
		},
		// Arango sync master container - end

		// Arango sync worker - start
		{
			Name: "Sync Worker Pod with monitoring, service account, node selector, lifecycle, license " +
				"liveness probe, priority class name, resource requirements without alpine",
			config: Config{
				LifecycleImage: testImageLifecycle,
				AlpineImage:    testImageAlpine,
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
				auth, err := k8sutil.GetTokenSecret(deployment.GetKubeCli().CoreV1().Secrets(testNamespace), name)
				require.NoError(t, err)

				testCase.ExpectedPod.Spec.Containers[0].LivenessProbe = createTestLivenessProbe(
					true, "bearer "+auth, k8sutil.ArangoSyncWorkerPort)
			},
			ExpectedEvent: "member syncworker is created",
			ExpectedPod: v1.Pod{
				Spec: v1.PodSpec{
					Volumes: []v1.Volume{
						k8sutil.LifecycleVolume(),
						k8sutil.CreateVolumeWithSecret(k8sutil.MasterJWTSecretVolumeName, testDeploymentName+"-sync-jwt"),
					},
					InitContainers: []v1.Container{
						createTestLifecycleContainer(emptyResources),
					},
					Containers: []v1.Container{
						{
							Name:    k8sutil.ServerContainerName,
							Image:   testImage,
							Command: createTestCommandForSyncWorker(firstSyncWorker.ID, true, true),
							Ports:   createTestPorts(),
							Env: []v1.EnvVar{
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
							ImagePullPolicy: v1.PullIfNotPresent,
							Resources:       resourcesUnfiltered,
							SecurityContext: securityContext.NewSecurityContext(),
							VolumeMounts: []v1.VolumeMount{
								k8sutil.LifecycleVolumeMount(),
								k8sutil.MasterJWTVolumeMount(),
							},
						},
					},
					PriorityClassName:             testPriorityClassName,
					RestartPolicy:                 v1.RestartPolicyNever,
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
		// Arango sync worker - end
	}

	for _, testCase := range testCases {
		//nolint:scopelint
		t.Run(testCase.Name, func(t *testing.T) {
			// Arrange
			d, eventRecorder := createTestDeployment(testCase.config, testCase.ArangoDeployment)

			err := d.resources.EnsureSecrets()
			require.NoError(t, err)

			if testCase.Helper != nil {
				testCase.Helper(t, d, &testCase)
			}

			// Create custom resource in the fake kubernetes API
			_, err = d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(testNamespace).Create(d.apiObject)
			require.NoError(t, err)

			// Act
			err = d.resources.EnsurePods()

			// Assert
			if testCase.ExpectedError != nil {

				if !assert.EqualError(t, err, testCase.ExpectedError.Error()) {
					println(fmt.Sprintf("%+v", err))
				}
				return
			}

			require.NoError(t, err)
			pods, err := d.deps.KubeCli.CoreV1().Pods(testNamespace).List(metav1.ListOptions{})
			require.NoError(t, err)
			require.Len(t, pods.Items, 1)
			require.Equal(t, testCase.ExpectedPod.Spec, pods.Items[0].Spec)
			require.Equal(t, testCase.ExpectedPod.ObjectMeta, pods.Items[0].ObjectMeta)

			if len(testCase.ExpectedEvent) > 0 {
				select {
				case msg := <-eventRecorder.Events:
					assert.Contains(t, msg, testCase.ExpectedEvent)
				default:
					assert.Fail(t, "expected event", "expected event with message '%s'", testCase.ExpectedEvent)
				}

				status, version := d.GetStatus()
				assert.Equal(t, int32(1), version)

				checkEachMember := func(group api.ServerGroup, groupSpec api.ServerGroupSpec, status *api.MemberStatusList) error {
					for _, m := range *status {
						require.Equal(t, api.MemberPhaseCreated, m.Phase)

						_, exist := m.Conditions.Get(api.ConditionTypeReady)
						require.Equal(t, false, exist)
						_, exist = m.Conditions.Get(api.ConditionTypeTerminated)
						require.Equal(t, false, exist)
						_, exist = m.Conditions.Get(api.ConditionTypeTerminating)
						require.Equal(t, false, exist)
						_, exist = m.Conditions.Get(api.ConditionTypeAgentRecoveryNeeded)
						require.Equal(t, false, exist)
						_, exist = m.Conditions.Get(api.ConditionTypeAutoUpgrade)
						require.Equal(t, false, exist)
					}
					return nil
				}

				d.GetServerGroupIterator().ForeachServerGroup(checkEachMember, &status)
			}
		})
	}
}

func createTestTLSVolume(serverGroupString, ID string) v1.Volume {
	return k8sutil.CreateVolumeWithSecret(k8sutil.TlsKeyfileVolumeName,
		k8sutil.CreateTLSKeyfileSecretName(testDeploymentName, serverGroupString, ID))
}

func createTestLifecycle() *v1.Lifecycle {
	lifecycle, _ := k8sutil.NewLifecycle()
	return lifecycle
}

func createTestToken(deployment *Deployment, testCase *testCaseStruct, paths []string) (string, error) {

	name := testCase.ArangoDeployment.Spec.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(deployment.GetKubeCli().CoreV1().Secrets(testNamespace), name)
	if err != nil {
		return "", err
	}

	return jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(s, "kube-arangodb", paths)
}

func createTestLivenessProbe(secure bool, authorization string, port int) *v1.Probe {
	return k8sutil.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        secure,
		Authorization: authorization,
		Port:          port,
	}.Create()
}

func createTestReadinessProbe(secure bool, authorization string) *v1.Probe {
	return k8sutil.HTTPProbeConfig{
		LocalPath:           "/_admin/server/availability",
		Secure:              secure,
		Authorization:       authorization,
		InitialDelaySeconds: 2,
		PeriodSeconds:       2,
	}.Create()
}

func createTestCommandForDBServer(name string, tls, auth, encryptionRocksDB bool) []string {
	command := []string{resources.ArangoDExecutor}
	if tls {
		command = append(command, "--cluster.my-address=ssl://"+testDeploymentName+"-"+
			api.ServerGroupDBServersString+"-"+name+".test-int.default.svc:8529")
	} else {
		command = append(command, "--cluster.my-address=tcp://"+testDeploymentName+"-"+
			api.ServerGroupDBServersString+"-"+name+".test-int.default.svc:8529")
	}

	command = append(command, "--cluster.my-role=PRIMARY", "--database.directory=/data",
		"--foxx.queues=false", "--log.level=INFO", "--log.output=+")

	if encryptionRocksDB {
		command = append(command, "--rocksdb.encryption-keyfile=/secrets/rocksdb/encryption/key")
	}

	if auth {
		command = append(command, "--server.authentication=true")
	} else {
		command = append(command, "--server.authentication=false")
	}

	if tls {
		command = append(command, "--server.endpoint=ssl://[::]:8529")
	} else {
		command = append(command, "--server.endpoint=tcp://[::]:8529")
	}

	if auth {
		command = append(command, "--server.jwt-secret-keyfile=/secrets/cluster/jwt/token")
	}

	command = append(command, "--server.statistics=true", "--server.storage-engine=rocksdb")

	if tls {
		command = append(command, "--ssl.ecdh-curve=", "--ssl.keyfile=/secrets/tls/tls.keyfile")
	}
	return command
}

func createTestCommandForCoordinator(name string, tls, auth, encryptionRocksDB bool) []string {
	command := []string{resources.ArangoDExecutor}
	if tls {
		command = append(command, "--cluster.my-address=ssl://"+testDeploymentName+"-"+
			api.ServerGroupCoordinatorsString+"-"+name+".test-int.default.svc:8529")
	} else {
		command = append(command, "--cluster.my-address=tcp://"+testDeploymentName+"-"+
			api.ServerGroupCoordinatorsString+"-"+name+".test-int.default.svc:8529")
	}

	command = append(command, "--cluster.my-role=COORDINATOR", "--database.directory=/data",
		"--foxx.queues=true", "--log.level=INFO", "--log.output=+")

	if encryptionRocksDB {
		command = append(command, "--rocksdb.encryption-keyfile=/secrets/rocksdb/encryption/key")
	}

	if auth {
		command = append(command, "--server.authentication=true")
	} else {
		command = append(command, "--server.authentication=false")
	}

	if tls {
		command = append(command, "--server.endpoint=ssl://[::]:8529")
	} else {
		command = append(command, "--server.endpoint=tcp://[::]:8529")
	}

	if auth {
		command = append(command, "--server.jwt-secret-keyfile=/secrets/cluster/jwt/token")
	}

	command = append(command, "--server.statistics=true", "--server.storage-engine=rocksdb")

	if tls {
		command = append(command, "--ssl.ecdh-curve=", "--ssl.keyfile=/secrets/tls/tls.keyfile")
	}
	return command
}

func createTestCommandForSingleMode(name string, tls, auth, encryptionRocksDB bool) []string {
	command := []string{resources.ArangoDExecutor}

	command = append(command, "--database.directory=/data", "--foxx.queues=true", "--log.level=INFO",
		"--log.output=+")

	if encryptionRocksDB {
		command = append(command, "--rocksdb.encryption-keyfile=/secrets/rocksdb/encryption/key")
	}

	if auth {
		command = append(command, "--server.authentication=true")
	} else {
		command = append(command, "--server.authentication=false")
	}

	if tls {
		command = append(command, "--server.endpoint=ssl://[::]:8529")
	} else {
		command = append(command, "--server.endpoint=tcp://[::]:8529")
	}

	if auth {
		command = append(command, "--server.jwt-secret-keyfile=/secrets/cluster/jwt/token")
	}

	command = append(command, "--server.statistics=true", "--server.storage-engine=rocksdb")

	if tls {
		command = append(command, "--ssl.ecdh-curve=", "--ssl.keyfile=/secrets/tls/tls.keyfile")
	}
	return command
}

func createTestCommandForAgent(name string, tls, auth, encryptionRocksDB bool) []string {
	command := []string{
		resources.ArangoDExecutor,
		"--agency.activate=true",
		"--agency.disaster-recovery-id=" + name}

	if tls {
		command = append(command, "--agency.my-address=ssl://"+testDeploymentName+"-"+
			api.ServerGroupAgentsString+"-"+name+"."+testDeploymentName+"-int."+testNamespace+".svc:8529")
	} else {
		command = append(command, "--agency.my-address=tcp://"+testDeploymentName+"-"+
			api.ServerGroupAgentsString+"-"+name+"."+testDeploymentName+"-int."+testNamespace+".svc:8529")
	}

	command = append(command,
		"--agency.size=3",
		"--agency.supervision=true",
		"--database.directory=/data",
		"--foxx.queues=false",
		"--log.level=INFO",
		"--log.output=+")

	if encryptionRocksDB {
		command = append(command, "--rocksdb.encryption-keyfile=/secrets/rocksdb/encryption/key")
	}

	if auth {
		command = append(command, "--server.authentication=true")
	} else {
		command = append(command, "--server.authentication=false")
	}

	if tls {
		command = append(command, "--server.endpoint=ssl://[::]:8529")
	} else {
		command = append(command, "--server.endpoint=tcp://[::]:8529")
	}

	if auth {
		command = append(command, "--server.jwt-secret-keyfile=/secrets/cluster/jwt/token")
	}

	command = append(command, "--server.statistics=false", "--server.storage-engine=rocksdb")

	if tls {
		command = append(command, "--ssl.ecdh-curve=", "--ssl.keyfile=/secrets/tls/tls.keyfile")
	}

	return command
}

func createTestCommandForSyncMaster(name string, tls, auth, monitoring bool) []string {
	command := []string{resources.ArangoSyncExecutor, "run", "master"}

	if tls {
		command = append(command, "--cluster.endpoint=https://"+testDeploymentName+":8529")
	} else {
		command = append(command, "--cluster.endpoint=http://"+testDeploymentName+":8529")
	}

	if auth {
		command = append(command, "--cluster.jwt-secret=/secrets/cluster/jwt/token")
	}

	command = append(command, "--master.endpoint=https://"+testDeploymentName+"-sync.default.svc:8629")

	command = append(command, "--master.jwt-secret=/secrets/master/jwt/token")

	if monitoring {
		command = append(command, "--monitoring.token="+"$("+constants.EnvArangoSyncMonitoringToken+")")
	}

	command = append(command, "--mq.type=direct", "--server.client-cafile=/secrets/client-auth/ca/ca.crt")

	command = append(command, "--server.endpoint=https://"+testDeploymentName+
		"-syncmaster-"+name+".test-int."+testNamespace+".svc:8629",
		"--server.keyfile=/secrets/tls/tls.keyfile", "--server.port=8629")

	return command
}

func createTestCommandForSyncWorker(name string, tls, monitoring bool) []string {
	command := []string{resources.ArangoSyncExecutor, "run", "worker"}

	scheme := "http"
	if tls {
		scheme = "https"
	}

	command = append(command,
		"--master.endpoint=https://"+testDeploymentName+"-sync:8629",
		"--master.jwt-secret=/secrets/master/jwt/token")

	if monitoring {
		command = append(command, "--monitoring.token="+"$("+constants.EnvArangoSyncMonitoringToken+")")
	}

	command = append(command,
		"--server.endpoint="+scheme+"://"+testDeploymentName+"-syncworker-"+name+".test-int."+testNamespace+".svc:8729",
		"--server.port=8729")

	return command
}

func createTestDeployment(config Config, arangoDeployment *api.ArangoDeployment) (*Deployment, *recordfake.FakeRecorder) {

	eventRecorder := recordfake.NewFakeRecorder(10)
	kubernetesClientSet := fake.NewSimpleClientset()

	arangoDeployment.ObjectMeta = metav1.ObjectMeta{
		Name:      testDeploymentName,
		Namespace: testNamespace,
	}

	deps := Dependencies{
		Log:           zerolog.New(ioutil.Discard),
		KubeCli:       kubernetesClientSet,
		DatabaseCRCli: arangofake.NewSimpleClientset(&api.ArangoDeployment{}),
		EventRecorder: eventRecorder,
	}

	d := &Deployment{
		apiObject:   arangoDeployment,
		config:      config,
		deps:        deps,
		eventCh:     make(chan *deploymentEvent, deploymentEventQueueSize),
		stopCh:      make(chan struct{}),
		clientCache: newClientCache(deps.KubeCli, arangoDeployment),
	}

	arangoDeployment.Spec.SetDefaults(arangoDeployment.GetName())
	d.resources = resources.NewResources(deps.Log, d)

	return d, eventRecorder
}

func createTestPorts() []v1.ContainerPort {
	return []v1.ContainerPort{
		{
			Name:          "server",
			ContainerPort: 8529,
			Protocol:      "TCP",
		},
	}
}

func createTestImages(enterprise bool) api.ImageInfoList {
	return api.ImageInfoList{
		{
			Image:           testImage,
			ArangoDBVersion: testVersion,
			ImageID:         testImage,
			Enterprise:      enterprise,
		},
	}
}

func createTestExporterPorts() []v1.ContainerPort {
	return []v1.ContainerPort{
		{
			Name:          "exporter",
			ContainerPort: 9101,
			Protocol:      "TCP",
		},
	}
}

func createTestExporterCommand(secure bool) []string {
	command := []string{
		"/app/arangodb-exporter",
	}

	if secure {
		command = append(command, "--arangodb.endpoint=https://localhost:8529")
	} else {
		command = append(command, "--arangodb.endpoint=http://localhost:8529")
	}

	command = append(command, "--arangodb.jwt-file=/secrets/exporter/jwt/token")

	if secure {
		command = append(command, "--ssl.keyfile=/secrets/tls/tls.keyfile")
	}
	return command
}

func createTestExporterLivenessProbe(secure bool) *v1.Probe {
	return k8sutil.HTTPProbeConfig{
		LocalPath: "/",
		Port:      k8sutil.ArangoExporterPort,
		Secure:    secure,
	}.Create()
}

func createTestLifecycleContainer(resources v1.ResourceRequirements) v1.Container {
	binaryPath, _ := os.Executable()
	var securityContext api.ServerGroupSpecSecurityContext

	return v1.Container{
		Name:    "init-lifecycle",
		Image:   testImageLifecycle,
		Command: []string{binaryPath, "lifecycle", "copy", "--target", "/lifecycle/tools"},
		VolumeMounts: []v1.VolumeMount{
			k8sutil.LifecycleVolumeMount(),
		},
		ImagePullPolicy: "IfNotPresent",
		Resources:       resources,
		SecurityContext: securityContext.NewSecurityContext(),
	}
}

func createTestAlpineContainer(name string, requireUUID bool) v1.Container {
	var securityContext api.ServerGroupSpecSecurityContext
	return k8sutil.ArangodInitContainer("uuid", name, "rocksdb", testImageAlpine, requireUUID, securityContext.NewSecurityContext())
}

func (testCase *testCaseStruct) createTestPodData(deployment *Deployment, group api.ServerGroup,
	memberStatus api.MemberStatus) {

	podName := k8sutil.CreatePodName(testDeploymentName, group.AsRoleAbbreviated(), memberStatus.ID,
		resources.CreatePodSuffix(testCase.ArangoDeployment.Spec))

	testCase.ExpectedPod.ObjectMeta = metav1.ObjectMeta{
		Name:      podName,
		Namespace: testNamespace,
		Labels:    k8sutil.LabelsForDeployment(testDeploymentName, group.AsRole()),
		OwnerReferences: []metav1.OwnerReference{
			testCase.ArangoDeployment.AsOwner(),
		},
		Finalizers: deployment.resources.CreatePodFinalizers(group),
	}

	groupSpec := testCase.ArangoDeployment.Spec.GetServerGroupSpec(group)
	testCase.ExpectedPod.Spec.Tolerations = deployment.resources.CreatePodTolerations(group, groupSpec)
}

func testCreateExporterContainer(secure bool, resources v1.ResourceRequirements) v1.Container {
	var securityContext api.ServerGroupSpecSecurityContext

	return v1.Container{
		Name:    k8sutil.ExporterContainerName,
		Image:   testExporterImage,
		Command: createTestExporterCommand(secure),
		Ports:   createTestExporterPorts(),
		VolumeMounts: []v1.VolumeMount{
			k8sutil.ExporterJWTVolumeMount(),
		},
		Resources:       resources,
		LivenessProbe:   createTestExporterLivenessProbe(secure),
		ImagePullPolicy: v1.PullIfNotPresent,
		SecurityContext: securityContext.NewSecurityContext(),
	}
}
