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
	"fmt"
	"net"
	"strconv"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/arangod"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

type optionPair struct {
	Key   string
	Value string
}

// createArangodArgs creates command line arguments for an arangod server in the given group.
func (d *Deployment) createArangodArgs(apiObject *api.ArangoDeployment, group api.ServerGroup, spec api.ServerGroupSpec, agents api.MemberStatusList, id string) []string {
	options := make([]optionPair, 0, 64)

	// Endpoint
	listenAddr := "[::]"
	/*	if apiObject.Spec.Di.DisableIPv6 {
		listenAddr = "0.0.0.0"
	}*/
	//scheme := NewURLSchemes(bsCfg.SslKeyFile != "").Arangod
	scheme := "tcp"
	options = append(options,
		optionPair{"--server.endpoint", fmt.Sprintf("%s://%s:%d", scheme, listenAddr, k8sutil.ArangoPort)},
	)

	// Authentication
	if apiObject.Spec.Authentication.JWTSecretName != "" {
		// With authentication
		options = append(options,
			optionPair{"--server.authentication", "true"},
			// TODO jwt-secret file
		)
	} else {
		// Without authentication
		options = append(options,
			optionPair{"--server.authentication", "false"},
		)
	}

	// Storage engine
	options = append(options,
		optionPair{"--server.storage-engine", string(apiObject.Spec.StorageEngine)},
	)

	// Logging
	options = append(options,
		optionPair{"--log.level", "INFO"},
	)

	// SSL
	/*if bsCfg.SslKeyFile != "" {
		sslSection := &configSection{
			Name: "ssl",
			Settings: map[string]string{
				"keyfile": bsCfg.SslKeyFile,
			},
		}
		if bsCfg.SslCAFile != "" {
			sslSection.Settings["cafile"] = bsCfg.SslCAFile
		}
		config = append(config, sslSection)
	}*/

	// RocksDB
	if apiObject.Spec.RocksDB.Encryption.KeySecretName != "" {
		/*args = append(args,
			fmt.Sprintf("--rocksdb.encryption-keyfile=%s", apiObject.Spec.StorageEngine),
		)
		rocksdbSection := &configSection{
			Name: "rocksdb",
			Settings: map[string]string{
				"encryption-keyfile": bsCfg.RocksDBEncryptionKeyFile,
			},
		}
		config = append(config, rocksdbSection)*/
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
			optionPair{"--cluster.my-id", id},
			optionPair{"--agency.activate", "true"},
			optionPair{"--agency.my-address", myTCPURL},
			optionPair{"--agency.size", strconv.Itoa(apiObject.Spec.Agents.Count)},
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
		/*if agentRecoveryID != "" {
			options = append(options,
				optionPair{"--agency.disaster-recovery-id", agentRecoveryID},
			)
		}*/
	case api.ServerGroupDBServers:
		addAgentEndpoints = true
		options = append(options,
			optionPair{"--cluster.my-id", id},
			optionPair{"--cluster.my-address", myTCPURL},
			optionPair{"--cluster.my-role", "PRIMARY"},
			optionPair{"--foxx.queues", "false"},
			optionPair{"--server.statistics", "true"},
		)
	case api.ServerGroupCoordinators:
		addAgentEndpoints = true
		options = append(options,
			optionPair{"--cluster.my-id", id},
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
		if apiObject.Spec.Mode == api.DeploymentModeResilientSingle {
			addAgentEndpoints = true
			options = append(options,
				optionPair{"--replication.automatic-failover", "true"},
				optionPair{"--cluster.my-id", id},
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

	args := make([]string, 0, len(options)+len(spec.Args))
	for _, o := range options {
		args = append(args, o.Key+"="+o.Value)
	}
	args = append(args, spec.Args...)

	return args
}

// createArangoSyncArgs creates command line arguments for an arangosync server in the given group.
func (d *Deployment) createArangoSyncArgs(apiObject *api.ArangoDeployment, group api.ServerGroup, spec api.ServerGroupSpec, agents api.MemberStatusList, id string) []string {
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

	if err := apiObject.ForeachServerGroup(func(group api.ServerGroup, spec api.ServerGroupSpec, status *api.MemberStatusList) error {
		for _, m := range *status {
			if m.State != api.MemberStateNone {
				continue
			}
			// Create pod
			role := group.AsRole()
			if group.IsArangod() {
				args := d.createArangodArgs(apiObject, group, spec, d.status.Members.Agents, m.ID)
				env := make(map[string]string)
				livenessProbe, err := d.createLivenessProbe(apiObject, group)
				if err != nil {
					return maskAny(err)
				}
				readinessProbe, err := d.createReadinessProbe(apiObject, group)
				if err != nil {
					return maskAny(err)
				}
				if err := k8sutil.CreateArangodPod(kubecli, apiObject.Spec.IsDevelopment(), apiObject, role, m.ID, m.PersistentVolumeClaimName, apiObject.Spec.Image, apiObject.Spec.ImagePullPolicy, args, env, livenessProbe, readinessProbe); err != nil {
					return maskAny(err)
				}
			} else if group.IsArangosync() {
				args := d.createArangoSyncArgs(apiObject, group, spec, d.status.Members.Agents, m.ID)
				env := make(map[string]string)
				livenessProbe, err := d.createLivenessProbe(apiObject, group)
				if err != nil {
					return maskAny(err)
				}
				affinityWithRole := ""
				if group == api.ServerGroupSyncWorkers {
					affinityWithRole = api.ServerGroupDBServers.AsRole()
				}
				if err := k8sutil.CreateArangoSyncPod(kubecli, apiObject.Spec.IsDevelopment(), apiObject, role, m.ID, apiObject.Spec.Sync.Image, apiObject.Spec.Sync.ImagePullPolicy, args, env, livenessProbe, affinityWithRole); err != nil {
					return maskAny(err)
				}
			}
			// Record new member state
			m.State = api.MemberStateCreating
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
