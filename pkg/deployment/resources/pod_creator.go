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
// Author Ewout Prangsma
//

package resources

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"

	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func versionHasAdvertisedEndpoint(v driver.Version) bool {
	return v.CompareTo("3.4.0") >= 0
}

// versionHasJWTSecretKeyfile derives from the version number of arangod has
// the option --auth.jwt-secret-keyfile which can take the JWT secret from
// a file in the file system.
func versionHasJWTSecretKeyfile(v driver.Version) bool {
	if v.CompareTo("3.3.22") >= 0 && v.CompareTo("3.4.0") < 0 {
		return true
	}
	if v.CompareTo("3.4.2") >= 0 {
		return true
	}

	return false
}

// createArangodArgs creates command line arguments for an arangod server in the given group.
func createArangodArgs(input pod.Input) []string {
	options := k8sutil.CreateOptionPairs(64)

	//scheme := NewURLSchemes(bsCfg.SslKeyFile != "").Arangod
	scheme := "tcp"
	if input.Deployment.IsSecure() {
		scheme = "ssl"
	}

	options.Addf("--server.endpoint", "%s://%s:%d", scheme, input.Deployment.GetListenAddr(), k8sutil.ArangoPort)

	// Authentication
	if input.Deployment.IsAuthenticated() {
		// With authentication
		options.Add("--server.authentication", "true")

		if versionHasJWTSecretKeyfile(input.Version) {
			keyPath := filepath.Join(k8sutil.ClusterJWTSecretVolumeMountDir, constants.SecretKeyToken)
			options.Add("--server.jwt-secret-keyfile", keyPath)
		} else {
			options.Addf("--server.jwt-secret", "$(%s)", constants.EnvArangodJWTSecret)
		}
	} else {
		// Without authentication
		options.Add("--server.authentication", "false")
	}

	// Storage engine
	options.Add("--server.storage-engine", input.Deployment.GetStorageEngine().AsArangoArgument())

	// Logging
	options.Add("--log.level", "INFO")

	// TLS
	if input.Deployment.IsSecure() {
		keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
		options.Add("--ssl.keyfile", keyPath)
		options.Add("--ssl.ecdh-curve", "") // This way arangod accepts curves other than P256 as well.
	}

	// RocksDB
	options.Merge(pod.Encryption().Args(input))

	options.Add("--database.directory", k8sutil.ArangodVolumeMountDir)
	options.Add("--log.output", "+")

	options.Merge(pod.AutoUpgrade().Args(input))
	options.Merge(pod.SNI().Args(input))

	versionHasAdvertisedEndpoint := versionHasAdvertisedEndpoint(input.Version)

	/*	if config.ServerThreads != 0 {
		options = append(options,
			k8sutil.OptionPair{"--server.threads", strconv.Itoa(config.ServerThreads)})
	}*/
	/*if config.DebugCluster {
		options = append(options,
			k8sutil.OptionPair{"--log.level", "startup=trace"})
	}*/
	myTCPURL := scheme + "://" + net.JoinHostPort(k8sutil.CreatePodDNSName(input.ApiObject, input.Group.AsRole(), input.ID), strconv.Itoa(k8sutil.ArangoPort))
	addAgentEndpoints := false
	switch input.Group {
	case api.ServerGroupAgents:
		options.Add("--agency.disaster-recovery-id", input.ID)
		options.Add("--agency.activate", "true")
		options.Add("--agency.my-address", myTCPURL)
		options.Addf("--agency.size", "%d", input.Deployment.Agents.GetCount())
		options.Add("--agency.supervision", "true")
		options.Add("--foxx.queues", "false")
		options.Add("--server.statistics", "false")
		for _, p := range input.Status.Members.Agents {
			if p.ID != input.ID {
				dnsName := k8sutil.CreatePodDNSName(input.ApiObject, api.ServerGroupAgents.AsRole(), p.ID)
				options.Addf("--agency.endpoint", "%s://%s", scheme, net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoPort)))
			}
		}
	case api.ServerGroupDBServers:
		addAgentEndpoints = true
		options.Add("--cluster.my-address", myTCPURL)
		options.Add("--cluster.my-role", "PRIMARY")
		options.Add("--foxx.queues", "false")
		options.Add("--server.statistics", "true")
	case api.ServerGroupCoordinators:
		addAgentEndpoints = true
		options.Add("--cluster.my-address", myTCPURL)
		options.Add("--cluster.my-role", "COORDINATOR")
		options.Add("--foxx.queues", "true")
		options.Add("--server.statistics", "true")
		if input.Deployment.ExternalAccess.HasAdvertisedEndpoint() && versionHasAdvertisedEndpoint {
			options.Add("--cluster.my-advertised-endpoint", input.Deployment.ExternalAccess.GetAdvertisedEndpoint())
		}
	case api.ServerGroupSingle:
		options.Add("--foxx.queues", "true")
		options.Add("--server.statistics", "true")
		if input.Deployment.GetMode() == api.DeploymentModeActiveFailover {
			addAgentEndpoints = true
			options.Add("--replication.automatic-failover", "true")
			options.Add("--cluster.my-address", myTCPURL)
			options.Add("--cluster.my-role", "SINGLE")
			if input.Deployment.ExternalAccess.HasAdvertisedEndpoint() && versionHasAdvertisedEndpoint {
				options.Add("--cluster.my-advertised-endpoint", input.Deployment.ExternalAccess.GetAdvertisedEndpoint())
			}
		}
	}
	if addAgentEndpoints {
		for _, p := range input.Status.Members.Agents {
			dnsName := k8sutil.CreatePodDNSName(input.ApiObject, api.ServerGroupAgents.AsRole(), p.ID)
			options.Addf("--cluster.agency-endpoint", "%s://%s", scheme, net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoPort)))
		}
	}

	args := append(options.Copy().Sort().AsArgs())
	if len(input.GroupSpec.Args) > 0 {
		args = append(args, input.GroupSpec.Args...)
	}

	return args
}

