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
	BackupsBackupPolicyPolicyVersion = driver.Version("1.0.1")
)

func init() {
	mustLoadCRD(backupsBackupPolicy, backupsBackupPolicySchemaRaw, &backupsBackupPolicyCRD, &backupsBackupPolicyCRDWithSchema)
}

func BackupsBackupPolicyPolicy(opts ...GetCRDOptions) *apiextensions.CustomResourceDefinition {
	return getCRD(backupsBackupPolicyCRD, backupsBackupPolicyCRDWithSchema, opts...)
}

func BackupsBackupPolicyDefinition() Definition {
	return Definition{
		Version:       BackupsBackupPolicyPolicyVersion,
		CRD:           backupsBackupPolicyCRD.DeepCopy(),
		CRDWithSchema: backupsBackupPolicyCRDWithSchema.DeepCopy(),
	}
}

var backupsBackupPolicyCRD apiextensions.CustomResourceDefinition
var backupsBackupPolicyCRDWithSchema apiextensions.CustomResourceDefinition

//go:embed backups-backuppolicy.yaml
var backupsBackupPolicy []byte

//go:embed backups-backuppolicy.schema.generated.yaml
var backupsBackupPolicySchemaRaw []byte
