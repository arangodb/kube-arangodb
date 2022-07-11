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
	"math/rand"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	mockVersion = "1.0.0"
)

func newMockArangoClientBackupErrorFactory(err error) ArangoClientFactory {
	return func(deployment *database.ArangoDeployment, backup *backupApi.ArangoBackup) (ArangoBackupClient, error) {
		return nil, err
	}
}

func newMockArangoClientBackupFactory(mock *mockArangoClientBackupState) ArangoClientFactory {
	return func(deployment *database.ArangoDeployment, backup *backupApi.ArangoBackup) (ArangoBackupClient, error) {
		return &mockArangoClientBackup{
			backup: backup,
			state:  mock,
		}, nil
	}
}

func newMockArangoClientBackup(errors mockErrorsArangoClientBackup) *mockArangoClientBackupState {
	return &mockArangoClientBackupState{
		backups:    map[driver.BackupID]driver.BackupMeta{},
		progresses: map[driver.BackupTransferJobID]ArangoBackupProgress{},
		errors:     errors,
	}
}

type mockErrorsArangoClientBackup struct {
	createError, listError, getError, uploadError, downloadError, progressError, existsError, deleteError, abortError error
}

type mockArangoClientBackupState struct {
	lock sync.Mutex

	backups    map[driver.BackupID]driver.BackupMeta
	progresses map[driver.BackupTransferJobID]ArangoBackupProgress

	errors mockErrorsArangoClientBackup
}

type mockArangoClientBackup struct {
	backup *backupApi.ArangoBackup
	state  *mockArangoClientBackupState
}

func (m *mockArangoClientBackup) List() (map[driver.BackupID]driver.BackupMeta, error) {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.listError != nil {
		return nil, m.state.errors.listError
	}

	return m.state.backups, nil
}

func (m *mockArangoClientBackup) Abort(d driver.BackupTransferJobID) error {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.abortError != nil {
		return m.state.errors.abortError
	}

	delete(m.state.progresses, d)

	return nil
}

func (m *mockArangoClientBackup) Exists(id driver.BackupID) (bool, error) {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.existsError != nil {
		return false, m.state.errors.existsError
	}

	_, ok := m.state.backups[id]

	return ok, nil
}

func (m *mockArangoClientBackup) Delete(id driver.BackupID) error {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.deleteError != nil {
		return m.state.errors.deleteError
	}

	delete(m.state.backups, id)

	return nil
}

func (m *mockArangoClientBackup) Download(driver.BackupID) (driver.BackupTransferJobID, error) {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.downloadError != nil {
		return "", m.state.errors.downloadError
	}

	id := driver.BackupTransferJobID(uuid.NewUUID())

	m.state.progresses[id] = ArangoBackupProgress{}

	return id, nil
}

func (m *mockArangoClientBackup) Progress(id driver.BackupTransferJobID) (ArangoBackupProgress, error) {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.progressError != nil {
		return ArangoBackupProgress{}, m.state.errors.progressError
	}

	return m.state.progresses[id], nil
}

func (m *mockArangoClientBackup) Upload(driver.BackupID) (driver.BackupTransferJobID, error) {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.uploadError != nil {
		return "", m.state.errors.uploadError
	}

	id := driver.BackupTransferJobID(uuid.NewUUID())

	m.state.progresses[id] = ArangoBackupProgress{}

	return id, nil
}

func (m *mockArangoClientBackup) Get(id driver.BackupID) (driver.BackupMeta, error) {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.getError != nil {
		return driver.BackupMeta{}, m.state.errors.getError
	}

	if meta, ok := m.state.backups[id]; ok {
		return meta, nil
	}

	return driver.BackupMeta{}, errors.Newf("not found")
}

func (m *mockArangoClientBackup) Create() (ArangoBackupCreateResponse, error) {
	m.state.lock.Lock()
	defer m.state.lock.Unlock()

	if m.state.errors.createError != nil {
		return ArangoBackupCreateResponse{}, m.state.errors.createError
	}

	id := driver.BackupID(uuid.NewUUID())

	inconsistent := false

	if m.backup != nil {
		if m.backup.Spec.Options != nil {
			if m.backup.Spec.Options.AllowInconsistent != nil {
				inconsistent = *m.backup.Spec.Options.AllowInconsistent
			}
		}
	}

	servers := uint(rand.Uint32())

	meta := driver.BackupMeta{
		ID:                      id,
		Version:                 mockVersion,
		NumberOfDBServers:       servers,
		DateTime:                time.Now(),
		SizeInBytes:             rand.Uint64(),
		PotentiallyInconsistent: inconsistent,
		NumberOfFiles:           uint(rand.Uint32()),
		NumberOfPiecesPresent:   servers,
		Available:               true,
	}

	m.state.backups[id] = meta

	return ArangoBackupCreateResponse{
		BackupMeta:              meta,
		PotentiallyInconsistent: meta.PotentiallyInconsistent,
	}, nil
}

func (m *mockArangoClientBackup) getIDs() []string {
	ret := make([]string, 0, len(m.state.backups))

	for key := range m.state.backups {
		ret = append(ret, string(key))
	}

	return ret
}

func (m *mockArangoClientBackup) getProgressIDs() []string {
	ret := make([]string, 0, len(m.state.progresses))

	for key := range m.state.progresses {
		ret = append(ret, string(key))
	}

	return ret
}

var _ ArangoBackupClient = &mockArangoClientBackup{}
