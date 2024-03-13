//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package kubernetes

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/cli"
	"github.com/arangodb/kube-arangodb/pkg/debug_package/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func backupBackups(logger zerolog.Logger, files chan<- shared.File, client kclient.Client) error {
	backups, err := listBackupBackups(client)
	if err != nil {
		if kerrors.IsForbiddenOrNotFound(err) {
			return nil
		}

		return err
	}

	if err := errors.ExecuteWithErrorArrayP2(backupBackup, client, files, backups...); err != nil {
		logger.Err(err).Msgf("Error while collecting arango ml batchjobs")
		return err
	}

	return nil
}

func backupBackup(client kclient.Client, files chan<- shared.File, ext *backupApi.ArangoBackup) error {
	files <- shared.NewYAMLFile(fmt.Sprintf("kubernetes/arango/backupBackup/backups/%s.yaml", ext.GetName()), func() ([]interface{}, error) {
		return []interface{}{ext}, nil
	})

	return nil
}

func listBackupBackups(client kclient.Client) ([]*backupApi.ArangoBackup, error) {
	return ListObjects[*backupApi.ArangoBackupList, *backupApi.ArangoBackup](context.Background(), client.Arango().BackupV1().ArangoBackups(cli.GetInput().Namespace), func(result *backupApi.ArangoBackupList) []*backupApi.ArangoBackup {
		q := make([]*backupApi.ArangoBackup, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
}
