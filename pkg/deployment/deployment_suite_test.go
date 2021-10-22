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
// Author Adam Janikowski
// Author Tomasz Mielech
//

package deployment

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/deployment/client"

	monitoringFakeClient "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/fake"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/probes"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod/conn"

	"github.com/arangodb/go-driver"

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
	testVersion                   = "3.7.0"
	testImage                     = "arangodb/arangodb:" + testVersion
	testCASecretName              = "testCA"
	testJWTSecretName             = "testJWT"
	testExporterToken             = "testExporterToken"
	testRocksDBEncryptionKey      = "testRocksDB"
	testPersistentVolumeClaimName = "testClaim"
	testLicense                   = "testLicense"
	testServiceAccountName        = "testServiceAccountName"
	testPriorityClassName         = "testPriority"
	testImageOperator             = "arangodb/kube-arangodb:0.3.16"

	testYes = "yes"
)

type testCaseFeatures struct {
	TLSSNI, TLSRotation, JWTRotation, EncryptionRotation bool
	Graceful *bool
}

type testCaseStruct struct {
	Name             string
	ArangoDeployment *api.ArangoDeployment
	Helper           func(*testing.T, *Deployment, *testCaseStruct)
	Resources        func(*testing.T, *Deployment)
	config           Config
	CompareChecksum  *bool
	ExpectedError    error
	ExpectedEvent    string
	ExpectedPod      core.Pod
	Features         testCaseFeatures
	DropInit         bool
}

func createTestTLSVolume(serverGroupString, ID string) core.Volume {
	return k8sutil.CreateVolumeWithSecret(k8sutil.TlsKeyfileVolumeName,
		k8sutil.CreateTLSKeyfileSecretName(testDeploymentName, serverGroupString, ID))
}

func createTestLifecycle(group api.ServerGroup) *core.Lifecycle {
	if group.IsArangosync() {
		lifecycle, _ := k8sutil.NewLifecycleFinalizers()
		return lifecycle
	}
	lifecycle, _ := k8sutil.NewLifecyclePort()
	return lifecycle
}

func createTestToken(deployment *Deployment, testCase *testCaseStruct, paths []string) (string, error) {

	name := testCase.ArangoDeployment.Spec.Authentication.GetJWTSecretName()
	s, err := k8sutil.GetTokenSecret(context.Background(), deployment.GetCachedStatus().SecretReadInterface(), name)
	if err != nil {
		return "", err
	}

	return jwt.CreateArangodJwtAuthorizationHeaderAllowedPaths(s, "kube-arangodb", paths)
}

func modTestLivenessProbe(mode string, secure bool, authorization string, port int, mod func(*core.Probe)) *core.Probe {
	probe := createTestLivenessProbe(mode, secure, authorization, port)

	mod(probe)

	return probe
}

func createTestReadinessSimpleProbe(mode string, secure bool, authorization string) *core.Probe {
	probe := createTestReadinessProbe(mode, secure, authorization)

	probe.InitialDelaySeconds = 15
	probe.PeriodSeconds = 10

	return probe
}

func createTestLivenessProbe(mode string, secure bool, authorization string, port int) *core.Probe {
	return getProbeCreator(mode)(secure, authorization, "/_api/version", port).Create()
}

func createTestReadinessProbe(mode string, secure bool, authorization string) *core.Probe {
	p := getProbeCreator(mode)(secure, authorization, "/_admin/server/availability", k8sutil.ArangoPort).Create()

	p.InitialDelaySeconds = 2
	p.PeriodSeconds = 2

	return p
}

type probeCreator func(secure bool, authorization, endpoint string, port int) resources.Probe

const (
	cmdProbe  = "cmdProbe"
	httpProbe = "http"
)

func getProbeCreator(t string) probeCreator {
	switch t {
	case cmdProbe:
		return getCMDProbeCreator()
	default:
		return getHTTPProbeCreator()
	}
}