// createArangoSyncArgs creates command line arguments for an arangosync server in the given group.
func createArangoSyncArgs(apiObject metav1.Object, spec api.DeploymentSpec, group api.ServerGroup,
	groupSpec api.ServerGroupSpec, id string) []string {
	options := k8sutil.CreateOptionPairs(64)
	var runCmd string
	var port int

	/*if config.DebugCluster {
		options = append(options,
			k8sutil.OptionPair{"--log.level", "debug"})
	}*/
	if spec.Sync.Monitoring.GetTokenSecretName() != "" {
		options.Addf("--monitoring.token", "$(%s)", constants.EnvArangoSyncMonitoringToken)
	}
	masterSecretPath := filepath.Join(k8sutil.MasterJWTSecretVolumeMountDir, constants.SecretKeyToken)
	options.Add("--master.jwt-secret", masterSecretPath)

	var masterEndpoint []string
	switch group {
	case api.ServerGroupSyncMasters:
		runCmd = "master"
		port = k8sutil.ArangoSyncMasterPort
		masterEndpoint = spec.Sync.ExternalAccess.ResolveMasterEndpoint(k8sutil.CreateSyncMasterClientServiceDNSName(apiObject), port)
		keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
		clientCAPath := filepath.Join(k8sutil.ClientAuthCAVolumeMountDir, constants.SecretCACertificate)
		options.Add("--server.keyfile", keyPath)
		options.Add("--server.client-cafile", clientCAPath)
		options.Add("--mq.type", "direct")
		if spec.IsAuthenticated() {
			clusterSecretPath := filepath.Join(k8sutil.ClusterJWTSecretVolumeMountDir, constants.SecretKeyToken)
			options.Add("--cluster.jwt-secret", clusterSecretPath)
		}
		dbServiceName := k8sutil.CreateDatabaseClientServiceName(apiObject.GetName())
		scheme := "http"
		if spec.IsSecure() {
			scheme = "https"
		}
		options.Addf("--cluster.endpoint", "%s://%s:%d", scheme, dbServiceName, k8sutil.ArangoPort)
	case api.ServerGroupSyncWorkers:
		runCmd = "worker"
		port = k8sutil.ArangoSyncWorkerPort
		masterEndpointHost := k8sutil.CreateSyncMasterClientServiceName(apiObject.GetName())
		masterEndpoint = []string{"https://" + net.JoinHostPort(masterEndpointHost, strconv.Itoa(k8sutil.ArangoSyncMasterPort))}
	}
	for _, ep := range masterEndpoint {
		options.Add("--master.endpoint", ep)
	}
	serverEndpoint := "https://" + net.JoinHostPort(k8sutil.CreatePodDNSName(apiObject, group.AsRole(), id), strconv.Itoa(port))
	options.Add("--server.endpoint", serverEndpoint)
	options.Add("--server.port", strconv.Itoa(port))

	args := []string{
		"run",
		runCmd,
	}

	args = append(args, options.Sort().AsArgs()...)

	if len(groupSpec.Args) > 0 {
		args = append(args, groupSpec.Args...)
	}

	return args
}

