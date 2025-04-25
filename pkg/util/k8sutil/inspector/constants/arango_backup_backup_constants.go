//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package constants

import (
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	v1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
)

// ArangoBackup
const (
	ArangoBackupGroup           = backup.ArangoBackupGroupName
	ArangoBackupResource        = backup.ArangoBackupResourcePlural
	ArangoBackupKind            = backup.ArangoBackupResourceKind
	ArangoBackupVersionV1Alpha1 = v1.ArangoBackupVersion
)

func init() {
	register[*v1.ArangoBackup](ArangoBackupGKv1(), ArangoBackupGRv1())
}

func ArangoBackupGK() schema.GroupKind {
	return schema.GroupKind{
		Group: ArangoBackupGroup,
		Kind:  ArangoBackupKind,
	}
}

func ArangoBackupGKv1() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ArangoBackupGroup,
		Kind:    ArangoBackupKind,
		Version: ArangoBackupVersionV1Alpha1,
	}
}

func ArangoBackupGR() schema.GroupResource {
	return schema.GroupResource{
		Group:    ArangoBackupGroup,
		Resource: ArangoBackupResource,
	}
}

func ArangoBackupGRv1() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    ArangoBackupGroup,
		Resource: ArangoBackupResource,
		Version:  ArangoBackupVersionV1Alpha1,
	}
}
