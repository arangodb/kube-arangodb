package backup

import (
	"fmt"
	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

type ArangoClientFactory func(deployment *database.ArangoDeployment) (ArangoBackupClient, error)

type ArangoBackupProgress struct {
	Progress int
	Failed, Completed bool
	FailMessage string
}

type ArangoBackupClient interface {
	Create() (driver.BackupMeta, error)
	Get(driver.BackupID) (driver.BackupMeta, error)

	Upload(driver.BackupID) (driver.BackupTransferJobID, error)
	Download(driver.BackupID) (driver.BackupTransferJobID, error)

	Progress(driver.BackupTransferJobID) (ArangoBackupProgress, error)
}

func NewTemporaryError(format string, a... interface{}) error {
	return TemporaryError{
		Message:fmt.Sprintf(format, a...),
	}
}

// TemporaryError defines error which will not update ArangoBackup object status
type TemporaryError struct {
	Message string
}

func (t TemporaryError) Error() string {
	return t.Message
}

func IsTemporaryError(err error) bool {
	_, ok := err.(TemporaryError)
	return ok
}