// CreatePodFinalizers creates a list of finalizers for a pod created for the given group.
func (r *Resources) CreatePodFinalizers(group api.ServerGroup) []string {
	switch group {
	case api.ServerGroupAgents:
		return []string{constants.FinalizerPodAgencyServing}
	case api.ServerGroupDBServers:
		return []string{constants.FinalizerPodDrainDBServer}
	default:
		return nil
	}
}

// CreatePodTolerations creates a list of tolerations for a pod created for the given group.
func (r *Resources) CreatePodTolerations(group api.ServerGroup, groupSpec api.ServerGroupSpec) []core.Toleration {
	notReadyDur := k8sutil.TolerationDuration{Forever: false, TimeSpan: time.Minute}
	unreachableDur := k8sutil.TolerationDuration{Forever: false, TimeSpan: time.Minute}
	switch group {
	case api.ServerGroupAgents:
		notReadyDur.Forever = true
		unreachableDur.Forever = true
	case api.ServerGroupCoordinators:
		notReadyDur.TimeSpan = 15 * time.Second
		unreachableDur.TimeSpan = 15 * time.Second
	case api.ServerGroupDBServers:
		notReadyDur.TimeSpan = 5 * time.Minute
		unreachableDur.TimeSpan = 5 * time.Minute
	case api.ServerGroupSingle:
		if r.context.GetSpec().GetMode() == api.DeploymentModeSingle {
			notReadyDur.Forever = true
			unreachableDur.Forever = true
		} else {
			notReadyDur.TimeSpan = 5 * time.Minute
			unreachableDur.TimeSpan = 5 * time.Minute
		}
	case api.ServerGroupSyncMasters:
		notReadyDur.TimeSpan = 15 * time.Second
		unreachableDur.TimeSpan = 15 * time.Second
	case api.ServerGroupSyncWorkers:
		notReadyDur.TimeSpan = 1 * time.Minute
		unreachableDur.TimeSpan = 1 * time.Minute
	}
	tolerations := groupSpec.GetTolerations()
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations, k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeNotReady, notReadyDur))
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations, k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeUnreachable, unreachableDur))
	tolerations = k8sutil.AddTolerationIfNotFound(tolerations, k8sutil.NewNoExecuteToleration(k8sutil.TolerationKeyNodeAlphaUnreachable, unreachableDur))
	return tolerations
}

