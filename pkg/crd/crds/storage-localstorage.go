//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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

package crds

import (
	_ "embed"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// Deprecated: use StorageLocalStorageWithOptions instead
func StorageLocalStorage() *apiextensions.CustomResourceDefinition {
	return StorageLocalStorageWithOptions()
}

func StorageLocalStorageWithOptions(opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition {
	return getCRD(StorageLocalStorageDefinitionData(), opts...)
}

// Deprecated: use StorageLocalStorageDefinitionWithOptions instead
func StorageLocalStorageDefinition() Definition {
	return StorageLocalStorageDefinitionWithOptions()
}

func StorageLocalStorageDefinitionWithOptions(opts ...func(*CRDOptions)) Definition {
	return Definition{
		DefinitionData: StorageLocalStorageDefinitionData(),
		CRD:            StorageLocalStorageWithOptions(opts...),
	}
}

func StorageLocalStorageDefinitionData() DefinitionData {
	return DefinitionData{
		definition:       storageLocalStorage,
		schemaDefinition: storageLocalStorageSchemaRaw,
	}
}

//go:embed storage-localstorage.yaml
var storageLocalStorage []byte

//go:embed storage-localstorage.schema.generated.yaml
var storageLocalStorageSchemaRaw []byte
