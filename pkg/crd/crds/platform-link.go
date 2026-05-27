//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

func PlatformLinkWithOptions(opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition {
	return getCRD(PlatformLinkDefinitionData(), opts...)
}

func PlatformLinkDefinitionWithOptions(opts ...func(*CRDOptions)) Definition {
	return Definition{
		DefinitionData: PlatformLinkDefinitionData(),
		CRD:            PlatformLinkWithOptions(opts...),
	}
}

func PlatformLinkDefinitionData() DefinitionData {
	return DefinitionData{
		definition:       platformLink,
		schemaDefinition: platformLinkSchemaRaw,
	}
}

//go:embed platform-link.yaml
var platformLink []byte

//go:embed platform-link.schema.generated.yaml
var platformLinkSchemaRaw []byte
