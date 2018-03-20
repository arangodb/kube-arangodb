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

package deployment

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
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

// createArangodArgs creates command line arguments for an arangod server in the given group.
func createArangodArgs(apiObject metav1.Object, deplSpec api.DeploymentSpec, group api.ServerGroup, agents api.MemberStatusList, id string) []string {
	options := make([]optionPair, 0, 64)
	svrSpec := deplSpec.GetServerGroupSpec(group)

	// Endpoint
	listenAddr := "[::]"
	/*	if apiObject.Spec.Di.DisableIPv6 {
		listenAddr = "0.0.0.0"
	}*/
	//scheme := NewURLSchemes(bsCfg.SslKeyFile != "").Arangod
	scheme := "tcp"
	if deplSpec.IsSecure() {
		scheme = "ssl"
	}
	options = append(options,
		optionPair{"--server.endpoint", fmt.Sprintf("%s://%s:%d", scheme, listenAddr, k8sutil.ArangoPort)},
	)

	// Authentication
	if deplSpec.IsAuthenticated() {
		// With authentication
		options = append(options,
			optionPair{"--server.authentication", "true"},
			optionPair{"--server.jwt-secret", "$(" + constants.EnvArangodJWTSecret + ")"},
		)
	} else {
		// Without authentication
		options = append(options,
			optionPair{"--server.authentication", "false"},
		)
	}

	// Storage engine
	options = append(options,
		optionPair{"--server.storage-engine", string(deplSpec.StorageEngine)},
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
			optionPair{"--agency.size", strconv.Itoa(deplSpec.Agents.Count)},
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
	case api.ServerGroupSingle:
		options = append(options,
			optionPair{"--foxx.queues", "true"},
			optionPair{"--server.statistics", "true"},
		)
		if deplSpec.Mode == api.DeploymentModeResilientSingle {
			addAgentEndpoints = true
			options = append(options,
				optionPair{"--replication.automatic-failover", "true"},
				optionPair{"--cluster.my-address", myTCPURL},
				optionPair{"--cluster.my-role", "SINGLE"},
			)
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
func createArangoSyncArgs(apiObject *api.ArangoDeployment, group api.ServerGroup, spec api.ServerGroupSpec, agents api.MemberStatusList, id string) []string {
	// TODO
	return nil
}

// createLivenessProbe creates configuration for a liveness probe of a server in the given group.
func (d *Deployment) createLivenessProbe(apiObject *api.ArangoDeployment, group api.ServerGroup) (*k8sutil.HTTPProbeConfig, error) {
	switch group {
	case api.ServerGroupSingle, api.ServerGroupAgents, api.ServerGroupDBServers:
		authorization := ""
		if apiObject.Spec.IsAuthenticated() {
			secretData, err := d.getJWTSecret(apiObject)
			if err != nil {
				return nil, maskAny(err)
			}
			authorization, err = arangod.CreateArangodJwtAuthorizationHeader(secretData)
			if err != nil {
				return nil, maskAny(err)
			}
		}
		return &k8sutil.HTTPProbeConfig{
			LocalPath:     "/_api/version",
			Secure:        apiObject.Spec.IsSecure(),
			Authorization: authorization,
		}, nil
	case api.ServerGroupCoordinators:
		return nil, nil
	case api.ServerGroupSyncMasters, api.ServerGroupSyncWorkers:
		authorization := ""
		if apiObject.Spec.Sync.Monitoring.TokenSecretName != "" {
			// Use monitoring token
			token, err := d.getSyncMonitoringToken(apiObject)
			if err != nil {
				return nil, maskAny(err)
			}
			authorization = "bearer: " + token
			if err != nil {
				return nil, maskAny(err)
			}
		} else if group == api.ServerGroupSyncMasters {
			// Fall back to JWT secret
			secretData, err := d.getSyncJWTSecret(apiObject)
			if err != nil {
				return nil, maskAny(err)
			}
			authorization, err = arangod.CreateArangodJwtAuthorizationHeader(secretData)
			if err != nil {
				return nil, maskAny(err)
			}
		} else {
			// Don't have a probe
			return nil, nil
		}
		return &k8sutil.HTTPProbeConfig{
			LocalPath:     "/_api/version",
			Secure:        apiObject.Spec.IsSecure(),
			Authorization: authorization,
		}, nil
	default:
		return nil, nil
	}
}

// createReadinessProbe creates configuration for a readiness probe of a server in the given group.
func (d *Deployment) createReadinessProbe(apiObject *api.ArangoDeployment, group api.ServerGroup) (*k8sutil.HTTPProbeConfig, error) {
	if group != api.ServerGroupCoordinators {
		return nil, nil
	}
	authorization := ""
	if apiObject.Spec.IsAuthenticated() {
		secretData, err := d.getJWTSecret(apiObject)
		if err != nil {
			return nil, maskAny(err)
		}
		authorization, err = arangod.CreateArangodJwtAuthorizationHeader(secretData)
		if err != nil {
			return nil, maskAny(err)
		}
	}
	return &k8sutil.HTTPProbeConfig{
		LocalPath:     "/_api/version",
		Secure:        apiObject.Spec.IsSecure(),
		Authorization: authorization,
	}, nil
}

// ensurePods creates all Pods listed in member status
func (d *Deployment) ensurePods(apiObject *api.ArangoDeployment) error {
	kubecli := d.deps.KubeCli
	log := d.deps.Log
	ns := apiObject.GetNamespace()

	if err := apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.State != api.MemberStateNone {
				continue
			}
			// Update pod name
			role := group.AsRole()
			podSuffix := createPodSuffix(apiObject.Spec)
			m.PodName = k8sutil.CreatePodName(apiObject.GetName(), role, m.ID, podSuffix)
			// Create pod
			if group.IsArangod() {
				// Find image ID
				info, found := apiObject.Status.Images.GetByImage(apiObject.Spec.Image)
				if !found {
					log.Debug().Str("image", apiObject.Spec.Image).Msg("Image ID is not known yet for image")
					return nil
				}
				// Prepare arguments
				args := createArangodArgs(apiObject, apiObject.Spec, group, d.status.Members.Agents, m.ID)
				env := make(map[string]k8sutil.EnvValue)
				livenessProbe, err := d.createLivenessProbe(apiObject, group)
				if err != nil {
					return maskAny(err)
				}
				readinessProbe, err := d.createReadinessProbe(apiObject, group)
				if err != nil {
					return maskAny(err)
				}
				tlsKeyfileSecretName := ""
				if apiObject.Spec.IsSecure() {
					tlsKeyfileSecretName = k8sutil.CreateTLSKeyfileSecretName(apiObject.GetName(), role, m.ID)
					serverNames := []string{
						k8sutil.CreateDatabaseClientServiceDNSName(apiObject),
						k8sutil.CreatePodDNSName(apiObject, role, m.ID),
					}
					owner := apiObject.AsOwner()
					if err := createServerCertificate(log, kubecli.CoreV1(), serverNames, apiObject.Spec.TLS, tlsKeyfileSecretName, ns, &owner); err != nil && !k8sutil.IsAlreadyExists(err) {
						return maskAny(errors.Wrapf(err, "Failed to create TLS keyfile secret"))
					}
				}
				rocksdbEncryptionSecretName := ""
				if apiObject.Spec.RocksDB.IsEncrypted() {
					rocksdbEncryptionSecretName = apiObject.Spec.RocksDB.Encryption.KeySecretName
					if err := k8sutil.ValidateEncryptionKeySecret(kubecli.CoreV1(), rocksdbEncryptionSecretName, ns); err != nil {
						return maskAny(errors.Wrapf(err, "RocksDB encryption key secret validation failed"))
					}
				}
				if apiObject.Spec.IsAuthenticated() {
					env[constants.EnvArangodJWTSecret] = k8sutil.EnvValue{
						SecretName: apiObject.Spec.Authentication.JWTSecretName,
						SecretKey:  constants.SecretKeyJWT,
					}
				}
				if err := k8sutil.CreateArangodPod(kubecli, apiObject.Spec.IsDevelopment(), apiObject, role, m.ID, m.PodName, m.PersistentVolumeClaimName, info.ImageID, apiObject.Spec.ImagePullPolicy, args, env, livenessProbe, readinessProbe, tlsKeyfileSecretName, rocksdbEncryptionSecretName); err != nil {
					return maskAny(err)
				}
			} else if group.IsArangosync() {
				// Find image ID
				info, found := apiObject.Status.Images.GetByImage(apiObject.Spec.Sync.Image)
				if !found {
					log.Debug().Str("image", apiObject.Spec.Sync.Image).Msg("Image ID is not known yet for image")
					return nil
				}
				// Prepare arguments
				args := createArangoSyncArgs(apiObject, group, spec, d.status.Members.Agents, m.ID)
				env := make(map[string]k8sutil.EnvValue)
				livenessProbe, err := d.createLivenessProbe(apiObject, group)
				if err != nil {
					return maskAny(err)
				}
				affinityWithRole := ""
				if group == api.ServerGroupSyncWorkers {
					affinityWithRole = api.ServerGroupDBServers.AsRole()
				}
				if err := k8sutil.CreateArangoSyncPod(kubecli, apiObject.Spec.IsDevelopment(), apiObject, role, m.ID, m.PodName, info.ImageID, apiObject.Spec.Sync.ImagePullPolicy, args, env, livenessProbe, affinityWithRole); err != nil {
					return maskAny(err)
				}
			}
			// Record new member state
			m.State = api.MemberStateCreated
			m.Conditions.Remove(api.ConditionTypeReady)
			m.Conditions.Remove(api.ConditionTypeTerminated)
			if err := status.Update(m); err != nil {
				return maskAny(err)
			}
			if err := d.updateCRStatus(); err != nil {
				return maskAny(err)
			}
			// Create event
			d.createEvent(k8sutil.NewMemberAddEvent(m.PodName, role, apiObject))
		}
		return nil
	}, &d.status); err != nil {
		return maskAny(err)
	}
	return nil
}

func createPodSuffix(spec api.DeploymentSpec) string {
	raw, _ := json.Marshal(spec)
	hash := sha1.Sum(raw)
	return fmt.Sprintf("%0x", hash)[:6]
}
