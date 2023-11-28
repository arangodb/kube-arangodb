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

package crds

import (
	_ "embed"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/go-driver"
)

const (
	DatabaseMemberVersion = driver.Version("1.0.1")
)

func init() {
	mustLoadCRD(databaseMember, databaseMemberSchemaRaw, &databaseMemberCRD, &databaseMemberCRDSchemas)
}

// Deprecated: use DatabaseMemberWithOptions instead
func DatabaseMember() *apiextensions.CustomResourceDefinition {
	return DatabaseMemberWithOptions()
}

func DatabaseMemberWithOptions(opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition {
	return getCRD(databaseMemberCRD, databaseMemberCRDSchemas, opts...)
}

// Deprecated: use DatabaseMemberDefinitionWithOptions instead
func DatabaseMemberDefinition() Definition {
	return DatabaseMemberDefinitionWithOptions()
}

func DatabaseMemberDefinitionWithOptions(opts ...func(*CRDOptions)) Definition {
	return Definition{
		Version: DatabaseMemberVersion,
		CRD:     DatabaseMemberWithOptions(opts...),
	}
}

var databaseMemberCRD apiextensions.CustomResourceDefinition
var databaseMemberCRDSchemas crdSchemas

//go:embed database-member.yaml
var databaseMember []byte

//go:embed database-member.schema.generated.yaml
var databaseMemberSchemaRaw []byte
