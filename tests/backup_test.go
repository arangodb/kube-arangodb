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
// Author Lars Maier
//

package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var backupAPIAvailable *bool

func waitUntilBackup(ci versioned.Interface, name, ns string, predicate func(*api.ArangoBackup, error) error, timeout ...time.Duration) (*api.ArangoBackup, error) {
	var result *api.ArangoBackup
	op := func() error {
		obj, err := ci.DatabaseV1alpha().ArangoBackups(ns).Get(name, metav1.GetOptions{})
		result = obj
		if predicate != nil {
			if err := predicate(obj, err); err != nil {
				return maskAny(err)
			}
		}
		return nil
	}
	actualTimeout := deploymentReadyTimeout
	if len(timeout) > 0 {
		actualTimeout = timeout[0]
	}
	if err := retry.Retry(op, actualTimeout); err != nil {
		return nil, maskAny(err)
	}
	return result, nil
}

func backupIsReady(backup *api.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if backup.Status.State == api.ArangoBackupStateReady {
		return nil
	}

	return fmt.Errorf("Backup not ready - status: %s", backup.Status.State)
}

func backupIsAvailable(backup *api.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if backup.Status.Available {
		return nil
	}

	return fmt.Errorf("Backup not available - status: %s", backup.Status.State)
}

func backupIsNotAvailable(backup *api.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if !backup.Status.Available {
		return nil
	}

	return fmt.Errorf("Backup is still available - status: %s", backup.Status.State)
}

func backupIsNotFound(backup *api.ArangoBackup, err error) error {
	if err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return err
	}

	return fmt.Errorf("Backup resource still exists")
}

type EnsureBackupOptions struct {
	Options  *api.ArangoBackupSpecOptions
	Download *api.ArangoBackupSpecDownload
	Upload   *api.ArangoBackupSpecOperation
}

func newBackup(name, deployment string, options *EnsureBackupOptions) *api.ArangoBackup {
	backup := &api.ArangoBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(name),
		},
		Spec: api.ArangoBackupSpec{
			Deployment: api.ArangoBackupSpecDeployment{
				Name: deployment,
			},
		},
	}

	if options != nil {
		backup.Spec.Options = options.Options
		backup.Spec.Upload = options.Upload
		backup.Spec.Download = options.Download
	}

	return backup
}

func skipIfBackupUnavailable(t *testing.T, client driver.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := client.Backup().List(ctx, nil); err != nil {
		t.Skipf("Backup API not available: %s", err.Error())
	}
}

func statBackupMeta(client driver.Client, backupID driver.BackupID) (bool, driver.BackupMeta, error) {

	list, err := client.Backup().List(nil, &driver.BackupListOptions{ID: backupID})
	if err != nil {
		if driver.IsNotFound(err) {
			return false, driver.BackupMeta{}, nil
		}

		return false, driver.BackupMeta{}, err
	}

	if meta, ok := list[driver.BackupID(backupID)]; ok {
		return true, meta, nil
	}

	return false, driver.BackupMeta{}, fmt.Errorf("List does not contain backup")
}

func ensureBackup(t *testing.T, deployment, ns string, deploymentClient versioned.Interface, predicate func(*api.ArangoBackup, error) error, options *EnsureBackupOptions) (*api.ArangoBackup, string, driver.BackupID) {
	backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), deployment, options)
	_, err := deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Create(backup)
	assert.NoError(t, err, "failed to create backup: %s", err)
	name := backup.GetName()

	backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, predicate)
	assert.NoError(t, err, "backup did not become available: %s", err)
	var backupID string
	if backup.Status.Details != nil {
		backupID = backup.Status.Details.ID
	}
	return backup, name, driver.BackupID(backupID)
}

func skipOrRemotePath(t *testing.T) (repoPath string) {
	repoPath = os.Getenv("TEST_REMOTE_REPOSITORY")
	if repoPath == "" {
		t.Skip("TEST_REMOTE_REPOSITORY not set")
	}
	return repoPath
}

func newOperation() *api.ArangoBackupSpecOperation {
	return &api.ArangoBackupSpecOperation{
		RepositoryURL:         os.Getenv("TEST_REMOTE_REPOSITORY"),
		CredentialsSecretName: testBackupRemoteSecretName,
	}
}

func TestBackupCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster()
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-backup-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.DBServers.Count = util.NewInt(2)
	depl.Spec.Coordinators.Count = util.NewInt(2)
	depl.Spec.SetDefaults(depl.GetName()) // this must be last
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Create deployment
	apiObject, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	defer removeDeployment(deploymentClient, depl.GetName(), ns)
	assert.NoError(t, err, "failed to create deplyment: %s", err)

	_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, deploymentIsReady())
	assert.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))

	ctx := context.Background()
	databaseClient := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	skipIfBackupUnavailable(t, databaseClient)

	t.Run("create backup", func(t *testing.T) {
		backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), depl.GetName(), nil)
		_, err := deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Create(backup)
		assert.NoError(t, err, "failed to create backup: %s", err)
		defer deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(backup.GetName(), &metav1.DeleteOptions{})

		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsAvailable)
		assert.NoError(t, err, "backup did not become available: %s", err)
		backupID := backup.Status.Details.ID

		// check that the backup is actually available
		found, meta, err := statBackupMeta(databaseClient, driver.BackupID(backupID))
		assert.NoError(t, err, "Backup test failed: %s", err)
		assert.True(t, found)
		assert.Equal(t, meta.Version, backup.Status.Details.Version)
	})

	t.Run("create backup and delete", func(t *testing.T) {
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, nil)
		defer deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(name, &metav1.DeleteOptions{})

		// check that the backup is actually available
		found, meta, err := statBackupMeta(databaseClient, id)
		assert.NoError(t, err, "Backup test failed: %s", err)
		assert.True(t, found)
		assert.Equal(t, meta.Version, backup.Status.Details.Version)

		// now remove the backup
		deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(name, &metav1.DeleteOptions{})
		_, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotFound)
		assert.NoError(t, err, "Backup test failed: %s", err)

		// check that the actual backup has been deleted
		found, _, err = statBackupMeta(databaseClient, id)
		assert.False(t, found)
	})

	t.Run("remove backup locally", func(t *testing.T) {
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, nil)
		defer deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(name, &metav1.DeleteOptions{})

		// now remove the backup locally
		err := databaseClient.Backup().Delete(nil, id)
		assert.NoError(t, err, "Failed to delete backup: %s", err)

		// wait for the backup to become unavailable
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotAvailable, 30*time.Second)
		assert.NoError(t, err, "Backup test failed: %s", err)
		assert.Equal(t, api.ArangoBackupStateDeleted, backup.Status.State)
	})

	t.Run("handle existing backups", func(t *testing.T) {
		// create a local backup manually
		id, _, err := databaseClient.Backup().Create(nil, nil)
		assert.NoError(t, err, "Creating backup failed: %s", err)
		found, meta, err := statBackupMeta(databaseClient, driver.BackupID(id))
		assert.NoError(t, err, "Backup test failed: %s", err)
		assert.True(t, found)

		// create a backup resource manually with that id
		backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), depl.GetName(), nil)
		backup.Status.Details = &api.ArangoBackupDetails{ID: string(id), Version: meta.Version}
		_, err = deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Create(backup)
		assert.NoError(t, err, "failed to create backup: %s", err)
		defer deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(backup.GetName(), &metav1.DeleteOptions{})

		// wait until the backup becomes available
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsAvailable)
		assert.NoError(t, err, "backup did not become available: %s", err)
		assert.Equal(t, api.ArangoBackupStateReady, backup.Status.State)
	})

	t.Run("create-restore-cycle", func(t *testing.T) {
		type Book struct {
			Title  string
			Author string
		}

		ctx := context.Background()
		// first add collections, insert data into the cluster
		dbname := "backup-test-db"
		db, err := databaseClient.CreateDatabase(ctx, dbname, nil)
		assert.NoError(t, err, "failed to create database: %s", err)

		colname := "backup-test-col"
		col, err := db.CreateCollection(ctx, colname, nil)
		assert.NoError(t, err, "failed to create collection: %s", err)

		meta1, err := col.CreateDocument(ctx, &Book{Title: "My first Go-Program", Author: "Adam"})
		assert.NoError(t, err, "failed to create document: %s", err)

		// Now create a backup
		_, name, _ := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, nil)
		defer deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(name, &metav1.DeleteOptions{})

		// insert yet another document
		meta2, err := col.CreateDocument(ctx, &Book{Title: "Bad book title", Author: "Lars"})
		assert.NoError(t, err, "failed to create document: %s", err)

		// now restore the backup
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = util.NewString(string(name))
		})
		assert.NoError(t, err, "Failed to update deployment: %s", err)

		_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore != nil {
				result := status.Restore

				if result.RequestedFrom != name {
					return fmt.Errorf("Wrong backup in RequestedFrom: %s, expected %s", result.RequestedFrom, name)
				}

				if !result.Restored {
					t.Fatalf("Failed to restore backup: %s", result.Message)
				}

				return nil
			}

			return fmt.Errorf("Restore is not set on deployment")
		})
		assert.NoError(t, err, "Deployment did not restore in time: %s", err)

		// restore was completed, check if documents are there
		found, err := col.DocumentExists(ctx, meta1.Key)
		assert.NoError(t, err, "Failed to check if document exists: %s", err)
		assert.True(t, found)

		// second document should not exist
		found, err = col.DocumentExists(ctx, meta2.Key)
		assert.NoError(t, err, "Failed to check if document exists: %s", err)
		assert.False(t, found)

		// delete the RestoreFrom entry
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = nil
		})
		assert.NoError(t, err, "Failed to update deployment: %s", err)

		// wait for it to be deleted in the status
		waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore == nil {
				return nil
			}

			return fmt.Errorf("Restore is not set to nil")
		})

	})

	t.Run("restore-nonexistent", func(t *testing.T) {
		// try to restore a backup that doesn not exist
		name := "does-not-exist"

		_, err := updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = util.NewString(name)
		})
		assert.NoError(t, err, "Failed to update deployment: %s", err)

		depl, err := waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore != nil {
				result := status.Restore

				if result.RequestedFrom != name {
					return fmt.Errorf("Wrong backup in RequestedFrom: %s, expected %s", result.RequestedFrom, name)
				}

				if result.Restored {
					t.Fatalf("Backup has been restored!")
				}

				return nil
			}

			return fmt.Errorf("Restore is not set on deployment")
		})
		assert.NoError(t, err, "Deployment did not restore in time: %s", err)
		assert.False(t, depl.Status.Restore.Restored)
	})

	t.Run("upload", func(t *testing.T) {
		skipOrRemotePath(t)

		// create backup with upload operation
		backup, name, _ := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Upload: newOperation()})
		defer deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(name, &metav1.DeleteOptions{})

		// wait until the backup becomes ready
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsReady)
		assert.NoError(t, err, "backup did not become ready: %s", err)

		// TODO there is no way to tell from the status if a upload happened or not
	})

	/*t.Run("create-upload-download-restore-cycle", func(t *testing.T) {
		skipOrRemotePath(t)

		type Book struct {
			Title  string
			Author string
		}

		ctx := context.Background()
		// first add collections, insert data into the cluster
		dbname := "backup-test-db"
		db, err := databaseClient.CreateDatabase(ctx, dbname, nil)
		assert.NoError(t, err, "failed to create database: %s", err)

		colname := "backup-test-col"
		col, err := db.CreateCollection(ctx, colname, nil)
		assert.NoError(t, err, "failed to create collection: %s", err)

		meta1, err := col.CreateDocument(ctx, &Book{Title: "My first Go-Program", Author: "Adam"})
		assert.NoError(t, err, "failed to create document: %s", err)

		// Now create a backup
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Upload: newOperation()})
		defer deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(name, &metav1.DeleteOptions{})

		// insert yet another document
		meta2, err := col.CreateDocument(ctx, &Book{Title: "Bad book title", Author: "Lars"})
		assert.NoError(t, err, "failed to create document: %s", err)

		// now remove the backup locally
		err = databaseClient.Backup().Delete(nil, id)
		assert.NoError(t, err, "Failed to delete backup: %s", err)

		// wait for the backup to become unavailable
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotAvailable, 30*time.Second)
		assert.NoError(t, err, "Backup test failed: %s", err)
		assert.Equal(t, api.ArangoBackupStateDeleted, backup.Status.State)

		// now remove the backup
		deploymentClient.DatabaseV1alpha().ArangoBackups(ns).Delete(name, &metav1.DeleteOptions{})
		_, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotFound)
		assert.NoError(t, err, "Backup test failed: %s", err)

		// create the backup with a download

	})*/
}
