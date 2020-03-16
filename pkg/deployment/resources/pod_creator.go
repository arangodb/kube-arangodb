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
	"sort"
	"strconv"
	"sync"
	"time"

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
func createArangodArgs(apiObject metav1.Object, deplSpec api.DeploymentSpec, group api.ServerGroup,
	agents api.MemberStatusList, id string, version driver.Version, autoUpgrade bool) []string {
	options := make([]pod.OptionPair, 0, 64)
	svrSpec := deplSpec.GetServerGroupSpec(group)

	i := pod.Input{
		Deployment:  deplSpec,
		Group:       group,
		GroupSpec:   svrSpec,
		Version:     version,
		AutoUpgrade: autoUpgrade,
	}

	//scheme := NewURLSchemes(bsCfg.SslKeyFile != "").Arangod
	scheme := "tcp"
	if deplSpec.IsSecure() {
		scheme = "ssl"
	}
	options = append(options,
		pod.OptionPair{"--server.endpoint", fmt.Sprintf("%s://%s:%d", scheme, deplSpec.GetListenAddr(), k8sutil.ArangoPort)},
	)

	// Authentication
	if deplSpec.IsAuthenticated() {
		// With authentication
		options = append(options,
			pod.OptionPair{"--server.authentication", "true"},
		)
		if versionHasJWTSecretKeyfile(version) {
			keyPath := filepath.Join(k8sutil.ClusterJWTSecretVolumeMountDir, constants.SecretKeyToken)
			options = append(options,
				pod.OptionPair{"--server.jwt-secret-keyfile", keyPath},
			)
		} else {
			options = append(options,
				pod.OptionPair{"--server.jwt-secret", "$(" + constants.EnvArangodJWTSecret + ")"},
			)
		}
	} else {
		// Without authentication
		options = append(options,
			pod.OptionPair{"--server.authentication", "false"},
		)
	}

	// Storage engine
	options = append(options,
		pod.OptionPair{"--server.storage-engine", deplSpec.GetStorageEngine().AsArangoArgument()},
	)

	// Logging
	options = append(options,
		pod.OptionPair{"--log.level", "INFO"},
	)

	// TLS
	if deplSpec.IsSecure() {
		keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
		options = append(options,
			pod.OptionPair{"--ssl.keyfile", keyPath},
			pod.OptionPair{"--ssl.ecdh-curve", ""}, // This way arangod accepts curves other than P256 as well.
		)
		/*if bsCfg.SslKeyFile != "" {
			if bsCfg.SslCAFile != "" {
				sslSection.Settings["cafile"] = bsCfg.SslCAFile
			}
			config = append(config, sslSection)
		}*/
	}

	// RocksDB
	if deplSpec.RocksDB.IsEncrypted() {
		keyPath := filepath.Join(k8sutil.RocksDBEncryptionVolumeMountDir, constants.SecretEncryptionKey)
		options = append(options,
			pod.OptionPair{"--rocksdb.encryption-keyfile", keyPath},
		)
	}

	options = append(options,
		pod.OptionPair{"--database.directory", k8sutil.ArangodVolumeMountDir},
		pod.OptionPair{"--log.output", "+"},
	)

	options = append(options, pod.AutoUpgrade().Create(i)...)

	versionHasAdvertisedEndpoint := versionHasAdvertisedEndpoint(version)

	/*	if config.ServerThreads != 0 {
		options = append(options,
			pod.OptionPair{"--server.threads", strconv.Itoa(config.ServerThreads)})
	}*/
	/*if config.DebugCluster {
		options = append(options,
			pod.OptionPair{"--log.level", "startup=trace"})
	}*/
	myTCPURL := scheme + "://" + net.JoinHostPort(k8sutil.CreatePodDNSName(apiObject, group.AsRole(), id), strconv.Itoa(k8sutil.ArangoPort))
	addAgentEndpoints := false
	switch group {
	case api.ServerGroupAgents:
		options = append(options,
			pod.OptionPair{"--agency.disaster-recovery-id", id},
			pod.OptionPair{"--agency.activate", "true"},
			pod.OptionPair{"--agency.my-address", myTCPURL},
			pod.OptionPair{"--agency.size", strconv.Itoa(deplSpec.Agents.GetCount())},
			pod.OptionPair{"--agency.supervision", "true"},
			pod.OptionPair{"--foxx.queues", "false"},
			pod.OptionPair{"--server.statistics", "false"},
		)
		for _, p := range agents {
			if p.ID != id {
				dnsName := k8sutil.CreatePodDNSName(apiObject, api.ServerGroupAgents.AsRole(), p.ID)
				options = append(options,
					pod.OptionPair{"--agency.endpoint", fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoPort)))},
				)
			}
		}
	case api.ServerGroupDBServers:
		addAgentEndpoints = true
		options = append(options,
			pod.OptionPair{"--cluster.my-address", myTCPURL},
			pod.OptionPair{"--cluster.my-role", "PRIMARY"},
			pod.OptionPair{"--foxx.queues", "false"},
			pod.OptionPair{"--server.statistics", "true"},
		)
	case api.ServerGroupCoordinators:
		addAgentEndpoints = true
		options = append(options,
			pod.OptionPair{"--cluster.my-address", myTCPURL},
			pod.OptionPair{"--cluster.my-role", "COORDINATOR"},
			pod.OptionPair{"--foxx.queues", "true"},
			pod.OptionPair{"--server.statistics", "true"},
		)
		if deplSpec.ExternalAccess.HasAdvertisedEndpoint() && versionHasAdvertisedEndpoint {
			options = append(options,
				pod.OptionPair{"--cluster.my-advertised-endpoint", deplSpec.ExternalAccess.GetAdvertisedEndpoint()},
			)
		}
	case api.ServerGroupSingle:
		options = append(options,
			pod.OptionPair{"--foxx.queues", "true"},
			pod.OptionPair{"--server.statistics", "true"},
		)
		if deplSpec.GetMode() == api.DeploymentModeActiveFailover {
			addAgentEndpoints = true
			options = append(options,
				pod.OptionPair{"--replication.automatic-failover", "true"},
				pod.OptionPair{"--cluster.my-address", myTCPURL},
				pod.OptionPair{"--cluster.my-role", "SINGLE"},
			)
			if deplSpec.ExternalAccess.HasAdvertisedEndpoint() && versionHasAdvertisedEndpoint {
				options = append(options,
					pod.OptionPair{"--cluster.my-advertised-endpoint", deplSpec.ExternalAccess.GetAdvertisedEndpoint()},
				)
			}
		}
	}
	if addAgentEndpoints {
		for _, p := range agents {
			dnsName := k8sutil.CreatePodDNSName(apiObject, api.ServerGroupAgents.AsRole(), p.ID)
			options = append(options,
				pod.OptionPair{"--cluster.agency-endpoint",
					fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoPort)))},
			)
		}
	}

	args := make([]string, 0, len(options)+len(svrSpec.Args))
	sort.Slice(options, func(i, j int) bool {
		return options[i].CompareTo(options[j]) < 0
	})
	for _, o := range options {
		args = append(args, o.Key+"="+o.Value)
	}
	args = append(args, svrSpec.Args...)

	return args
}

