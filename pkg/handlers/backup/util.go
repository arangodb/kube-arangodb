//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package backup

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	clientBackup "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/backup/v1"
	"github.com/arangodb/kube-arangodb/pkg/handlers/backup/state"
)

type backupStates []*backupApi.ArangoBackup

func (b backupStates) filter(f func(b *backupApi.ArangoBackup) bool) backupStates {
	if f == nil {
		return nil
	}

	r := make(backupStates, 0, len(b))

	for id := range b {
		if f(b[id]) {
			r = append(r, b[id])
		}
	}

	return r
}

type backupStatesCount map[state.State]backupStates

func (b backupStatesCount) get(states ...state.State) backupStates {
	i := 0

	for _, s := range states {
		i += len(b[s])
	}

	if i == 0 {
		return nil
	}

	r := make(backupStates, 0, i)

	for _, s := range states {
		r = append(r, b[s]...)
	}

	return r
}

func countBackupStates(backup *backupApi.ArangoBackup, client clientBackup.ArangoBackupInterface) (backupStatesCount, error) {
	backups, err := client.List(context.Background(), meta.ListOptions{})

	if err != nil {
		return nil, newTemporaryError(err)
	}

	ret := map[state.State]backupStates{}

	for _, existingBackup := range backups.Items {
		// Skip same backup from count
		if existingBackup.Name == backup.Name {
			continue
		}

		// Skip backups which are not on same deployment from count
		if existingBackup.Spec.Deployment.Name != backup.Spec.Deployment.Name {
			continue
		}

		ret[existingBackup.Status.State] = append(ret[existingBackup.Status.State], existingBackup.DeepCopy())
	}

	return ret, nil
}
