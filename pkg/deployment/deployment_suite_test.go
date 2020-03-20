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
	"io/ioutil"
	"os"
	"testing"

	"github.com/arangodb/go-driver/jwt"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	arangofake "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	recordfake "k8s.io/client-go/tools/record"
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
	CompareChecksum  *bool
	ExpectedError    error
	ExpectedEvent    string
	ExpectedPod      core.Pod
}

func createTestTLSVolume(serverGroupString, ID string) core.Volume {
	return k8sutil.CreateVolumeWithSecret(k8sutil.TlsKeyfileVolumeName,
		k8sutil.CreateTLSKeyfileSecretName(testDeploymentName, serverGroupString, ID))
}

func createTestLifecycle() *core.Lifecycle {
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

func modTestLivenessProbe(secure bool, authorization string, port int, mod func(*core.Probe)) *core.Probe {
	probe := createTestLivenessProbe(secure, authorization, port)

	mod(probe)

	return probe
}

func createTestReadinessSimpleProbe(secure bool, authorization string, port int) *core.Probe {
	probe := createTestLivenessProbe(secure, authorization, port)

	probe.InitialDelaySeconds = 15
	probe.PeriodSeconds = 10

	return probe
}

func createTestLivenessProbe(secure bool, authorization string, port int) *core.Probe {
	return k8sutil.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        secure,
		Authorization: authorization,
		Port:          port,
	}.Create()
}

func createTestReadinessProbe(secure bool, authorization string) *core.Probe {
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

func createTestPorts() []core.ContainerPort {
	return []core.ContainerPort{
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

func createTestExporterPorts() []core.ContainerPort {
	return []core.ContainerPort{
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

func createTestExporterLivenessProbe(secure bool) *core.Probe {
	return k8sutil.HTTPProbeConfig{
		LocalPath: "/",
		Port:      k8sutil.ArangoExporterPort,
		Secure:    secure,
	}.Create()
}

func createTestLifecycleContainer(resources core.ResourceRequirements) core.Container {
	binaryPath, _ := os.Executable()
	var securityContext api.ServerGroupSpecSecurityContext

	return core.Container{
		Name:    "init-lifecycle",
		Image:   testImageLifecycle,
		Command: []string{binaryPath, "lifecycle", "copy", "--target", "/lifecycle/tools"},
		VolumeMounts: []core.VolumeMount{
			k8sutil.LifecycleVolumeMount(),
		},
		ImagePullPolicy: "IfNotPresent",
		Resources:       resources,
		SecurityContext: securityContext.NewSecurityContext(),
	}
}

func createTestAlpineContainer(name string, requireUUID bool) core.Container {
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

func testCreateExporterContainer(secure bool, resources core.ResourceRequirements) core.Container {
	var securityContext api.ServerGroupSpecSecurityContext

	return core.Container{
		Name:    k8sutil.ExporterContainerName,
		Image:   testExporterImage,
		Command: createTestExporterCommand(secure),
		Ports:   createTestExporterPorts(),
		VolumeMounts: []core.VolumeMount{
			k8sutil.ExporterJWTVolumeMount(),
		},
		Resources:       k8sutil.ExtractPodResourceRequirement(resources),
		LivenessProbe:   createTestExporterLivenessProbe(secure),
		ImagePullPolicy: core.PullIfNotPresent,
		SecurityContext: securityContext.NewSecurityContext(),
	}
}
