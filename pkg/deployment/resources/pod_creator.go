//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
	"strings"
	"sync"
	"time"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/jwt"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type optionPair struct {
	Key   string
	Value string
}

// CompareTo returns -1 if o < other, 0 if o == other, 1 otherwise
func (o optionPair) CompareTo(other optionPair) int {
	rc := strings.Compare(o.Key, other.Key)
	if rc < 0 {
		return -1
	} else if rc > 0 {
		return 1
	}
	return strings.Compare(o.Value, other.Value)
}

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
	options := make([]optionPair, 0, 64)
	svrSpec := deplSpec.GetServerGroupSpec(group)

	//scheme := NewURLSchemes(bsCfg.SslKeyFile != "").Arangod
	scheme := "tcp"
	if deplSpec.IsSecure() {
		scheme = "ssl"
	}
	options = append(options,
		optionPair{"--server.endpoint", fmt.Sprintf("%s://%s:%d", scheme, deplSpec.GetListenAddr(), k8sutil.ArangoPort)},
	)

	// Authentication
	if deplSpec.IsAuthenticated() {
		// With authentication
		options = append(options,
			optionPair{"--server.authentication", "true"},
		)
		if versionHasJWTSecretKeyfile(version) {
			keyPath := filepath.Join(k8sutil.ClusterJWTSecretVolumeName, constants.SecretKeyToken)
			options = append(options,
				optionPair{"--server.jwt-secret-keyfile", keyPath},
			)
		} else {
			options = append(options,
				optionPair{"--server.jwt-secret", "$(" + constants.EnvArangodJWTSecret + ")"},
			)
		}
	} else {
		// Without authentication
		options = append(options,
			optionPair{"--server.authentication", "false"},
		)
	}

	// Storage engine
	options = append(options,
		optionPair{"--server.storage-engine", deplSpec.GetStorageEngine().AsArangoArgument()},
	)

	// Logging
	options = append(options,
		optionPair{"--log.level", "INFO"},
	)

	// TLS
	if deplSpec.IsSecure() {
		keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
		options = append(options,
			optionPair{"--ssl.keyfile", keyPath},
			optionPair{"--ssl.ecdh-curve", ""}, // This way arangod accepts curves other than P256 as well.
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
			optionPair{"--rocksdb.encryption-keyfile", keyPath},
		)
	}

	options = append(options,
		optionPair{"--database.directory", k8sutil.ArangodVolumeMountDir},
		optionPair{"--log.output", "+"},
	)

	// Auto upgrade?
	if autoUpgrade {
		options = append(options,
			optionPair{"--database.auto-upgrade", "true"},
		)
	}

	versionHasAdvertisedEndpoint := versionHasAdvertisedEndpoint(version)

	/*	if config.ServerThreads != 0 {
		options = append(options,
			optionPair{"--server.threads", strconv.Itoa(config.ServerThreads)})
	}*/
	/*if config.DebugCluster {
		options = append(options,
			optionPair{"--log.level", "startup=trace"})
	}*/
	myTCPURL := scheme + "://" + net.JoinHostPort(k8sutil.CreatePodDNSName(apiObject, group.AsRole(), id), strconv.Itoa(k8sutil.ArangoPort))
	addAgentEndpoints := false
	switch group {
	case api.ServerGroupAgents:
		options = append(options,
			optionPair{"--agency.disaster-recovery-id", id},
			optionPair{"--agency.activate", "true"},
			optionPair{"--agency.my-address", myTCPURL},
			optionPair{"--agency.size", strconv.Itoa(deplSpec.Agents.GetCount())},
			optionPair{"--agency.supervision", "true"},
			optionPair{"--foxx.queues", "false"},
			optionPair{"--server.statistics", "false"},
		)
		for _, p := range agents {
			if p.ID != id {
				dnsName := k8sutil.CreatePodDNSName(apiObject, api.ServerGroupAgents.AsRole(), p.ID)
				options = append(options,
					optionPair{"--agency.endpoint", fmt.Sprintf("%s://%s", scheme, net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoPort)))},
				)
			}
		}
	case api.ServerGroupDBServers:
		addAgentEndpoints = true
		options = append(options,
			optionPair{"--cluster.my-address", myTCPURL},
			optionPair{"--cluster.my-role", "PRIMARY"},
			optionPair{"--foxx.queues", "false"},
			optionPair{"--server.statistics", "true"},
		)
	case api.ServerGroupCoordinators:
		addAgentEndpoints = true
		options = append(options,
			optionPair{"--cluster.my-address", myTCPURL},
			optionPair{"--cluster.my-role", "COORDINATOR"},
			optionPair{"--foxx.queues", "true"},
			optionPair{"--server.statistics", "true"},
		)
		if deplSpec.ExternalAccess.HasAdvertisedEndpoint() && versionHasAdvertisedEndpoint {
			options = append(options,
				optionPair{"--cluster.my-advertised-endpoint", deplSpec.ExternalAccess.GetAdvertisedEndpoint()},
			)
		}
	case api.ServerGroupSingle:
		options = append(options,
			optionPair{"--foxx.queues", "true"},
			optionPair{"--server.statistics", "true"},
		)
		if deplSpec.GetMode() == api.DeploymentModeActiveFailover {
			addAgentEndpoints = true
			options = append(options,
				optionPair{"--replication.automatic-failover", "true"},
				optionPair{"--cluster.my-address", myTCPURL},
				optionPair{"--cluster.my-role", "SINGLE"},
			)
			if deplSpec.ExternalAccess.HasAdvertisedEndpoint() && versionHasAdvertisedEndpoint {
				options = append(options,
					optionPair{"--cluster.my-advertised-endpoint", deplSpec.ExternalAccess.GetAdvertisedEndpoint()},
				)
			}
		}
	}
	if addAgentEndpoints {
		for _, p := range agents {
			dnsName := k8sutil.CreatePodDNSName(apiObject, api.ServerGroupAgents.AsRole(), p.ID)
			options = append(options,
				optionPair{"--cluster.agency-endpoint",
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
func createArangoSyncArgs(apiObject metav1.Object, spec api.DeploymentSpec, group api.ServerGroup, groupSpec api.ServerGroupSpec, agents api.MemberStatusList, id string) []string {
	options := make([]optionPair, 0, 64)
	var runCmd string
	var port int

	/*if config.DebugCluster {
		options = append(options,
			optionPair{"--log.level", "debug"})
	}*/
	if spec.Sync.Monitoring.GetTokenSecretName() != "" {
		options = append(options,
			optionPair{"--monitoring.token", "$(" + constants.EnvArangoSyncMonitoringToken + ")"},
		)
	}
	masterSecretPath := filepath.Join(k8sutil.MasterJWTSecretVolumeMountDir, constants.SecretKeyToken)
	options = append(options,
		optionPair{"--master.jwt-secret", masterSecretPath},
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
			optionPair{"--server.keyfile", keyPath},
			optionPair{"--server.client-cafile", clientCAPath},
			optionPair{"--mq.type", "direct"},
		)
		if spec.IsAuthenticated() {
			clusterSecretPath := filepath.Join(k8sutil.ClusterJWTSecretVolumeMountDir, constants.SecretKeyToken)
			options = append(options,
				optionPair{"--cluster.jwt-secret", clusterSecretPath},
			)
		}
		dbServiceName := k8sutil.CreateDatabaseClientServiceName(apiObject.GetName())
		scheme := "http"
		if spec.IsSecure() {
			scheme = "https"
		}
		options = append(options,
			optionPair{"--cluster.endpoint", fmt.Sprintf("%s://%s:%d", scheme, dbServiceName, k8sutil.ArangoPort)})
	case api.ServerGroupSyncWorkers:
		runCmd = "worker"
		port = k8sutil.ArangoSyncWorkerPort
		masterEndpointHost := k8sutil.CreateSyncMasterClientServiceName(apiObject.GetName())
		masterEndpoint = []string{"https://" + net.JoinHostPort(masterEndpointHost, strconv.Itoa(k8sutil.ArangoSyncMasterPort))}
	}
	for _, ep := range masterEndpoint {
		options = append(options,
			optionPair{"--master.endpoint", ep})
	}
	serverEndpoint := "https://" + net.JoinHostPort(k8sutil.CreatePodDNSName(apiObject, group.AsRole(), id), strconv.Itoa(port))
	options = append(options,
		optionPair{"--server.endpoint", serverEndpoint},
		optionPair{"--server.port", strconv.Itoa(port)},
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

// createLivenessProbe creates configuration for a liveness probe of a server in the given group.
func (r *Resources) createLivenessProbe(spec api.DeploymentSpec, group api.ServerGroup) (*k8sutil.HTTPProbeConfig, error) {
	switch group {
	case api.ServerGroupSingle, api.ServerGroupAgents, api.ServerGroupDBServers:
		authorization := ""
		if spec.IsAuthenticated() {
			secretData, err := r.getJWTSecret(spec)
			if err != nil {
				return nil, maskAny(err)
			}
			authorization, err = jwt.CreateArangodJwtAuthorizationHeader(secretData, "kube-arangodb")
			if err != nil {
				return nil, maskAny(err)
			}
		}
		return &k8sutil.HTTPProbeConfig{
			LocalPath:     "/_api/version",
			Secure:        spec.IsSecure(),
			Authorization: authorization,
		}, nil
	case api.ServerGroupCoordinators:
		return nil, nil
	case api.ServerGroupSyncMasters, api.ServerGroupSyncWorkers:
		authorization := ""
		port := k8sutil.ArangoSyncMasterPort
		if group == api.ServerGroupSyncWorkers {
			port = k8sutil.ArangoSyncWorkerPort
		}
		if spec.Sync.Monitoring.GetTokenSecretName() != "" {
			// Use monitoring token
			token, err := r.getSyncMonitoringToken(spec)
			if err != nil {
				return nil, maskAny(err)
			}
			authorization = "bearer " + token
			if err != nil {
				return nil, maskAny(err)
			}
		} else if group == api.ServerGroupSyncMasters {
			// Fall back to JWT secret
			secretData, err := r.getSyncJWTSecret(spec)
			if err != nil {
				return nil, maskAny(err)
			}
			authorization, err = jwt.CreateArangodJwtAuthorizationHeader(secretData, "kube-arangodb")
			if err != nil {
				return nil, maskAny(err)
			}
		} else {
			// Don't have a probe
			return nil, nil
		}
		return &k8sutil.HTTPProbeConfig{
			LocalPath:     "/_api/version",
			Secure:        spec.IsSecure(),
			Authorization: authorization,
			Port:          port,
		}, nil
	default:
		return nil, nil
	}
}

// createReadinessProbe creates configuration for a readiness probe of a server in the given group.
func (r *Resources) createReadinessProbe(spec api.DeploymentSpec, group api.ServerGroup, version driver.Version) (*k8sutil.HTTPProbeConfig, error) {
	if group != api.ServerGroupSingle && group != api.ServerGroupCoordinators {
		return nil, nil
	}
	authorization := ""
	if spec.IsAuthenticated() {
		secretData, err := r.getJWTSecret(spec)
		if err != nil {
			return nil, maskAny(err)
		}
		authorization, err = jwt.CreateArangodJwtAuthorizationHeader(secretData, "kube-arangodb")
		if err != nil {
			return nil, maskAny(err)
		}
	}
	probeCfg := &k8sutil.HTTPProbeConfig{
		LocalPath:           "/_api/version",
		Secure:              spec.IsSecure(),
		Authorization:       authorization,
		InitialDelaySeconds: 2,
		PeriodSeconds:       2,
	}
	switch spec.GetMode() {
	case api.DeploymentModeActiveFailover:
		probeCfg.LocalPath = "/_admin/echo"
	}

	// /_admin/server/availability is the way to go, it is available since 3.3.9
	if version.CompareTo("3.3.9") >= 0 {
		probeCfg.LocalPath = "/_admin/server/availability"
	}

	return probeCfg, nil
}

// createPodFinalizers creates a list of finalizers for a pod created for the given group.
func (r *Resources) createPodFinalizers(group api.ServerGroup) []string {
	switch group {
	case api.ServerGroupAgents:
		return []string{constants.FinalizerPodAgencyServing}
	case api.ServerGroupDBServers:
		return []string{constants.FinalizerPodDrainDBServer}
	default:
		return nil
	}
}

// createPodTolerations creates a list of tolerations for a pod created for the given group.
func (r *Resources) createPodTolerations(group api.ServerGroup, groupSpec api.ServerGroupSpec) []v1.Toleration {
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

// createPodForMember creates all Pods listed in member status
func (r *Resources) createPodForMember(spec api.DeploymentSpec, memberID string, imageNotFoundOnce *sync.Once) error {
	kubecli := r.context.GetKubeCli()
	log := r.log
	apiObject := r.context.GetAPIObject()
	ns := r.context.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	status, lastVersion := r.context.GetStatus()
	m, group, found := status.Members.ElementByID(memberID)
	if !found {
		return maskAny(fmt.Errorf("Member '%s' not found", memberID))
	}
	groupSpec := spec.GetServerGroupSpec(group)
	lifecycleImage := r.context.GetLifecycleImage()
	alpineImage := r.context.GetAlpineImage()
	terminationGracePeriod := group.DefaultTerminationGracePeriod()
	tolerations := r.createPodTolerations(group, groupSpec)
	serviceAccountName := groupSpec.GetServiceAccountName()

	// Update pod name
	role := group.AsRole()
	roleAbbr := group.AsRoleAbbreviated()
	podSuffix := createPodSuffix(spec)
	m.PodName = k8sutil.CreatePodName(apiObject.GetName(), roleAbbr, m.ID, podSuffix)
	newPhase := api.MemberPhaseCreated
	// Select image
	var imageInfo api.ImageInfo
	if current := status.CurrentImage; current != nil {
		// Use current image
		imageInfo = *current
	} else {
		// Find image ID
		info, imageFound := status.Images.GetByImage(spec.GetImage())
		if !imageFound {
			imageNotFoundOnce.Do(func() {
				log.Debug().Str("image", spec.GetImage()).Msg("Image ID is not known yet for image")
			})
			return nil
		}
		imageInfo = info
		// Save image as current image
		status.CurrentImage = &info
	}
	// Create pod
	if group.IsArangod() {
		// Prepare arguments
		version := imageInfo.ArangoDBVersion
		autoUpgrade := m.Conditions.IsTrue(api.ConditionTypeAutoUpgrade)
		if autoUpgrade {
			newPhase = api.MemberPhaseUpgrading
		}
		args := createArangodArgs(apiObject, spec, group, status.Members.Agents, m.ID, version, autoUpgrade)
		env := make(map[string]k8sutil.EnvValue)
		livenessProbe, err := r.createLivenessProbe(spec, group)
		if err != nil {
			return maskAny(err)
		}
		readinessProbe, err := r.createReadinessProbe(spec, group, version)
		if err != nil {
			return maskAny(err)
		}
		tlsKeyfileSecretName := ""
		if spec.IsSecure() {
			tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
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
		rocksdbEncryptionSecretName := ""
		if spec.RocksDB.IsEncrypted() {
			rocksdbEncryptionSecretName = spec.RocksDB.Encryption.GetKeySecretName()
			if err := k8sutil.ValidateEncryptionKeySecret(secrets, rocksdbEncryptionSecretName); err != nil {
				return maskAny(errors.Wrapf(err, "RocksDB encryption key secret validation failed"))
			}
		}
		// Check cluster JWT secret
		var clusterJWTSecretName string
		if spec.IsAuthenticated() {
			if versionHasJWTSecretKeyfile(version) {
				clusterJWTSecretName = spec.Authentication.GetJWTSecretName()
				if err := k8sutil.ValidateTokenSecret(secrets, clusterJWTSecretName); err != nil {
					return maskAny(errors.Wrapf(err, "Cluster JWT secret validation failed"))
				}
			} else {
				env[constants.EnvArangodJWTSecret] = k8sutil.EnvValue{
					SecretName: spec.Authentication.GetJWTSecretName(),
					SecretKey:  constants.SecretKeyToken,
				}
			}

		}

		if spec.License.HasSecretName() {
			env[constants.EnvArangoLicenseKey] = k8sutil.EnvValue{
				SecretName: spec.License.GetSecretName(),
				SecretKey:  constants.SecretKeyToken,
			}
		}

		engine := spec.GetStorageEngine().AsArangoArgument()
		requireUUID := group == api.ServerGroupDBServers && m.IsInitialized
		finalizers := r.createPodFinalizers(group)
		if err := k8sutil.CreateArangodPod(kubecli, spec.IsDevelopment(), apiObject, role, m.ID, m.PodName, m.PersistentVolumeClaimName, imageInfo.ImageID, lifecycleImage, alpineImage, spec.GetImagePullPolicy(),
			engine, requireUUID, terminationGracePeriod, args, env, finalizers, livenessProbe, readinessProbe, tolerations, serviceAccountName, tlsKeyfileSecretName, rocksdbEncryptionSecretName,
			clusterJWTSecretName, groupSpec.GetNodeSelector()); err != nil {
			return maskAny(err)
		}
		log.Debug().Str("pod-name", m.PodName).Msg("Created pod")
	} else if group.IsArangosync() {
		// Check image
		if !imageInfo.Enterprise {
			log.Debug().Str("image", spec.GetImage()).Msg("Image is not an enterprise image")
			return maskAny(fmt.Errorf("Image '%s' does not contain an Enterprise version of ArangoDB", spec.GetImage()))
		}
		var tlsKeyfileSecretName, clientAuthCASecretName, masterJWTSecretName, clusterJWTSecretName string
		// Check master JWT secret
		masterJWTSecretName = spec.Sync.Authentication.GetJWTSecretName()
		if err := k8sutil.ValidateTokenSecret(secrets, masterJWTSecretName); err != nil {
			return maskAny(errors.Wrapf(err, "Master JWT secret validation failed"))
		}
		// Check monitoring token secret
		monitoringTokenSecretName := spec.Sync.Monitoring.GetTokenSecretName()
		if err := k8sutil.ValidateTokenSecret(secrets, monitoringTokenSecretName); err != nil {
			return maskAny(errors.Wrapf(err, "Monitoring token secret validation failed"))
		}
		if group == api.ServerGroupSyncMasters {
			// Create TLS secret
			tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
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
			// Check cluster JWT secret
			if spec.IsAuthenticated() {
				clusterJWTSecretName = spec.Authentication.GetJWTSecretName()
				if err := k8sutil.ValidateTokenSecret(secrets, clusterJWTSecretName); err != nil {
					return maskAny(errors.Wrapf(err, "Cluster JWT secret validation failed"))
				}
			}
			// Check client-auth CA certificate secret
			clientAuthCASecretName = spec.Sync.Authentication.GetClientCASecretName()
			if err := k8sutil.ValidateCACertificateSecret(secrets, clientAuthCASecretName); err != nil {
				return maskAny(errors.Wrapf(err, "Client authentication CA certificate secret validation failed"))
			}
		}

		// Prepare arguments
		args := createArangoSyncArgs(apiObject, spec, group, groupSpec, status.Members.Agents, m.ID)
		env := make(map[string]k8sutil.EnvValue)
		if spec.Sync.Monitoring.GetTokenSecretName() != "" {
			env[constants.EnvArangoSyncMonitoringToken] = k8sutil.EnvValue{
				SecretName: spec.Sync.Monitoring.GetTokenSecretName(),
				SecretKey:  constants.SecretKeyToken,
			}
		}
		if spec.License.HasSecretName() {
			env[constants.EnvArangoLicenseKey] = k8sutil.EnvValue{
				SecretName: spec.License.GetSecretName(),
				SecretKey:  constants.SecretKeyToken,
			}
		}
		livenessProbe, err := r.createLivenessProbe(spec, group)
		if err != nil {
			return maskAny(err)
		}
		affinityWithRole := ""
		if group == api.ServerGroupSyncWorkers {
			affinityWithRole = api.ServerGroupDBServers.AsRole()
		}
		if err := k8sutil.CreateArangoSyncPod(kubecli, spec.IsDevelopment(), apiObject, role, m.ID, m.PodName, imageInfo.ImageID, lifecycleImage, spec.GetImagePullPolicy(), terminationGracePeriod, args, env,
			livenessProbe, tolerations, serviceAccountName, tlsKeyfileSecretName, clientAuthCASecretName, masterJWTSecretName, clusterJWTSecretName, affinityWithRole, groupSpec.GetNodeSelector()); err != nil {
			return maskAny(err)
		}
		log.Debug().Str("pod-name", m.PodName).Msg("Created pod")
	}
	// Record new member phase
	m.Phase = newPhase
	m.Conditions.Remove(api.ConditionTypeReady)
	m.Conditions.Remove(api.ConditionTypeTerminated)
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

// EnsurePods creates all Pods listed in member status
func (r *Resources) EnsurePods() error {
	iterator := r.context.GetServerGroupIterator()
	status, _ := r.context.GetStatus()
	imageNotFoundOnce := &sync.Once{}
	if err := iterator.ForeachServerGroup(func(group api.ServerGroup, groupSpec api.ServerGroupSpec, status *api.MemberStatusList) error {
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
	}, &status); err != nil {
		return maskAny(err)
	}
	return nil
}

func createPodSuffix(spec api.DeploymentSpec) string {
	raw, _ := json.Marshal(spec)
	hash := sha1.Sum(raw)
	return fmt.Sprintf("%0x", hash)[:6]
}
