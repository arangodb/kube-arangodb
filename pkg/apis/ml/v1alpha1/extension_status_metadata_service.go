//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package v1alpha1

import sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"

type ArangoMLExtensionStatusMetadataService struct {
	// Local define the Local ArangoDeployment Metadata Service configuration
	Local *ArangoMLExtensionStatusMetadataServiceLocal `json:"local,omitempty"`

	// Secret define the Secret specification to store all the details
	Secret *sharedApi.Object `json:"secret,omitempty"`
}

type ArangoMLExtensionStatusMetadataServiceLocal struct {
	// ArangoPipeDatabase define Database name to be used as MetadataService Backend
	ArangoPipeDatabase string `json:"arangoPipe"`

	// ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend
	ArangoMLFeatureStoreDatabase string `json:"arangoMLFeatureStore,omitempty"`
}