// createArangoSyncArgs creates command line arguments for an arangosync server in the given group.
func createArangoSyncArgs(apiObject metav1.Object, spec api.DeploymentSpec, group api.ServerGroup,
	groupSpec api.ServerGroupSpec, id string) []string {
	options := make([]pod.OptionPair, 0, 64)
	var runCmd string
	var port int

	/*if config.DebugCluster {
		options = append(options,
			pod.OptionPair{"--log.level", "debug"})
	}*/
	if spec.Sync.Monitoring.GetTokenSecretName() != "" {
		options = append(options,
			pod.OptionPair{"--monitoring.token", "$(" + constants.EnvArangoSyncMonitoringToken + ")"},
		)
	}
	masterSecretPath := filepath.Join(k8sutil.MasterJWTSecretVolumeMountDir, constants.SecretKeyToken)
	options = append(options,
		pod.OptionPair{"--master.jwt-secret", masterSecretPath},
	)
	var masterEndpoint []string
	switch group {
	case api.ServerGroupSyncMasters:
		runCmd = "master"
		port = k8sutil.ArangoSyncMasterPort
		masterEndpoint = spec.Sync.ExternalAccess.ResolveMasterEndpoint(k8sutil.CreateSyncMasterClientServiceDNSName(apiObject), port)
		keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
		clientCAPath := filepath.Join(k8sutil.ClientAuthCAVolumeMountDir, constants.SecretCACertificate)
		options = append(options,
			pod.OptionPair{"--server.keyfile", keyPath},
			pod.OptionPair{"--server.client-cafile", clientCAPath},
			pod.OptionPair{"--mq.type", "direct"},
		)
		if spec.IsAuthenticated() {
			clusterSecretPath := filepath.Join(k8sutil.ClusterJWTSecretVolumeMountDir, constants.SecretKeyToken)
			options = append(options,
				pod.OptionPair{"--cluster.jwt-secret", clusterSecretPath},
			)
		}
		dbServiceName := k8sutil.CreateDatabaseClientServiceName(apiObject.GetName())
		scheme := "http"
		if spec.IsSecure() {
			scheme = "https"
		}
		options = append(options,
			pod.OptionPair{"--cluster.endpoint", fmt.Sprintf("%s://%s:%d", scheme, dbServiceName, k8sutil.ArangoPort)})
	case api.ServerGroupSyncWorkers:
		runCmd = "worker"
		port = k8sutil.ArangoSyncWorkerPort
		masterEndpointHost := k8sutil.CreateSyncMasterClientServiceName(apiObject.GetName())
		masterEndpoint = []string{"https://" + net.JoinHostPort(masterEndpointHost, strconv.Itoa(k8sutil.ArangoSyncMasterPort))}
	}
	for _, ep := range masterEndpoint {
		options = append(options,
			pod.OptionPair{"--master.endpoint", ep})
	}
	serverEndpoint := "https://" + net.JoinHostPort(k8sutil.CreatePodDNSName(apiObject, group.AsRole(), id), strconv.Itoa(port))
	options = append(options,
		pod.OptionPair{"--server.endpoint", serverEndpoint},
		pod.OptionPair{"--server.port", strconv.Itoa(port)},
	)

	args := make([]string, 0, 2+len(options)+len(groupSpec.Args))
	sort.Slice(options, func(i, j int) bool {
		return options[i].CompareTo(options[j]) < 0
	})
	args = append(args, "run", runCmd)
	for _, o := range options {
		args = append(args, o.Key+"="+o.Value)
	}
	args = append(args, groupSpec.Args...)

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

func (r *Resources) RenderPodForMember(spec api.DeploymentSpec, status api.DeploymentStatus, memberID string, imageInfo api.ImageInfo) (*core.Pod, error) {
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
		args := createArangodArgs(apiObject, spec, group, status.Members.Agents, m.ID, version, autoUpgrade)

		tlsKeyfileSecretName := ""
		if spec.IsSecure() {
			tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
		}

		rocksdbEncryptionSecretName := ""
		if spec.RocksDB.IsEncrypted() {
			rocksdbEncryptionSecretName = spec.RocksDB.Encryption.GetKeySecretName()
			if err := k8sutil.ValidateEncryptionKeySecret(secrets, rocksdbEncryptionSecretName); err != nil {
				return nil, maskAny(errors.Wrapf(err, "RocksDB encryption key secret validation failed"))
			}
		}
		// Check cluster JWT secret
		var clusterJWTSecretName string
		if spec.IsAuthenticated() {
			if versionHasJWTSecretKeyfile(version) {
				clusterJWTSecretName = spec.Authentication.GetJWTSecretName()
				if err := k8sutil.ValidateTokenSecret(secrets, clusterJWTSecretName); err != nil {
					return nil, maskAny(errors.Wrapf(err, "Cluster JWT secret validation failed"))
				}
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
func (r *Resources) createPodForMember(spec api.DeploymentSpec, memberID string, imageNotFoundOnce *sync.Once) error {
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

	pod, err := r.RenderPodForMember(spec, status, memberID, imageInfo)
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
		m.ArangoVersion = imageInfo.ArangoDBVersion
		m.ImageID = imageInfo.ImageID

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
	args []string, podCreator k8sutil.PodCreator) (*core.Pod, error) {

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

	// Add (anti-)affinity
	p.Spec.Affinity = k8sutil.CreateAffinity(deployment.GetName(), role, !podCreator.IsDeploymentMode(),
		podCreator.GetAffinityRole())

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
func (r *Resources) EnsurePods() error {
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
			if err := r.createPodForMember(spec, m.ID, imageNotFoundOnce); err != nil {
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