func getHTTPProbeCreator() probeCreator {
	return func(secure bool, authorization, endpoint string, port int) resources.Probe {
		return createHTTPTestProbe(secure, authorization, endpoint, port)
	}
}

func getCMDProbeCreator() probeCreator {
	return func(secure bool, authorization, endpoint string, port int) resources.Probe {
		return createCMDTestProbe(secure, authorization != "", endpoint)
	}
}

func createCMDTestProbe(secure, authorization bool, endpoint string) resources.Probe {
	bin, _ := os.Executable()
	args := []string{
		filepath.Join(k8sutil.LifecycleVolumeMountDir, filepath.Base(bin)),
		"lifecycle",
		"probe",
		fmt.Sprintf("--endpoint=%s", endpoint),
	}

	if secure {
		args = append(args, "--ssl")
	}

	if authorization {
		args = append(args, "--auth")
	}

	return &probes.CMDProbeConfig{
		Command: args,
	}
}

func createHTTPTestProbe(secure bool, authorization string, endpoint string, port int) resources.Probe {
	return &probes.HTTPProbeConfig{
		LocalPath:     endpoint,
		Secure:        secure,
		Authorization: authorization,
		Port:          port,
	}
}

func createTestCommandForDBServer(name string, tls, auth, encryptionRocksDB bool, mods ...func() k8sutil.OptionPairs) []string {
	command := []string{resources.ArangoDExecutor}

	args := k8sutil.OptionPairs{}

	if tls {
		args.Addf("--cluster.my-address", "ssl://%s-%s-%s.test-int.default.svc:8529",
			testDeploymentName,
			api.ServerGroupDBServersString,
			name)
	} else {
		args.Addf("--cluster.my-address", "tcp://%s-%s-%s.test-int.default.svc:8529",
			testDeploymentName,
			api.ServerGroupDBServersString,
			name)
	}

	args.Add("--cluster.my-role", "PRIMARY")
	args.Add("--database.directory", "/data")
	args.Add("--foxx.queues", "false")
	args.Add("--log.level", "INFO")
	args.Add("--log.output", "+")

	if encryptionRocksDB {
		args.Add("--rocksdb.encryption-keyfile", "/secrets/rocksdb/encryption/key")
	}

	args.Add("--server.authentication", auth)

	if tls {
		args.Add("--server.endpoint", "ssl://[::]:8529")
	} else {
		args.Add("--server.endpoint", "tcp://[::]:8529")
	}

	if auth {
		args.Add("--server.jwt-secret-keyfile", "/secrets/cluster/jwt/token")
	}

	args.Add("--server.statistics", "true")
	args.Add("--server.storage-engine", "rocksdb")

	if tls {
		args.Add("--ssl.ecdh-curve", "")
		args.Add("--ssl.keyfile", "/secrets/tls/tls.keyfile")
	}

	for _, mod := range mods {
		args.Merge(mod())
	}

	return append(command, args.Unique().AsArgs()...)
}

func createTestCommandForCoordinator(name string, tls, auth bool, mods ...func() k8sutil.OptionPairs) []string {
	command := []string{resources.ArangoDExecutor}

	args := k8sutil.OptionPairs{}

	if tls {
		args.Addf("--cluster.my-address", "ssl://%s-%s-%s.test-int.default.svc:8529",
			testDeploymentName,
			api.ServerGroupCoordinatorsString,
			name)
	} else {
		args.Addf("--cluster.my-address", "tcp://%s-%s-%s.test-int.default.svc:8529",
			testDeploymentName,
			api.ServerGroupCoordinatorsString,
			name)
	}

	args.Add("--cluster.my-role", "COORDINATOR")
	args.Add("--database.directory", "/data")
	args.Add("--foxx.queues", "true")
	args.Add("--log.level", "INFO")
	args.Add("--log.output", "+")
	args.Add("--server.authentication", auth)

	if tls {
		args.Add("--server.endpoint", "ssl://[::]:8529")
	} else {
		args.Add("--server.endpoint", "tcp://[::]:8529")
	}

	if auth {
		args.Add("--server.jwt-secret-keyfile", "/secrets/cluster/jwt/token")
	}

	args.Add("--server.statistics", "true")
	args.Add("--server.storage-engine", "rocksdb")

	if tls {
		args.Add("--ssl.ecdh-curve", "")
		args.Add("--ssl.keyfile", "/secrets/tls/tls.keyfile")
	}

	for _, mod := range mods {
		args.Merge(mod())
	}

	return append(command, args.Unique().AsArgs()...)
}

