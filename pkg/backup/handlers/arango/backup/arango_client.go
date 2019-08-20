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

	"github.com/arangodb/go-driver"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

type ArangoClientFactory func(deployment *database.ArangoDeployment, backup *database.ArangoBackup) (ArangoBackupClient, error)

type TemporaryErrorInterface interface {
	Temporary() bool
}

type ArangoBackupProgress struct {
	Progress          int
	Failed, Completed bool
	FailMessage       string
}

type ArangoBackupCreateResponse struct {
	driver.BackupMeta
	Forced bool
}

type ArangoBackupClient interface {
	Create() (ArangoBackupCreateResponse, error)
	Get(driver.BackupID) (driver.BackupMeta, error)

	Upload(driver.BackupID) (driver.BackupTransferJobID, error)
	Download(driver.BackupID) (driver.BackupTransferJobID, error)

	Progress(driver.BackupTransferJobID) (ArangoBackupProgress, error)
	Abort(driver.BackupTransferJobID) error

	Exists(driver.BackupID) (bool, error)
	Delete(driver.BackupID) error
}

func NewTemporaryError(format string, a ...interface{}) error {
	return TemporaryError{
		Message: fmt.Sprintf(format, a...),
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

func checkTemporaryError(err error) bool {
	if ok := IsTemporaryError(err); ok {
		return ok
	}

	if _, ok :=err.(TemporaryErrorInterface); ok {
		return ok
	}

	return false
}