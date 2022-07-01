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
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/go-driver"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	database "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type arangoClientBackupImpl struct {
	deployment *database.ArangoDeployment
	backup     *backupApi.ArangoBackup
	driver     driver.Client
	kubecli    kubernetes.Interface
}

func newArangoClientBackupFactory(handler *handler) ArangoClientFactory {
	return func(deployment *database.ArangoDeployment, backup *backupApi.ArangoBackup) (ArangoBackupClient, error) {
		ctx := context.Background()
		client, err := arangod.CreateArangodDatabaseClient(ctx, handler.kubeClient.CoreV1(), deployment, false)
		if err != nil {
			return nil, err
		}

		return &arangoClientBackupImpl{
			deployment: deployment,
			backup:     backup,
			driver:     client,
			kubecli:    handler.kubeClient,
		}, nil
	}
}

func (ac *arangoClientBackupImpl) List() (map[driver.BackupID]driver.BackupMeta, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultArangoClientTimeout)
	defer cancel()

	backups, err := ac.driver.Backup().List(ctx, nil)
	if err != nil {
		return nil, err
	}

	return backups, nil
}

func (ac *arangoClientBackupImpl) Create() (ArangoBackupCreateResponse, error) {
	dt := defaultArangoClientTimeout

	co := driver.BackupCreateOptions{}

	if opt := ac.backup.Spec.Options; opt != nil {
		if allowInconsistent := opt.AllowInconsistent; allowInconsistent != nil {
			co.AllowInconsistent = *allowInconsistent
		}
		if timeout := opt.Timeout; timeout != nil {
			co.Timeout = time.Duration(*timeout * float32(time.Second))
			dt += co.Timeout
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), dt)
	defer cancel()

	id, resp, err := ac.driver.Backup().Create(ctx, &co)
	if err != nil {
		return ArangoBackupCreateResponse{}, err
	}

	// Now ask for the version
	meta, err := ac.Get(id)
	if err != nil {
		return ArangoBackupCreateResponse{}, err
	}

	return ArangoBackupCreateResponse{
		PotentiallyInconsistent: resp.PotentiallyInconsistent,
		BackupMeta:              meta,
	}, nil
}

func (ac *arangoClientBackupImpl) Get(backupID driver.BackupID) (driver.BackupMeta, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultArangoClientTimeout)
	defer cancel()

	// list, err := ac.driver.Backup().List(ctx, &driver.BackupListOptions{ID: backupID})
	list, err := ac.driver.Backup().List(ctx, nil)
	if err != nil {
		return driver.BackupMeta{}, err
	}

	meta, ok := list[backupID]

	if ok {
		return meta, nil
	}

	return driver.BackupMeta{}, driver.ArangoError{
		ErrorMessage: fmt.Sprintf("backup %s was not found", backupID),
		Code:         404,
	}
}

func (ac *arangoClientBackupImpl) getCredentialsFromSecret(ctx context.Context, secretName string) (interface{}, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	token, err := k8sutil.GetTokenSecret(ctxChild, ac.kubecli.CoreV1().Secrets(ac.backup.Namespace), secretName)
	if err != nil {
		return nil, err
	}

	var raw json.RawMessage
	if err := json.Unmarshal([]byte(token), &raw); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal credentials: ")
	}

	return raw, nil
}

func (ac *arangoClientBackupImpl) Upload(backupID driver.BackupID) (driver.BackupTransferJobID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultArangoClientTimeout)
	defer cancel()

	uploadSpec := ac.backup.Spec.Upload
	if uploadSpec == nil {
		return "", errors.Newf("upload was called but no upload spec was given")
	}

	cred, err := ac.getCredentialsFromSecret(ctx, uploadSpec.CredentialsSecretName)
	if err != nil {
		return "", err
	}

	return ac.driver.Backup().Upload(ctx, backupID, uploadSpec.RepositoryURL, cred)
}

func (ac *arangoClientBackupImpl) Download(backupID driver.BackupID) (driver.BackupTransferJobID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultArangoClientTimeout)
	defer cancel()

	downloadSpec := ac.backup.Spec.Download
	if downloadSpec == nil {
		return "", errors.Newf("Download was called but not download spec was given")
	}

	cred, err := ac.getCredentialsFromSecret(ctx, downloadSpec.CredentialsSecretName)
	if err != nil {
		return "", err
	}

	return ac.driver.Backup().Download(ctx, backupID, downloadSpec.RepositoryURL, cred)
}

func (ac *arangoClientBackupImpl) Progress(jobID driver.BackupTransferJobID) (ArangoBackupProgress, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultArangoClientTimeout)
	defer cancel()

	report, err := ac.driver.Backup().Progress(ctx, jobID)
	if err != nil {
		return ArangoBackupProgress{}, err
	}

	if report.Cancelled {
		return ArangoBackupProgress{
			Failed:      true,
			FailMessage: "Upload cancelled",
		}, nil
	}

	var ret ArangoBackupProgress
	var completedCount int
	var total int
	var done int

	for _, status := range report.DBServers {
		total += status.Progress.Total
		done += status.Progress.Done

		switch status.Status {
		case driver.TransferFailed:
			ret.Failed = true
			ret.FailMessage = status.ErrorMessage
		case driver.TransferCompleted:
			completedCount++
		case driver.TransferAcknowledged:
		case driver.TransferStarted:
		case "":
			completedCount++
		default:
			return ArangoBackupProgress{}, errors.Newf("Unknown transfere status: %s", status.Status)
		}
	}

	// Check if all defined servers are completed and total number of files is greater than 0 (there is at least 1 file per server)
	ret.Completed = completedCount == len(report.DBServers) && total > 0
	if total != 0 {
		ret.Progress = (100 * done) / total
	}
	return ret, nil
}

func (ac *arangoClientBackupImpl) Exists(backupID driver.BackupID) (bool, error) {
	_, err := ac.Get(backupID)
	if err != nil {
		if driver.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (ac *arangoClientBackupImpl) Delete(backupID driver.BackupID) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultArangoClientTimeout)
	defer cancel()

	return ac.driver.Backup().Delete(ctx, backupID)
}

func (ac *arangoClientBackupImpl) Abort(jobID driver.BackupTransferJobID) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultArangoClientTimeout)
	defer cancel()

	return ac.driver.Backup().Abort(ctx, jobID)
}