func createTestCommandForSingleMode(tls, auth bool, mods ...func() k8sutil.OptionPairs) []string {
	command := []string{resources.ArangoDExecutor}

	args := k8sutil.OptionPairs{}

	args.Add("--database.directory", "/data")
	args.Add("--foxx.queues", "true")
	args.Add("--log.level", "INFO")
	args.Add("--log.output", "+")
	args.Add("--server.authentication", auth)

	if tls {
		args.Add("--server.endpoint", "ssl://[::]:8529")
	} else {
		args.Add("--server.endpoint", "tcp://[::]:8529")
	}

	if auth {
		args.Add("--server.jwt-secret-keyfile", "/secrets/cluster/jwt/token")
	}

	args.Add("--server.statistics", "true")
	args.Add("--server.storage-engine", "rocksdb")

	if tls {
		args.Add("--ssl.ecdh-curve", "")
		args.Add("--ssl.keyfile", "/secrets/tls/tls.keyfile")
	}

	for _, mod := range mods {
		args.Merge(mod())
	}

	return append(command, args.Unique().AsArgs()...)
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

func createTestDeployment(t *testing.T, config Config, arangoDeployment *api.ArangoDeployment) (*Deployment, *recordfake.FakeRecorder) {

	eventRecorder := recordfake.NewFakeRecorder(10)
	kubernetesClientSet := fake.NewSimpleClientset()
	monitoringClientSet := monitoringFakeClient.NewSimpleClientset()

	arangoDeployment.ObjectMeta = metav1.ObjectMeta{
		Name:      testDeploymentName,
		Namespace: testNamespace,
	}

	arangoDeployment.Status.Images = api.ImageInfoList{
		{
			Image:           "arangodb/arangodb:latest",
			ImageID:         "arangodb/arangodb:latest",
			ArangoDBVersion: "1.0.0",
			Enterprise:      false,
		},
	}

	arangoDeployment.Status.CurrentImage = &arangoDeployment.Status.Images[0]

	deps := Dependencies{
		Log:               zerolog.New(ioutil.Discard),
		KubeCli:           kubernetesClientSet,
		KubeMonitoringCli: monitoringClientSet.MonitoringV1(),
		DatabaseCRCli:     arangofake.NewSimpleClientset(&api.ArangoDeployment{}),
		EventRecorder:     eventRecorder,
	}

	d := &Deployment{
		apiObject: arangoDeployment,
		name:      arangoDeployment.GetName(),
		namespace: arangoDeployment.GetNamespace(),
		config:    config,
		deps:      deps,
		eventCh:   make(chan *deploymentEvent, deploymentEventQueueSize),
		stopCh:    make(chan struct{}),
	}
	d.clientCache = client.NewClientCache(d.getArangoDeployment, conn.NewFactory(d.getAuth, d.getConnConfig))

	cachedStatus, err := inspector.NewInspector(context.Background(), d.getKubeCli(), d.getMonitoringV1Cli(), d.getArangoCli(), d.GetNamespace())
	require.NoError(t, err)
	d.SetCachedStatus(cachedStatus)

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

func createTestImagesWithVersion(enterprise bool, version driver.Version) api.ImageInfoList {
	return api.ImageInfoList{
		{
			Image:           testImage,
			ArangoDBVersion: version,
			ImageID:         testImage,
			Enterprise:      enterprise,
		},
	}
}

func createTestImages(enterprise bool) api.ImageInfoList {
	return createTestImagesWithVersion(enterprise, testVersion)
}

func createTestExporterLivenessProbe(secure bool) *core.Probe {
	return probes.HTTPProbeConfig{
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
		Image:   testImageOperator,
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
	binaryPath, _ := os.Executable()
	var securityContext api.ServerGroupSpecSecurityContext
	return k8sutil.ArangodInitContainer("uuid", name, "rocksdb", binaryPath, testImageOperator, requireUUID, securityContext.NewSecurityContext())
}

func (testCase *testCaseStruct) createTestPodData(deployment *Deployment, group api.ServerGroup,
	memberStatus api.MemberStatus) {

	podName := k8sutil.CreatePodName(testDeploymentName, group.AsRoleAbbreviated(), memberStatus.ID,
		resources.CreatePodSuffix(testCase.ArangoDeployment.Spec))

	testCase.ExpectedPod.ObjectMeta = metav1.ObjectMeta{
		Name:      podName,
		Namespace: testNamespace,
		Labels:    k8sutil.LabelsForMember(testDeploymentName, group.AsRole(), memberStatus.ID),
		OwnerReferences: []metav1.OwnerReference{
			testCase.ArangoDeployment.AsOwner(),
		},
		Finalizers: finalizers(group),
	}

	groupSpec := testCase.ArangoDeployment.Spec.GetServerGroupSpec(group)
	testCase.ExpectedPod.Spec.Tolerations = deployment.resources.CreatePodTolerations(group, groupSpec)

	// Add image info
	if member, group, ok := deployment.apiObject.Status.Members.ElementByID(memberStatus.ID); ok {
		member.Image = deployment.apiObject.Status.CurrentImage

		deployment.apiObject.Status.Members.Update(member, group)
	}
}

func finalizers(group api.ServerGroup) []string {
	var finalizers []string
	switch group {
	case api.ServerGroupAgents:
		finalizers = append(finalizers, constants.FinalizerPodGracefulShutdown)
	case api.ServerGroupCoordinators:
		finalizers = append(finalizers, constants.FinalizerDelayPodTermination)
		finalizers = append(finalizers, constants.FinalizerPodGracefulShutdown)
	case api.ServerGroupDBServers:
		finalizers = append(finalizers, constants.FinalizerPodGracefulShutdown)
	case api.ServerGroupSingle:
		finalizers = append(finalizers, constants.FinalizerPodGracefulShutdown)
	}

	return finalizers
}

func defaultPodAppender(t *testing.T, pod *core.Pod, f ...func(t *testing.T, p *core.Pod)) *core.Pod {
	n := pod.DeepCopy()

	for _, a := range f {
		a(t, n)
	}

	return n
}

func podDataSort() func(t *testing.T, p *core.Pod) {
	sortVolumes := map[string]int{
		"rocksdb-encryption": -1,
		"cluster-jwt":        1,
		"tls-keyfile":        -2,
		"arangod-data":       -3,
		"exporter-jwt":       0,
		"lifecycle":          2,
		"uuid":               3,
		"volume":             40,
		"volume2":            40,
	}
	sortVolumeMounts := map[string]int{
		"tls-keyfile":        1,
		"arangod-data":       -1,
		"lifecycle":          0,
		"cluster-jwt":        5,
		"rocksdb-encryption": 4,
		"volume":             40,
		"volume2":            40,
	}
	sortInitContainers := map[string]int{
		"init-lifecycle": 0,
		"uuid":           1,
	}

	return func(t *testing.T, p *core.Pod) {
		sort.Slice(p.Spec.Volumes, func(i, j int) bool {
			av, ak := sortVolumes[p.Spec.Volumes[i].Name]
			if strings.HasPrefix(p.Spec.Volumes[i].Name, "sni-") {
				av = 100
				ak = true
			}
			bv, bk := sortVolumes[p.Spec.Volumes[j].Name]
			if strings.HasPrefix(p.Spec.Volumes[j].Name, "sni-") {
				bv = 100
				bk = true
			}

			if !ak && !bk {
				return false
			}

			if !ak {
				return true
			}

			if !bk {
				return false
			}

			return av < bv
		})

		if len(p.Spec.Containers) > 0 {
			sort.Slice(p.Spec.Containers[0].VolumeMounts, func(i, j int) bool {
				av, ak := sortVolumeMounts[p.Spec.Containers[0].VolumeMounts[i].Name]
				if strings.HasPrefix(p.Spec.Containers[0].VolumeMounts[i].Name, "sni-") {
					av = 100
					ak = true
				}
				bv, bk := sortVolumeMounts[p.Spec.Containers[0].VolumeMounts[j].Name]
				if strings.HasPrefix(p.Spec.Containers[0].VolumeMounts[j].Name, "sni-") {
					bv = 100
					bk = true
				}

				if !ak && !bk {
					return false
				}

				if !ak {
					return true
				}

				if !bk {
					return false
				}

				return av < bv
			})
		}

		sort.Slice(p.Spec.InitContainers, func(i, j int) bool {
			av, ak := sortInitContainers[p.Spec.InitContainers[i].Name]
			bv, bk := sortInitContainers[p.Spec.InitContainers[j].Name]

			if !ak && !bk {
				return false
			}

			if !ak {
				return true
			}

			if !bk {
				return false
			}

			return av < bv
		})
	}
}

func addLifecycle(name string, uuidRequired bool, license string, group api.ServerGroup) func(t *testing.T, p *core.Pod) {
	return func(t *testing.T, p *core.Pod) {
		if group.IsArangosync() {

			return
		}

		if len(p.Spec.Containers) > 0 {
			p.Spec.Containers[0].Env = append(k8sutil.GetLifecycleEnv(), p.Spec.Containers[0].Env...)
			if license != "" {
				p.Spec.Containers[0].Env = append([]core.EnvVar{
					k8sutil.CreateEnvSecretKeySelector(constants.EnvArangoLicenseKey,
						license, constants.SecretKeyToken)}, p.Spec.Containers[0].Env...)
			}
		}

		if _, ok := k8sutil.GetAnyVolumeByName(p.Spec.Volumes, k8sutil.LifecycleVolumeName); !ok {
			p.Spec.Volumes = append([]core.Volume{k8sutil.LifecycleVolume()}, p.Spec.Volumes...)
		}
		if _, ok := k8sutil.GetAnyVolumeByName(p.Spec.Volumes, "arangod-data"); !ok {
			p.Spec.Volumes = append([]core.Volume{k8sutil.LifecycleVolume()}, p.Spec.Volumes...)
		}

		if len(p.Spec.Containers) > 0 {
			p.Spec.Containers[0].Lifecycle = createTestLifecycle(api.ServerGroupAgents)
		}

		if len(p.Spec.Containers) > 0 {
			if _, ok := k8sutil.GetAnyVolumeMountByName(p.Spec.Containers[0].VolumeMounts, "lifecycle"); !ok {
				p.Spec.Containers[0].VolumeMounts = append(p.Spec.Containers[0].VolumeMounts, k8sutil.LifecycleVolumeMount())
			}

			if _, ok := k8sutil.GetAnyContainerByName(p.Spec.InitContainers, "init-lifecycle"); !ok {
				p.Spec.InitContainers = append([]core.Container{createTestLifecycleContainer(emptyResources)}, p.Spec.InitContainers...)

			}
		}

		if _, ok := k8sutil.GetAnyContainerByName(p.Spec.InitContainers, "uuid"); !ok {
			binaryPath, _ := os.Executable()
			p.Spec.InitContainers = append([]core.Container{k8sutil.ArangodInitContainer("uuid", name, "rocksdb", binaryPath, testImageOperator, uuidRequired, securityContext.NewSecurityContext())}, p.Spec.InitContainers...)

		}
	}
}
