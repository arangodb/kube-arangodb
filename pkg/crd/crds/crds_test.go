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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	"github.com/arangodb/kube-arangodb/pkg/apis/replication"
	"github.com/arangodb/kube-arangodb/pkg/apis/storage"
)

func ensureCRDCompliance(t *testing.T, name string, def Definition) {
	t.Run(name, func(t *testing.T) {
		require.Equal(t, name, def.CRD.GetName())
	})
}

func Test_CRD(t *testing.T) {
	ensureCRDCompliance(t, apps.ArangoJobCRDName, AppsJobDefinition())
	ensureCRDCompliance(t, backup.ArangoBackupCRDName, BackupsBackupDefinition())
	ensureCRDCompliance(t, backup.ArangoBackupPolicyCRDName, BackupsBackupPolicyDefinition())
	ensureCRDCompliance(t, deployment.ArangoClusterSynchronizationCRDName, DatabaseClusterSynchronizationDefinition())
	ensureCRDCompliance(t, deployment.ArangoDeploymentCRDName, DatabaseDeploymentDefinition())
	ensureCRDCompliance(t, deployment.ArangoMemberCRDName, DatabaseMemberDefinition())
	ensureCRDCompliance(t, deployment.ArangoTaskCRDName, DatabaseTaskDefinition())
	ensureCRDCompliance(t, replication.ArangoDeploymentReplicationCRDName, ReplicationDeploymentReplicationDefinition())
	ensureCRDCompliance(t, storage.ArangoLocalStorageCRDName, StorageLocalStorageDefinition())
}
