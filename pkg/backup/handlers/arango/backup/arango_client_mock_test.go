//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package backup

import (
	"fmt"
	"sync"

	"github.com/arangodb/go-driver"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"k8s.io/apimachinery/pkg/util/uuid"
)

const (
	mockVersion = "1.0.0"
)

func newMockArangoClientBackupErrorFactory(err error) ArangoClientFactory {
	return func(deployment *database.ArangoDeployment, backup *backupApi.ArangoBackup) (ArangoBackupClient, error) {
		return nil, err
	}
}

func newMockArangoClientBackupFactory(mock *mockArangoClientBackup) ArangoClientFactory {
	return func(deployment *database.ArangoDeployment, backup *backupApi.ArangoBackup) (ArangoBackupClient, error) {
		return mock, nil
	}
}

func newMockArangoClientBackup(errors mockErrorsArangoClientBackup) *mockArangoClientBackup {
	return &mockArangoClientBackup{
		backups:    map[driver.BackupID]driver.BackupMeta{},
		progresses: map[driver.BackupTransferJobID]ArangoBackupProgress{},
		errors:     errors,
	}
}

type mockErrorsArangoClientBackup struct {
	createError, listError, getError, uploadError, downloadError, progressError, existsError, deleteError, abortError string
	isTemporaryError                                                                                                  bool

	createForced bool
}

type mockArangoClientBackup struct {
	lock sync.Mutex

	backups    map[driver.BackupID]driver.BackupMeta
	progresses map[driver.BackupTransferJobID]ArangoBackupProgress

	errors mockErrorsArangoClientBackup
}

func (m *mockArangoClientBackup) List() (map[driver.BackupID]driver.BackupMeta, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.listError); err != nil {
		return nil, err
	}

	return m.backups, nil
}

func (m *mockArangoClientBackup) Abort(d driver.BackupTransferJobID) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.abortError); err != nil {
		return err
	}

	if _, ok := m.progresses[d]; ok {
		delete(m.progresses, d)
	}

	return nil
}

func (m *mockArangoClientBackup) Exists(id driver.BackupID) (bool, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.existsError); err != nil {
		return false, err
	}

	_, ok := m.backups[id]

	return ok, nil
}

func (m *mockArangoClientBackup) Delete(id driver.BackupID) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.deleteError); err != nil {
		return err
	}

	if _, ok := m.backups[id]; ok {
		delete(m.backups, id)
	}

	return nil
}

func (m *mockArangoClientBackup) newError(msg string) error {
	if msg == "" {
		return nil
	}

	if m.errors.isTemporaryError {
		return NewTemporaryError(msg)
	}
	return fmt.Errorf(msg)
}

func (m *mockArangoClientBackup) Download(driver.BackupID) (driver.BackupTransferJobID, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.downloadError); err != nil {
		return "", err
	}

	id := driver.BackupTransferJobID(uuid.NewUUID())

	m.progresses[id] = ArangoBackupProgress{}

	return id, nil
}

func (m *mockArangoClientBackup) Progress(id driver.BackupTransferJobID) (ArangoBackupProgress, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.progressError); err != nil {
		return ArangoBackupProgress{}, err
	}

	return m.progresses[id], nil
}

func (m *mockArangoClientBackup) Upload(driver.BackupID) (driver.BackupTransferJobID, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.uploadError); err != nil {
		return "", err
	}

	id := driver.BackupTransferJobID(uuid.NewUUID())

	m.progresses[id] = ArangoBackupProgress{}

	return id, nil
}

func (m *mockArangoClientBackup) Get(id driver.BackupID) (driver.BackupMeta, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.getError); err != nil {
		return driver.BackupMeta{}, err
	}

	if meta, ok := m.backups[id]; ok {
		return meta, nil
	}

	return driver.BackupMeta{}, fmt.Errorf("not found")
}

func (m *mockArangoClientBackup) Create() (ArangoBackupCreateResponse, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.createError); err != nil {
		return ArangoBackupCreateResponse{}, err
	}

	id := driver.BackupID(uuid.NewUUID())

	meta := driver.BackupMeta{
		ID:      id,
		Version: mockVersion,
	}

	m.backups[id] = meta

	return ArangoBackupCreateResponse{
		BackupMeta:              meta,
		PotentiallyInconsistent: m.errors.createForced,
	}, nil
}

func (m *mockArangoClientBackup) getIDs() []string {
	ret := make([]string, 0, len(m.backups))

	for key := range m.backups {
		ret = append(ret, string(key))
	}

	return ret
}

func (m *mockArangoClientBackup) getProgressIDs() []string {
	ret := make([]string, 0, len(m.progresses))

	for key := range m.progresses {
		ret = append(ret, string(key))
	}

	return ret
}

var _ ArangoBackupClient = &mockArangoClientBackup{}
