//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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

func PlatformServiceWithOptions(opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition {
	return getCRD(PlatformServiceDefinitionData(), opts...)
}

func PlatformServiceDefinitionWithOptions(opts ...func(*CRDOptions)) Definition {
	return Definition{
		DefinitionData: PlatformServiceDefinitionData(),
		CRD:            PlatformServiceWithOptions(opts...),
	}
}

func PlatformServiceDefinitionData() DefinitionData {
	return DefinitionData{
		definition:       platformService,
		schemaDefinition: platformServiceSchemaRaw,
	}
}

//go:embed platform-service.yaml
var platformService []byte

//go:embed platform-service.schema.generated.yaml
var platformServiceSchemaRaw []byte
