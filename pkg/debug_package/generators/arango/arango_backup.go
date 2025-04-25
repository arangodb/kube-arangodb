//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package arango

import (
	"context"

	v1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func Backup(f shared.FactoryGen) {
	f.AddSection("backup").
		Register("backup", true, shared.WithKubernetesItems[*v1.ArangoBackup](arangoBackupV1ArangoBackupList, shared.WithDefinitions[*v1.ArangoBackup])).
		Register("backuppolicy", true, shared.WithKubernetesItems[*v1.ArangoBackupPolicy](arangoBackupPolicyV1ArangoBackupPolicyList, shared.WithDefinitions[*v1.ArangoBackupPolicy]))
}

func arangoBackupV1ArangoBackupList(ctx context.Context, client kclient.Client, namespace string) ([]*v1.ArangoBackup, error) {
	return list.ListObjects[*v1.ArangoBackupList, *v1.ArangoBackup](ctx, client.Arango().BackupV1().ArangoBackups(namespace), func(result *v1.ArangoBackupList) []*v1.ArangoBackup {
		q := make([]*v1.ArangoBackup, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}

func arangoBackupPolicyV1ArangoBackupPolicyList(ctx context.Context, client kclient.Client, namespace string) ([]*v1.ArangoBackupPolicy, error) {
	return list.ListObjects[*v1.ArangoBackupPolicyList, *v1.ArangoBackupPolicy](ctx, client.Arango().BackupV1().ArangoBackupPolicies(namespace), func(result *v1.ArangoBackupPolicyList) []*v1.ArangoBackupPolicy {
		q := make([]*v1.ArangoBackupPolicy, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
