//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package v1

import meta "k8s.io/apimachinery/pkg/apis/meta/v1"

// ArangoBackupSpec Spec of the ArangoBackup Custom Resource
type ArangoBackupSpec struct {
	// Deployment describes the deployment which should have a backup
	Deployment ArangoBackupSpecDeployment `json:"deployment,omitempty"`

	// Options specifies backup options
	Options *ArangoBackupSpecOptions `json:"options,omitempty"`

	// Download Backup download settings
	Download *ArangoBackupSpecDownload `json:"download,omitempty"`

	// Upload Backup upload settings.
	// This field can be removed and created again with different values. This operation will trigger upload again.
	Upload *ArangoBackupSpecOperation `json:"upload,omitempty"`

	// PolicyName name of the ArangoBackupPolicy which created this Custom Resource
	// +doc/immutable: can't be changed after backup creation
	PolicyName *string `json:"policyName,omitempty"`

	Backoff *ArangoBackupSpecBackOff `json:"backoff,omitempty"`

	// Lifetime is the time after which the backup will be deleted. Format: "1.5h" or "2h45m".
	Lifetime *meta.Duration `json:"lifetime,omitempty"`
}

// ArangoBackupSpecDeployment describes the deployment which should have a backup
type ArangoBackupSpecDeployment struct {
	// Name of the ArangoDeployment Custom Resource within same namespace as ArangoBackup Custom Resource.
	// +doc/immutable: can't be changed after backup creation
	Name string `json:"name,omitempty"`
}

type ArangoBackupSpecOptions struct {
	// Timeout for Backup creation request in seconds. Works only when AsyncBackupCreation feature is set to false.
	// +doc/immutable: can't be changed after backup creation
	// +doc/default: 30
	Timeout *float32 `json:"timeout,omitempty"`
	// AllowInconsistent flag for Backup creation request.
	// If this value is set to true, backup is taken even if we are not able to acquire lock.
	// +doc/immutable: can't be changed after backup creation
	// +doc/default: false
	AllowInconsistent *bool `json:"allowInconsistent,omitempty"`
}

type ArangoBackupSpecOperation struct {
	// RepositoryURL is the URL path for file storage
	// Same repositoryURL needs to be defined in `credentialsSecretName` if protocol is other than local.
	// Format: `<protocol>:/<path>`
	// +doc/example: s3://my-bucket/test
	// +doc/example: azure://test
	// +doc/immutable: can't be changed after backup creation
	// +doc/link: rclone.org|https://rclone.org/docs/#syntax-of-remote-paths
	RepositoryURL string `json:"repositoryURL"`
	// CredentialsSecretName is the name of the secret used while accessing repository
	// +doc/immutable: can't be changed after backup creation
	// +doc/link: Defining a secret for backup upload or download|/docs/backup-resource.md#defining-a-secret-for-backup-upload-or-download
	CredentialsSecretName string `json:"credentialsSecretName,omitempty"`
}

type ArangoBackupSpecDownload struct {
	ArangoBackupSpecOperation `json:",inline"`

	// ID of the ArangoBackup to be downloaded
	// +doc/immutable: can't be changed after backup creation
	ID string `json:"id"`
}
