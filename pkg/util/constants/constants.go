//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package constants

import "github.com/arangodb/kube-arangodb/pkg/util"

const (
	EnvOperatorNodeName       = "MY_NODE_NAME"
	EnvOperatorNodeNameArango = "NODE_NAME"
	EnvOperatorPodName        = "MY_POD_NAME"
	EnvOperatorPodNamespace   = "MY_POD_NAMESPACE"
	EnvOperatorCoreContainer  = "MY_POD_CORE_CONTAINER"
	EnvOperatorPodIP          = "MY_POD_IP"

	EnvOperatorNodeNameEnv       util.EnvironmentVariable = EnvOperatorNodeName
	EnvOperatorNodeNameArangoEnv util.EnvironmentVariable = EnvOperatorNodeNameArango
	EnvOperatorPodNameEnv        util.EnvironmentVariable = EnvOperatorPodName
	EnvOperatorPodNamespaceEnv   util.EnvironmentVariable = EnvOperatorPodNamespace
	EnvOperatorCoreContainerEnv  util.EnvironmentVariable = EnvOperatorCoreContainer
	EnvOperatorPodIPEnv          util.EnvironmentVariable = EnvOperatorPodIP

	EnvArangoJobSAName = "ARANGOJOB_SA_NAME"

	PathMountServiceAccount          = "/var/run/secrets/kubernetes.io/serviceaccount"
	PathMountServiceAccountNamespace = PathMountServiceAccount + "/namespace"

	EnvArangoLicenseKey          = "ARANGO_LICENSE_KEY"          // Contains the License Key for the Docker Image
	EnvArangoSyncMonitoringToken = "ARANGOSYNC_MONITORING_TOKEN" // Constains monitoring token for ArangoSync servers

	SecretEncryptionKey = "key"        // Key in a Secret.Data used to store an 32-byte encryption key
	SecretKeyToken      = "token"      // Key inside a Secret used to hold a JWT or monitoring token
	SecretKeyV2Token    = "token-v2"   // Key inside a Secret used to hold a License in V2 Format
	SecretKeyV2License  = "license-v2" // Key inside a Secret used to hold a License in V2 Format

	SecretCACertificate = "ca.crt" // Key in Secret.data used to store a PEM encoded CA certificate (public key)
	SecretCAKey         = "ca.key" // Key in Secret.data used to store a PEM encoded CA private key

	SecretTLSKeyfile = "tls.keyfile" // Key in Secret.data used to store a PEM encoded TLS certificate in the format used by ArangoDB (`--ssl.keyfile`)

	SecretUsername = "username" // Key in Secret.data used to store a username used for basic authentication
	SecretPassword = "password" // Key in Secret.data used to store a password used for basic authentication

	SecretCredentialsAccessKey = "accessKey" // Key in Secret used to store an S3 API access key
	SecretCredentialsSecretKey = "secretKey" // Key in Secret used to store an S3 API secret key

	SecretCredentialsServiceAccount = "serviceAccount" // Key in Secret used to store an GCS ServiceAccount File

	SecretAccessPackageYaml = "accessPackage.yaml" // Key in Secret.data used to store a YAML encoded access package

	FinalizerDeplRemoveChildFinalizers = "database.arangodb.com/remove-child-finalizers" // Finalizer added to ArangoDeployment, indicating the need to remove finalizers from all children
	FinalizerDeplReplStopSync          = "replication.database.arangodb.com/stop-sync"   // Finalizer added to ArangoDeploymentReplication, indicating the need to stop synchronization
	FinalizerPodAgencyServing          = "agent.database.arangodb.com/agency-serving"    // Finalizer added to Agents, indicating the need for keeping enough agents alive
	FinalizerPodDrainDBServer          = "dbserver.database.arangodb.com/drain"          // Finalizer added to DBServers, indicating the need for draining that dbserver
	FinalizerPodGracefulShutdown       = "database.arangodb.com/graceful-shutdown"       // Finalizer added to All members, indicating the need for graceful shutdown
	FinalizerPVCMemberExists           = "pvc.database.arangodb.com/member-exists"       // Finalizer added to PVCs, indicating the need to keep is as long as its member exists
	FinalizerDelayPodTermination       = "pod.database.arangodb.com/delay"               // Finalizer added to Pod, delays termination

	AnnotationShutdownManagedContainer      = "shutdown.arangodb.com/managed"
	AnnotationShutdownContainer             = "container.shutdown.arangodb.com"
	AnnotationShutdownCoreContainer         = "core.shutdown.arangodb.com"
	AnnotationShutdownCoreContainerModeWait = "wait"

	AnnotationEnforceAntiAffinity = "database.arangodb.com/enforce-anti-affinity" // Key of annotation added to PVC. Value is a boolean "true" or "false"

	BackupLabelRole      = "backup/role"
	MLLabelRole          = "ml/role"
	AnalyticsLabelRole   = "analytics/role"
	NetworkingLabelRole  = "networking/role"
	PlatformLabelRole    = "platform/role"
	SchedulerLabelRole   = "scheduler/role"
	AppsLabelRole        = "apps/role"
	ClusterSyncLabelRole = "clustersync/role"
	LabelRole            = "role"
	LabelRoleLeader      = "leader"
)

const (
	MyContainerNameEnv util.EnvironmentVariable = "MY_CONTAINER_NAME"
)

func ManagedFinalizers() []string {
	return []string{
		FinalizerDeplRemoveChildFinalizers,
		FinalizerDeplReplStopSync,
		FinalizerPodAgencyServing,
		FinalizerPodDrainDBServer,
		FinalizerPodGracefulShutdown,
		FinalizerPVCMemberExists,
		FinalizerDelayPodTermination,
	}
}