func (r *Resources) RenderPodForMember(cachedStatus inspector.Inspector, spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
	log := r.log
	apiObject := r.context.GetAPIObject()
	m, group, found := status.Members.ElementByID(memberID)
	if !found {
		return nil, maskAny(fmt.Errorf("Member '%s' not found", memberID))
	}
	groupSpec := spec.GetServerGroupSpec(group)

	kubecli := r.context.GetKubeCli()
	ns := r.context.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)

	// Update pod name
	role := group.AsRole()
	roleAbbr := group.AsRoleAbbreviated()

	m.PodName = k8sutil.CreatePodName(apiObject.GetName(), roleAbbr, m.ID, CreatePodSuffix(spec))

	// Render pod
	if group.IsArangod() {
		// Prepare arguments
		version := imageInfo.ArangoDBVersion
		autoUpgrade := m.Conditions.IsTrue(api.ConditionTypeAutoUpgrade)

		tlsKeyfileSecretName := ""
		if spec.IsSecure() {
			tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
		}

		rocksdbEncryptionSecretName := ""
		if spec.RocksDB.IsEncrypted() {
			rocksdbEncryptionSecretName = spec.RocksDB.Encryption.GetKeySecretName()
		}

		var clusterJWTSecretName string
		if spec.IsAuthenticated() {
			if versionHasJWTSecretKeyfile(version) {
				clusterJWTSecretName = spec.Authentication.GetJWTSecretName()
			}
		}

		memberPod := MemberArangoDPod{
			status:                      m,
			tlsKeyfileSecretName:        tlsKeyfileSecretName,
			rocksdbEncryptionSecretName: rocksdbEncryptionSecretName,
			clusterJWTSecretName:        clusterJWTSecretName,
			groupSpec:                   groupSpec,
			spec:                        spec,
			group:                       group,
			resources:                   r,
			imageInfo:                   imageInfo,
			context:                     r.context,
			autoUpgrade:                 autoUpgrade,
			deploymentStatus:            status,
			id:                          memberID,
		}

		input := memberPod.AsInput()

		args := createArangodArgs(input)

		if err := memberPod.Validate(cachedStatus); err != nil {
			return nil, maskAny(errors.Wrapf(err, "Validation of pods resources failed"))
		}

		return RenderArangoPod(apiObject, role, m.ID, m.PodName, args, &memberPod)
	} else if group.IsArangosync() {
		// Check image
		if !imageInfo.Enterprise {
			log.Debug().Str("image", spec.GetImage()).Msg("Image is not an enterprise image")
			return nil, maskAny(fmt.Errorf("Image '%s' does not contain an Enterprise version of ArangoDB", spec.GetImage()))
		}
		// Check if the sync image is overwritten by the SyncSpec
		imageInfo := imageInfo
		if spec.Sync.HasSyncImage() {
			imageInfo.Image = spec.Sync.GetSyncImage()
		}

		var tlsKeyfileSecretName, clientAuthCASecretName, masterJWTSecretName, clusterJWTSecretName string
		// Check master JWT secret
		masterJWTSecretName = spec.Sync.Authentication.GetJWTSecretName()

		if err := k8sutil.ValidateTokenSecret(secrets, masterJWTSecretName); err != nil {
			return nil, maskAny(errors.Wrapf(err, "Master JWT secret validation failed"))
		}

		monitoringTokenSecretName := spec.Sync.Monitoring.GetTokenSecretName()
		if err := k8sutil.ValidateTokenSecret(secrets, monitoringTokenSecretName); err != nil {
			return nil, maskAny(errors.Wrapf(err, "Monitoring token secret validation failed"))
		}

		if group == api.ServerGroupSyncMasters {
			// Create TLS secret
			tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
			// Check cluster JWT secret
			if spec.IsAuthenticated() {
				clusterJWTSecretName = spec.Authentication.GetJWTSecretName()
				if err := k8sutil.ValidateTokenSecret(secrets, clusterJWTSecretName); err != nil {
					return nil, maskAny(errors.Wrapf(err, "Cluster JWT secret validation failed"))
				}
			}
			// Check client-auth CA certificate secret
			clientAuthCASecretName = spec.Sync.Authentication.GetClientCASecretName()
			if err := k8sutil.ValidateCACertificateSecret(secrets, clientAuthCASecretName); err != nil {
				return nil, maskAny(errors.Wrapf(err, "Client authentication CA certificate secret validation failed"))
			}
		}

		// Prepare arguments
		args := createArangoSyncArgs(apiObject, spec, group, groupSpec, m.ID)

		memberSyncPod := MemberSyncPod{
			tlsKeyfileSecretName:   tlsKeyfileSecretName,
			clientAuthCASecretName: clientAuthCASecretName,
			masterJWTSecretName:    masterJWTSecretName,
			clusterJWTSecretName:   clusterJWTSecretName,
			groupSpec:              groupSpec,
			spec:                   spec,
			group:                  group,
			resources:              r,
			imageInfo:              imageInfo,
		}

		return RenderArangoPod(apiObject, role, m.ID, m.PodName, args, &memberSyncPod)
	} else {
		return nil, errors.Errorf("unable to render Pod")
	}
}

