//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package shared

const (
	// Arango constants
	ArangoPort           = 8529
	ArangoSyncMasterPort = 8629
	ArangoSyncWorkerPort = 8729
	ArangoExporterPort   = 9101

	ArangoExporterStatusEndpoint        = "/_api/version"
	ArangoExporterClusterHealthEndpoint = "/_admin/cluster/health"
	ArangoExporterInternalEndpoint      = "/_admin/metrics"
	ArangoExporterInternalEndpointV2    = "/_admin/metrics/v2"
	ArangoExporterDefaultEndpoint       = "/metrics"

	ArangoSyncStatusEndpoint = "/_api/version"

	// K8s constants
	ClusterIPNone       = "None"
	TopologyKeyHostname = "kubernetes.io/hostname"

	NodeArchAffinityLabel     = "kubernetes.io/arch"
	NodeArchAffinityLabelBeta = "beta.kubernetes.io/arch"

	// Pod constants
	ServerContainerName             = "server"
	ExporterContainerName           = "exporter"
	ArangodVolumeName               = "arangod-data"
	TlsKeyfileVolumeName            = "tls-keyfile"
	ClientAuthCAVolumeName          = "client-auth-ca"
	ClusterJWTSecretVolumeName      = "cluster-jwt"
	MasterJWTSecretVolumeName       = "master-jwt"
	LifecycleVolumeName             = "lifecycle"
	FoxxAppEphemeralVolumeName      = "ephemeral-apps"
	TMPEphemeralVolumeName          = "ephemeral-tmp"
	ArangoDTimezoneVolumeName       = "arangod-timezone"
	RocksdbEncryptionVolumeName     = "rocksdb-encryption"
	ExporterJWTVolumeName           = "exporter-jwt"
	ArangodVolumeMountDir           = "/data"
	RocksDBEncryptionVolumeMountDir = "/secrets/rocksdb/encryption"
	TLSKeyfileVolumeMountDir        = "/secrets/tls"
	TLSSNIKeyfileVolumeMountDir     = "/secrets/sni"
	ClientAuthCAVolumeMountDir      = "/secrets/client-auth/ca"
	ClusterJWTSecretVolumeMountDir  = "/secrets/cluster/jwt"
	ExporterJWTVolumeMountDir       = "/secrets/exporter/jwt"
	MasterJWTSecretVolumeMountDir   = "/secrets/master/jwt"

	ServerPortName   = "server"
	ExporterPortName = "exporter"
)
