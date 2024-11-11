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

func PlatformChartWithOptions(opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition {
	return getCRD(PlatformChartDefinitionData(), opts...)
}

func PlatformChartDefinitionWithOptions(opts ...func(*CRDOptions)) Definition {
	return Definition{
		DefinitionData: PlatformChartDefinitionData(),
		CRD:            PlatformChartWithOptions(opts...),
	}
}

func PlatformChartDefinitionData() DefinitionData {
	return DefinitionData{
		definition:       platformChart,
		schemaDefinition: platformChartSchemaRaw,
	}
}

//go:embed platform-chart.yaml
var platformChart []byte

//go:embed platform-chart.schema.generated.yaml
var platformChartSchemaRaw []byte
