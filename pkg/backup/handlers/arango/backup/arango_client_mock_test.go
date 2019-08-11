package backup

import (
	"fmt"
	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sync"
)

const (
	mockVersion = "1.0.0"
)

func newMockArangoClientBackupErrorFactory(err error) ArangoClientFactory {
	return func(deployment *v1alpha.ArangoDeployment) (ArangoBackupClient, error) {
		return nil, err
	}
}

func newMockArangoClientBackupFactory(mock * mockArangoClientBackup) ArangoClientFactory {
	return func(deployment *v1alpha.ArangoDeployment) (ArangoBackupClient, error) {
		return mock, nil
	}
}

func newMockArangoClientBackup(errors mockErrorsArangoClientBackup) *mockArangoClientBackup {
	return &mockArangoClientBackup{
		backups:    map[driver.BackupID]driver.BackupMeta{},
		progresses: map[driver.BackupTransferJobID]ArangoBackupProgress{},
		errors: errors,
	}
}

type mockErrorsArangoClientBackup struct {
	createError, getError, uploadError, downloadError, progressError string
	isTemporaryError bool
}

type mockArangoClientBackupProgress struct {
	progress int
}

type mockArangoClientBackup struct {
	lock sync.Mutex

	backups    map[driver.BackupID]driver.BackupMeta
	progresses map[driver.BackupTransferJobID]ArangoBackupProgress

	errors mockErrorsArangoClientBackup
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

	if meta, ok := m.backups[id]; !ok {
		return driver.BackupMeta{}, fmt.Errorf("not found")
	} else {
		return meta, nil
	}
}

func (m*mockArangoClientBackup) Create() (driver.BackupMeta, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if err := m.newError(m.errors.createError); err != nil {
		return driver.BackupMeta{}, err
	}

	id := driver.BackupID(uuid.NewUUID())

	meta := driver.BackupMeta{
		ID:id,
		Version:mockVersion,
	}

	m.backups[id] = meta

	return meta, nil
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