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

package internal

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	replicationApi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	storageApi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func Test_GenerateAPIDocs(t *testing.T) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	// package path -> result doc file name -> name of the top-level field to be described -> field instance for reflection
	input := map[string]map[string]map[string]interface{}{
		fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
			"ArangoDeployment.V1": {
				"Spec": deploymentApi.ArangoDeployment{}.Spec,
			},
			"ArangoMember.V1": {
				"Spec": deploymentApi.ArangoMember{}.Spec,
			},
		},
		fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
			"ArangoBackup.V1": {
				"Spec":   backupApi.ArangoBackup{}.Spec,
				"Status": backupApi.ArangoBackup{}.Status,
			},
			"ArangoBackupPolicy.V1": {
				"Spec":   backupApi.ArangoBackupPolicy{}.Spec,
				"Status": backupApi.ArangoBackupPolicy{}.Status,
			},
		},
		fmt.Sprintf("%s/pkg/apis/replication/v1", root): {
			"ArangoDeploymentReplication.V1": {
				"Spec": replicationApi.ArangoDeploymentReplication{}.Spec,
			},
		},
		fmt.Sprintf("%s/pkg/apis/storage/v1alpha", root): {
			"ArangoLocalStorage.V1Alpha": {
				"Spec": storageApi.ArangoLocalStorage{}.Spec,
			},
		},
		fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
			"ArangoMLStorage.V1Alpha1": {
				"Spec": mlApi.ArangoMLStorage{}.Spec,
			},
			"ArangoMLExtension.V1Alpha1": {
				"Spec": mlApi.ArangoMLExtension{}.Spec,
			},
		},
	}

	resultPaths := make(map[string]string)
	for apiDir, docs := range input {
		fields, fileSets := parseSourceFiles(t, apiDir)
		util.CopyMap(resultPaths, generateDocs(t, docs, fields, fileSets))
	}
	generateIndex(t, resultPaths)
}
