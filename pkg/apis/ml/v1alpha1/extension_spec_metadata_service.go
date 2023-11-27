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

import (
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoPipeDatabase           = "arangopipe"
	ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoMLFeatureStoreDatabase = "arangomlfeaturestore"
)

type ArangoMLExtensionSpecMetadataService struct {
	// Local define to use Local ArangoDeployment as the Metadata Service
	Local *ArangoMLExtensionSpecMetadataServiceLocal `json:"local,omitempty"`
}

func (a *ArangoMLExtensionSpecMetadataService) GetLocal() *ArangoMLExtensionSpecMetadataServiceLocal {
	if a == nil || a.Local == nil {
		return nil
	}

	return a.Local
}

func (a *ArangoMLExtensionSpecMetadataService) Validate() error {
	// If Nil then we use default
	if a == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("local", a.GetLocal().Validate()),
	)
}

type ArangoMLExtensionSpecMetadataServiceLocal struct {
	// ArangoPipeDatabase define Database name to be used as MetadataService Backend in ArangoPipe
	// +doc/default: arangopipe
	ArangoPipeDatabase *string `json:"arangoPipeDatabase,omitempty"`

	// ArangoMLFeatureStoreDatabase define Database name to be used as MetadataService Backend in ArangoMLFeatureStoreDatabase
	// +doc/default: arangomlfeaturestore
	ArangoMLFeatureStoreDatabase *string `json:"arangoMLFeatureStore,omitempty"`
}

func (a *ArangoMLExtensionSpecMetadataServiceLocal) GetArangoPipeDatabase() string {
	if a == nil {
		return ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoPipeDatabase
	}

	if d := a.ArangoPipeDatabase; d == nil {
		return ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoPipeDatabase
	} else {
		return *d
	}
}

func (a *ArangoMLExtensionSpecMetadataServiceLocal) GetArangoMLFeatureStoreDatabase() string {
	if a == nil {
		return ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoMLFeatureStoreDatabase
	}

	if d := a.ArangoMLFeatureStoreDatabase; d == nil {
		return ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoMLFeatureStoreDatabase
	} else {
		return *d
	}
}

func (a *ArangoMLExtensionSpecMetadataServiceLocal) Validate() error {
	// If Nil then we use default

	return shared.WithErrors(
		shared.PrefixResourceErrors("arangoPipeDatabase", util.BoolSwitch(a.GetArangoPipeDatabase() != ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoPipeDatabase, errors.Newf("Database name is hardcoded"), nil)),
		shared.PrefixResourceErrors("arangoMLFeatureStore", util.BoolSwitch(a.GetArangoMLFeatureStoreDatabase() != ArangoMLExtensionSpecMetadataServiceLocalDefaultArangoMLFeatureStoreDatabase, errors.Newf("Database name is hardcoded"), nil)),
	)
}