func (r *Resources) SelectImage(spec api.DeploymentSpec, status api.DeploymentStatus) (api.ImageInfo, bool) {
	var imageInfo api.ImageInfo
	if current := status.CurrentImage; current != nil {
		// Use current image
		imageInfo = *current
	} else {
		// Find image ID
		info, imageFound := status.Images.GetByImage(spec.GetImage())
		if !imageFound {
			return api.ImageInfo{}, false
		}
		imageInfo = info
		// Save image as current image
		status.CurrentImage = &info
	}
	return imageInfo, true
}

// createPodForMember creates all Pods listed in member status
func (r *Resources) createPodForMember(spec api.DeploymentSpec, memberID string, imageNotFoundOnce *sync.Once, cachedStatus inspector.Inspector) error {
	log := r.log
	status, lastVersion := r.context.GetStatus()

	// Select image
	imageInfo, imageFound := r.SelectImage(spec, status)
	if !imageFound {
		imageNotFoundOnce.Do(func() {
			log.Debug().Str("image", spec.GetImage()).Msg("Image ID is not known yet for image")
		})
		return nil
	}
	status.CurrentImage = &imageInfo

	pod, err := r.RenderPodForMember(cachedStatus, spec, status, memberID, imageInfo)
	if err != nil {
		return maskAny(err)
	}

	kubecli := r.context.GetKubeCli()
	apiObject := r.context.GetAPIObject()
	ns := r.context.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	m, group, found := status.Members.ElementByID(memberID)
	if !found {
		return maskAny(fmt.Errorf("Member '%s' not found", memberID))
	}
	groupSpec := spec.GetServerGroupSpec(group)

	// Update pod name
	role := group.AsRole()
	roleAbbr := group.AsRoleAbbreviated()

	m.PodName = k8sutil.CreatePodName(apiObject.GetName(), roleAbbr, m.ID, CreatePodSuffix(spec))
	newPhase := api.MemberPhaseCreated
	// Create pod
	if group.IsArangod() {
		// Prepare arguments
		autoUpgrade := m.Conditions.IsTrue(api.ConditionTypeAutoUpgrade)
		if autoUpgrade {
			newPhase = api.MemberPhaseUpgrading
		}
		if spec.IsSecure() {
			tlsKeyfileSecretName := k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
			serverNames := []string{
				k8sutil.CreateDatabaseClientServiceDNSName(apiObject),
				k8sutil.CreatePodDNSName(apiObject, role, m.ID),
			}
			if ip := spec.ExternalAccess.GetLoadBalancerIP(); ip != "" {
				serverNames = append(serverNames, ip)
			}
			owner := apiObject.AsOwner()
			if err := createTLSServerCertificate(log, secrets, serverNames, spec.TLS, tlsKeyfileSecretName, &owner); err != nil && !k8sutil.IsAlreadyExists(err) {
				return maskAny(errors.Wrapf(err, "Failed to create TLS keyfile secret"))
			}
		}

		uid, checksum, err := CreateArangoPod(kubecli, apiObject, pod)
		if err != nil {
			return maskAny(err)
		}

		m.PodUID = uid
		m.PodSpecVersion = checksum
		m.ArangoVersion = status.CurrentImage.ArangoDBVersion
		m.ImageID = status.CurrentImage.ImageID

		// Check for missing side cars in
		m.SideCarSpecs = make(map[string]core.Container)
		for _, specSidecar := range groupSpec.GetSidecars() {
			m.SideCarSpecs[specSidecar.Name] = *specSidecar.DeepCopy()
		}

		log.Debug().Str("pod-name", m.PodName).Msg("Created pod")
	} else if group.IsArangosync() {
		// Check monitoring token secret
		if group == api.ServerGroupSyncMasters {
			// Create TLS secret
			tlsKeyfileSecretName := k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
			serverNames := []string{
				k8sutil.CreateSyncMasterClientServiceName(apiObject.GetName()),
				k8sutil.CreateSyncMasterClientServiceDNSName(apiObject),
				k8sutil.CreatePodDNSName(apiObject, role, m.ID),
			}
			masterEndpoint := spec.Sync.ExternalAccess.ResolveMasterEndpoint(k8sutil.CreateSyncMasterClientServiceDNSName(apiObject), k8sutil.ArangoSyncMasterPort)
			for _, ep := range masterEndpoint {
				if u, err := url.Parse(ep); err == nil {
					serverNames = append(serverNames, u.Hostname())
				}
			}
			owner := apiObject.AsOwner()
			if err := createTLSServerCertificate(log, secrets, serverNames, spec.Sync.TLS, tlsKeyfileSecretName, &owner); err != nil && !k8sutil.IsAlreadyExists(err) {
				return maskAny(errors.Wrapf(err, "Failed to create TLS keyfile secret"))
			}
		}

		uid, checksum, err := CreateArangoPod(kubecli, apiObject, pod)
		if err != nil {
			return maskAny(err)
		}
		log.Debug().Str("pod-name", m.PodName).Msg("Created pod")

		m.PodUID = uid
		m.PodSpecVersion = checksum
	}
	// Record new member phase
	m.Phase = newPhase
	m.Conditions.Remove(api.ConditionTypeReady)
	m.Conditions.Remove(api.ConditionTypeTerminated)
	m.Conditions.Remove(api.ConditionTypeTerminating)
	m.Conditions.Remove(api.ConditionTypeAgentRecoveryNeeded)
	m.Conditions.Remove(api.ConditionTypeAutoUpgrade)
	if err := status.Members.Update(m, group); err != nil {
		return maskAny(err)
	}
	if err := r.context.UpdateStatus(status, lastVersion); err != nil {
		return maskAny(err)
	}
	// Create event
	r.context.CreateEvent(k8sutil.NewPodCreatedEvent(m.PodName, role, apiObject))

	return nil
}

