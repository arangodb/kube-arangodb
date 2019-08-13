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
// Author Adam Janikowski
//

package v1alpha

type ArangoBackupSpec struct {
	// Deployment
	Deployment ArangoBackupSpecDeployment `json:"deployment,omitempty"`

	Options *ArangoBackupSpecOptions `json:"options,omitempty"`

	// Download
	Download *ArangoBackupSpecOperation `json:"download,omitempty"`

	// Upload
	Upload *ArangoBackupSpecOperation `json:"export,omitempty"`
}

type ArangoBackupSpecDeployment struct {
	Name string `json:"name,omitempty"`
}

type ArangoBackupSpecOptions struct {
	Timeout *float32 `json:"timeout,omitempty"`
	Force   *bool    `json:"force,omitempty"`
}

type ArangoBackupSpecOperation struct {
	RepositoryURL         string `json:"repositoryURL"`
	CredentialsSecretName string `json:"credentialsSecretName,omitempty"`
}