// RenderArangoPod renders new ArangoD Pod
func RenderArangoPod(deployment k8sutil.APIObject, role, id, podName string,
	args []string, podCreator interfaces.PodCreator) (*core.Pod, error) {

	// Prepare basic pod
	p := k8sutil.NewPod(deployment.GetName(), role, id, podName, podCreator)

	podCreator.Init(&p)

	if initContainers, err := podCreator.GetInitContainers(); err != nil {
		return nil, maskAny(err)
	} else if initContainers != nil {
		p.Spec.InitContainers = append(p.Spec.InitContainers, initContainers...)
	}

	c, err := k8sutil.NewContainer(args, podCreator.GetContainerCreator())
	if err != nil {
		return nil, maskAny(err)
	}

	p.Spec.Volumes, c.VolumeMounts = podCreator.GetVolumes()
	p.Spec.Containers = append(p.Spec.Containers, c)
	podCreator.GetSidecars(&p)

	// Add affinity
	p.Spec.Affinity = &core.Affinity{
		NodeAffinity:    podCreator.GetNodeAffinity(),
		PodAntiAffinity: podCreator.GetPodAntiAffinity(),
		PodAffinity:     podCreator.GetPodAffinity(),
	}

	return &p, nil
}

// CreateArangoPod creates a new Pod with container provided by parameter 'containerCreator'
// If the pod already exists, nil is returned.
// If another error occurs, that error is returned.
func CreateArangoPod(kubecli kubernetes.Interface, deployment k8sutil.APIObject, pod *core.Pod) (types.UID, string, error) {
	uid, checksum, err := k8sutil.CreatePod(kubecli, pod, deployment.GetNamespace(), deployment.AsOwner())
	if err != nil {
		return "", "", maskAny(err)
	}
	return uid, checksum, nil
}

// EnsurePods creates all Pods listed in member status
func (r *Resources) EnsurePods(cachedStatus inspector.Inspector) error {
	iterator := r.context.GetServerGroupIterator()
	deploymentStatus, _ := r.context.GetStatus()
	imageNotFoundOnce := &sync.Once{}

	createPodMember := func(group api.ServerGroup, groupSpec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.Phase != api.MemberPhaseNone {
				continue
			}
			if m.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
				continue
			}
			spec := r.context.GetSpec()
			if err := r.createPodForMember(spec, m.ID, imageNotFoundOnce, cachedStatus); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}

	if err := iterator.ForeachServerGroup(createPodMember, &deploymentStatus); err != nil {
		return maskAny(err)
	}

	return nil
}

func CreatePodSuffix(spec api.DeploymentSpec) string {
	raw, _ := json.Marshal(spec)
	hash := sha1.Sum(raw)
	return fmt.Sprintf("%0x", hash)[:6]
}
