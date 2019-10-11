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
	"github.com/arangodb/kube-arangodb/pkg/backup/utils"
	"os"
	"strings"
	"testing"
	"time"

	backupClient "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/backup/v1alpha"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/go-driver"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/dchest/uniuri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var backupAPIAvailable *bool

func waitUntilBackup(ci versioned.Interface, name, ns string, predicate func(*backupApi.ArangoBackup, error) error, timeout ...time.Duration) (*backupApi.ArangoBackup, error) {
	var result *backupApi.ArangoBackup
	op := func() error {
		obj, err := ci.BackupV1alpha().ArangoBackups(ns).Get(name, metav1.GetOptions{})
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

func backupIsReady(backup *backupApi.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if backup.Status.State == backupApi.ArangoBackupStateReady {
		return nil
	}

	return fmt.Errorf("Backup not ready - status: %s", backup.Status.State)
}

func backupIsUploaded(backup *backupApi.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if backup.Status.Backup.Uploaded != nil && *backup.Status.Backup.Uploaded {
		return nil
	}

	return fmt.Errorf("Backup not ready - status: %s", backup.Status.State)
}

func backupIsNotUploaded(backup *backupApi.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if backup.Status.Backup.Uploaded == nil || !*backup.Status.Backup.Uploaded {
		return nil
	}

	return fmt.Errorf("Backup not ready - status: %s", backup.Status.State)
}

func backupIsAvailable(backup *backupApi.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if backup.Status.Available {
		return nil
	}

	return fmt.Errorf("Backup not available - status: %s", backup.Status.State)
}

func backupIsNotAvailable(backup *backupApi.ArangoBackup, err error) error {
	if err != nil {
		return err
	}

	if !backup.Status.Available {
		return nil
	}

	return fmt.Errorf("Backup is still available - status: %s", backup.Status.State)
}

func backupIsNotFound(backup *backupApi.ArangoBackup, err error) error {
	if err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return err
	}

	return fmt.Errorf("Backup resource still exists")
}

type EnsureBackupOptions struct {
	Options  *backupApi.ArangoBackupSpecOptions
	Download *backupApi.ArangoBackupSpecDownload
	Upload   *backupApi.ArangoBackupSpecOperation
}

func newBackup(name, deployment string, options *EnsureBackupOptions) *backupApi.ArangoBackup {
	backup := &backupApi.ArangoBackup{
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(name),
			Finalizers: []string{
				backupApi.FinalizerArangoBackup,
			},
		},
		Spec: backupApi.ArangoBackupSpec{
			Deployment: backupApi.ArangoBackupSpecDeployment{
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

func newBackupPolicy(name, schedule string, labels map[string]string, options *EnsureBackupOptions) *backupApi.ArangoBackupPolicy {
	policy := &backupApi.ArangoBackupPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:   strings.ToLower(name),
			Labels: labels,
		},
		Spec: backupApi.ArangoBackupPolicySpec{
			DeploymentSelector: &metav1.LabelSelector{
				MatchLabels: labels,
			},

			Schedule: schedule,
		},
	}

	if options != nil {
		policy.Spec.BackupTemplate.Options = options.Options
		policy.Spec.BackupTemplate.Upload = options.Upload
	}

	return policy
}

func skipIfBackupUnavailable(t *testing.T, client driver.Client) {
	err := utils.Retry(10, time.Second, func() error {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if _, err := client.Backup().List(ctx, nil); err != nil {
			t.Logf("Backup API not yet ready: %s", err.Error())
			return err
		}

		return nil
	})

	if err != nil {
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

func ensureBackup(t *testing.T, deployment, ns string, deploymentClient versioned.Interface, predicate func(*backupApi.ArangoBackup, error) error, options *EnsureBackupOptions) (*backupApi.ArangoBackup, string, driver.BackupID) {
	backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), deployment, options)
	_, err := deploymentClient.BackupV1alpha().ArangoBackups(ns).Create(backup)
	require.NoError(t, err, "failed to create backup: %s", err)
	name := backup.GetName()

	backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, predicate)
	require.NoError(t, err, "backup did not become available: %s", err)
	var backupID string
	if backup.Status.Backup != nil {
		backupID = backup.Status.Backup.ID
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

func newOperation() *backupApi.ArangoBackupSpecOperation {
	return &backupApi.ArangoBackupSpecOperation{
		RepositoryURL:         os.Getenv("TEST_REMOTE_REPOSITORY"),
		CredentialsSecretName: testBackupRemoteSecretName,
	}
}

func newDownload(ID string) *backupApi.ArangoBackupSpecDownload {
	return &backupApi.ArangoBackupSpecDownload{
		ArangoBackupSpecOperation: backupApi.ArangoBackupSpecOperation{
			RepositoryURL:         os.Getenv("TEST_REMOTE_REPOSITORY"),
			CredentialsSecretName: testBackupRemoteSecretName,
		},
		ID: ID,
	}
}

func timeoutWaitForBackups(t *testing.T, backupClient backupClient.ArangoBackupInterface, labels metav1.LabelSelector, size int) func() error {
	return func() error {
		backups, err := backupClient.List(metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&labels)})
		if err != nil {
			return err
		}

		require.Len(t, backups.Items, size)

		done := 0

		for _, backup := range backups.Items {
			switch backup.Status.State {
			case backupApi.ArangoBackupStateFailed:
				log.Error().Str("backup", backup.Name).Str("Message", backup.Status.Message).Msg("Failed")
				require.Fail(t, "Backup object failed", backup.Status.Message)
			case backupApi.ArangoBackupStateReady:
				done++
			}
		}

		log.Info().Int("expected", size).Int("done", done).Msg("Iteration")

		if done == size {
			return interrupt{}
		}

		return nil
	}
}

func compareBackup(t *testing.T, meta driver.BackupMeta, backup *backupApi.ArangoBackup) {
	require.NotNil(t, backup.Status.Backup)
	require.Equal(t, meta.Version, backup.Status.Backup.Version)
	require.True(t, meta.SizeInBytes > 0)
	require.True(t, meta.NumberOfDBServers == 2)
	require.True(t, meta.SizeInBytes == backup.Status.Backup.SizeInBytes)
	require.True(t, meta.NumberOfDBServers == backup.Status.Backup.NumberOfDBServers)
}

func TestBackupCluster(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	deploymentClient := kubeArangoClient.MustNewInCluster()
	ns := getNamespace(t)

	backupPolicyClient := deploymentClient.BackupV1alpha().ArangoBackupPolicies(ns)
	backupClient := deploymentClient.BackupV1alpha().ArangoBackups(ns)

	cmd := []string{
		"--backup.api-enabled=jwt",
	}

	// Prepare deployment config
	deplLabels := map[string]string{
		"COMMON": "1",
		"TEST":   "1",
	}

	depl := newDeployment("test-backup-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.DBServers.Count = util.NewInt(2)
	depl.Spec.DBServers.Args = cmd
	depl.Spec.Coordinators.Count = util.NewInt(2)
	depl.Spec.Coordinators.Args = cmd
	depl.Spec.Agents.Args = cmd
	depl.Spec.SetDefaults(depl.GetName()) // this must be last
	depl.Labels = deplLabels
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Prepare deployment config
	depl2Labels := map[string]string{
		"COMMON": "1",
		"TEST":   "2",
	}

	depl2 := newDeployment("test-backup-two-" + uniuri.NewLen(4))
	depl2.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl2.Spec.DBServers.Count = util.NewInt(2)
	depl2.Spec.DBServers.Args = cmd
	depl2.Spec.Coordinators.Count = util.NewInt(2)
	depl2.Spec.Coordinators.Args = cmd
	depl2.Spec.Agents.Args = cmd
	depl2.Spec.SetDefaults(depl2.GetName()) // this must be last
	depl2.Labels = depl2Labels
	defer deferedCleanupDeployment(c, depl2.GetName(), ns)

	// Create deployment
	apiObject, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	defer removeDeployment(deploymentClient, depl.GetName(), ns)
	require.NoError(t, err, "failed to create deployment: %s", err)

	api2Object, err := deploymentClient.DatabaseV1alpha().ArangoDeployments(ns).Create(depl2)
	defer removeDeployment(deploymentClient, depl2.GetName(), ns)
	require.NoError(t, err, "failed to create deployment two: %s", err)

	_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))

	_, err = waitUntilDeployment(deploymentClient, depl2.GetName(), ns, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment two not running in time: %s", err))

	ctx := context.Background()
	databaseClient := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	database2Client := mustNewArangodDatabaseClient(ctx, kubecli, api2Object, t, nil)

	skipIfBackupUnavailable(t, databaseClient)
	skipIfBackupUnavailable(t, database2Client)

	deployments := []*api.ArangoDeployment{depl, depl2}
	databaseClients := map[*api.ArangoDeployment]driver.Client{
		depl:  databaseClient,
		depl2: database2Client,
	}

	t.Run("create-backups-on-multiple-databases", func(t *testing.T) {
		size := 8
		expected := size * len(deployments)
		labels := metav1.LabelSelector{
			MatchLabels: map[string]string{
				"type": string(uuid.NewUUID()),
			},
		}

		for id := 0; id < size; id++ {
			for _, deployment := range deployments {
				backup := newBackup(fmt.Sprintf("my-backup-%s-%s", deployment.GetName(), uniuri.NewLen(4)), deployment.GetName(), nil)

				backup.Labels = labels.MatchLabels

				_, err := backupClient.Create(backup)
				require.NoError(t, err, "failed to create backup: %s", err)
				defer backupClient.Delete(backup.GetName(), &metav1.DeleteOptions{})
			}
		}

		err := timeout(time.Second, 30*time.Minute, timeoutWaitForBackups(t, backupClient, labels, expected))
		require.NoError(t, err)
	})

	t.Run("create backup", func(t *testing.T) {
		backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), depl.GetName(), nil)
		_, err := backupClient.Create(backup)
		require.NoError(t, err, "failed to create backup: %s", err)
		defer backupClient.Delete(backup.GetName(), &metav1.DeleteOptions{})

		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsAvailable)
		require.NoError(t, err, "backup did not become available: %s", err)
		backupID := backup.Status.Backup.ID

		// check that the backup is actually available
		found, meta, err := statBackupMeta(databaseClient, driver.BackupID(backupID))
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		compareBackup(t, meta, backup)
	})

	t.Run("create-upload backup", func(t *testing.T) {
		backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), depl.GetName(), nil)
		_, err := backupClient.Create(backup)
		require.NoError(t, err, "failed to create backup: %s", err)
		defer backupClient.Delete(backup.GetName(), &metav1.DeleteOptions{})

		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsReady)
		require.NoError(t, err, "backup did not become available: %s", err)
		backupID := backup.Status.Backup.ID

		// check that the backup is actually available
		found, meta, err := statBackupMeta(databaseClient, driver.BackupID(backupID))
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		compareBackup(t, meta, backup)
		require.Nil(t, backup.Status.Backup.Uploaded)
		require.Nil(t, backup.Status.Backup.Downloaded)

		t.Logf("Add upload")
		// add upload part
		currentBackup, err := backupClient.Get(backup.Name, metav1.GetOptions{})
		require.NoError(t, err)

		currentBackup.Spec.Upload = newOperation()

		_, err = backupClient.Update(currentBackup)
		require.NoError(t, err)

		// After backup went thru uploading phase wait for finnish
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

		found, meta, err = statBackupMeta(databaseClient, driver.BackupID(backupID))
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		compareBackup(t, meta, backup)
		require.NotNil(t, backup.Status.Backup.Uploaded, "Upload flag is nil")
		require.Nil(t, backup.Status.Backup.Downloaded)
	})

	t.Run("create backup and delete", func(t *testing.T) {
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, nil)
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// check that the backup is actually available
		found, meta, err := statBackupMeta(databaseClient, id)
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		compareBackup(t, meta, backup)

		// now remove the backup
		backupClient.Delete(name, &metav1.DeleteOptions{})
		_, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotFound)
		require.NoError(t, err, "Backup test failed: %s", err)

		// check that the actual backup has been deleted
		found, _, err = statBackupMeta(databaseClient, id)
		require.False(t, found)
	})

	t.Run("remove backup locally", func(t *testing.T) {
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, nil)
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// now remove the backup locally
		err := databaseClient.Backup().Delete(nil, id)
		require.NoError(t, err, "Failed to delete backup: %s", err)

		// wait for the backup to become unavailable
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotAvailable, 30*time.Second)
		require.NoError(t, err, "Backup test failed: %s", err)
		require.Equal(t, backupApi.ArangoBackupStateDeleted, backup.Status.State)
	})

	t.Run("handle existing backups", func(t *testing.T) {
		// create a local backup manually
		id, _, err := databaseClient.Backup().Create(nil, nil)
		require.NoError(t, err, "Creating backup failed: %s", err)
		found, meta, err := statBackupMeta(databaseClient, driver.BackupID(id))
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)

		// create a backup resource manually with that id
		var backup *backupApi.ArangoBackup
		err = timeout(3*time.Second, 2*time.Minute, func() error {
			backups, err := backupClient.List(metav1.ListOptions{})
			if err != nil {
				return err
			}

			if len(backups.Items) == 0 {
				return nil
			}

			if len(backups.Items) > 1 {
				return fmt.Errorf("Too many backups")
			}

			backup = &backups.Items[0]

			return interrupt{}
		})
		require.NoError(t, err, "failed to create backup: %s", err)
		defer backupClient.Delete(backup.GetName(), &metav1.DeleteOptions{})

		// wait until the backup becomes available
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsAvailable)
		require.NoError(t, err, "backup did not become available: %s", err)
		require.Equal(t, backupApi.ArangoBackupStateReady, backup.Status.State)
		compareBackup(t, meta, backup)
		require.NotNil(t, backup.Status.Backup.Imported)
		require.True(t, *backup.Status.Backup.Imported)
	})

	t.Run("create-multiple-restore-cycle", func(t *testing.T) {
		type Book struct {
			Title  string
			Author string
		}

		ctx := context.Background()
		// first add collections, insert data into the cluster
		dbname := "backup-test-db-two"
		db, err := databaseClient.CreateDatabase(ctx, dbname, nil)
		require.NoError(t, err, "failed to create database: %s", err)

		colname := "backup-test-col"
		col, err := db.CreateCollection(ctx, colname, nil)
		require.NoError(t, err, "failed to create collection: %s", err)

		meta1, err := col.CreateDocument(ctx, &Book{Title: "My first Go-Program", Author: "Adam"})
		require.NoError(t, err, "failed to create document: %s", err)

		// Now create a backups, a lot of them
		size := 8
		labels := metav1.LabelSelector{
			MatchLabels: map[string]string{
				"type": string(uuid.NewUUID()),
			},
		}

		for id := 0; id < size; id++ {
			backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), depl.GetName(), nil)

			backup.Labels = labels.MatchLabels

			_, err := backupClient.Create(backup)
			require.NoError(t, err, "failed to create backup: %s", err)
			defer backupClient.Delete(backup.GetName(), &metav1.DeleteOptions{})
		}

		err = timeout(time.Second, 5*time.Minute, timeoutWaitForBackups(t, backupClient, labels, size))

		require.NoError(t, err)

		// Get first backup
		backups, err := backupClient.List(metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&labels)})
		require.NoError(t, err)
		require.Len(t, backups.Items, size)

		// Create backup from which we are gonna restore
		backup := newBackup(fmt.Sprintf("my-backup-%s", uniuri.NewLen(4)), depl.GetName(), nil)

		backup.Labels = labels.MatchLabels

		_, err = backupClient.Create(backup)
		require.NoError(t, err, "failed to create backup: %s", err)
		defer backupClient.Delete(backup.GetName(), &metav1.DeleteOptions{})

		name := backup.Name

		err = timeout(time.Second, 5*time.Minute, timeoutWaitForBackups(t, backupClient, labels, size+1))

		// insert yet another document
		meta2, err := col.CreateDocument(ctx, &Book{Title: "Bad book title", Author: "Lars"})
		require.NoError(t, err, "failed to create document: %s", err)

		// now restore the backup
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = util.NewString(string(name))
		})
		require.NoError(t, err, "Failed to update deployment: %s", err)

		_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore != nil {
				result := status.Restore

				if result.RequestedFrom != name {
					return fmt.Errorf("Wrong backup in RequestedFrom: %s, expected %s", result.RequestedFrom, name)
				}

				if result.State == api.DeploymentRestoreStateRestoreFailed {
					t.Fatalf("Failed to restore backup: %s", result.Message)
				}

				if result.State == api.DeploymentRestoreStateRestored {
					return nil
				}

				return fmt.Errorf("Not yet restored - staate %s", result.State)
			}

			return fmt.Errorf("Restore is not set on deployment")
		})
		require.NoError(t, err, "Deployment did not restore in time: %s", err)

		// restore was completed, check if documents are there
		found, err := col.DocumentExists(ctx, meta1.Key)
		require.NoError(t, err, "Failed to check if document exists: %s", err)
		require.True(t, found)

		// second document should not exist
		found, err = col.DocumentExists(ctx, meta2.Key)
		require.NoError(t, err, "Failed to check if document exists: %s", err)
		require.False(t, found)

		// delete the RestoreFrom entry
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = nil
		})
		require.NoError(t, err, "Failed to update deployment: %s", err)

		// wait for it to be deleted in the status
		waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore == nil {
				return nil
			}

			return fmt.Errorf("Restore is not set to nil")
		})

		// Assert that all of the backups are in valid state
		backups, err = backupClient.List(metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&labels)})
		require.NoError(t, err)
		require.Len(t, backups.Items, size + 1)

		for _, b := range backups.Items {
			require.Equal(t, backupApi.ArangoBackupStateReady, b.Status.State, b.Status.Message)
		}
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
		require.NoError(t, err, "failed to create database: %s", err)

		colname := "backup-test-col"
		col, err := db.CreateCollection(ctx, colname, nil)
		require.NoError(t, err, "failed to create collection: %s", err)

		meta1, err := col.CreateDocument(ctx, &Book{Title: "My first Go-Program", Author: "Adam"})
		require.NoError(t, err, "failed to create document: %s", err)

		// Now create a backup
		_, name, _ := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, nil)
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// insert yet another document
		meta2, err := col.CreateDocument(ctx, &Book{Title: "Bad book title", Author: "Lars"})
		require.NoError(t, err, "failed to create document: %s", err)

		// now restore the backup
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = util.NewString(string(name))
		})
		require.NoError(t, err, "Failed to update deployment: %s", err)

		_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore != nil {
				result := status.Restore

				if result.RequestedFrom != name {
					return fmt.Errorf("Wrong backup in RequestedFrom: %s, expected %s", result.RequestedFrom, name)
				}

				if result.State == api.DeploymentRestoreStateRestoreFailed {
					t.Fatalf("Failed to restore backup: %s", result.Message)
				}

				if result.State == api.DeploymentRestoreStateRestored {
					return nil
				}

				return fmt.Errorf("Not yet restored - staate %s", result.State)
			}

			return fmt.Errorf("Restore is not set on deployment")
		})
		require.NoError(t, err, "Deployment did not restore in time: %s", err)

		// restore was completed, check if documents are there
		found, err := col.DocumentExists(ctx, meta1.Key)
		require.NoError(t, err, "Failed to check if document exists: %s", err)
		require.True(t, found)

		// second document should not exist
		found, err = col.DocumentExists(ctx, meta2.Key)
		require.NoError(t, err, "Failed to check if document exists: %s", err)
		require.False(t, found)

		// delete the RestoreFrom entry
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = nil
		})
		require.NoError(t, err, "Failed to update deployment: %s", err)

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
		require.NoError(t, err, "Failed to update deployment: %s", err)

		depl, err := waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore != nil {
				result := status.Restore

				if result.RequestedFrom != name {
					return fmt.Errorf("Wrong backup in RequestedFrom: %s, expected %s", result.RequestedFrom, name)
				}

				if result.State == api.DeploymentRestoreStateRestored {
					t.Fatalf("Restore backup - not expected: %s", result.Message)
				}

				if result.State == api.DeploymentRestoreStateRestoreFailed {
					return nil
				}

				return fmt.Errorf("Not yet restored - staate %s", result.State)
			}

			return fmt.Errorf("Restore is not set on deployment")
		})
		require.NoError(t, err, "Deployment did not restore in time: %s", err)
		require.NotNil(t, depl.Status.Restore)
		require.Equal(t, api.DeploymentRestoreStateRestoreFailed, depl.Status.Restore.State)
	})

	t.Run("upload", func(t *testing.T) {
		skipOrRemotePath(t)

		// create backup with upload operation
		backup, name, _ := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Upload: newOperation()})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup will be uploaded
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

		require.NotNil(t, backup.Status.Backup)
		require.NotNil(t, backup.Status.Backup.Uploaded)
		require.Nil(t, backup.Status.Backup.Downloaded)

		require.True(t, *backup.Status.Backup.Uploaded)
	})

	t.Run("re-upload", func(t *testing.T) {
		skipOrRemotePath(t)

		// create backup with upload operation
		backup, name, _ := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Upload: newOperation()})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup will be uploaded
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

		require.NotNil(t, backup.Status.Backup)
		require.NotNil(t, backup.Status.Backup.Uploaded)
		require.Nil(t, backup.Status.Backup.Downloaded)

		require.True(t, *backup.Status.Backup.Uploaded)

		// Remove upload option
		currentBackup, err := backupClient.Get(backup.Name, metav1.GetOptions{})
		require.NoError(t, err)

		currentBackup.Spec.Upload = nil

		_, err = backupClient.Update(currentBackup)
		require.NoError(t, err)

		// Wait for uploaded flag to disappear
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

		// Append again upload flag
		currentBackup, err = backupClient.Get(backup.Name, metav1.GetOptions{})
		require.NoError(t, err)

		currentBackup.Spec.Upload = newOperation()

		_, err = backupClient.Update(currentBackup)
		require.NoError(t, err)

		// Wait for uploaded flag to appear
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

	})

	t.Run("upload-download-cycle", func(t *testing.T) {
		skipOrRemotePath(t)

		// create backup with upload operation
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Upload: newOperation()})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup will be uploaded
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

		// check that the backup is actually available
		found, meta, err := statBackupMeta(databaseClient, id)
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		require.Equal(t, meta.Version, backup.Status.Backup.Version)

		require.NotNil(t, backup.Status.Backup)
		require.NotNil(t, backup.Status.Backup.Uploaded)
		require.Nil(t, backup.Status.Backup.Downloaded)

		require.True(t, *backup.Status.Backup.Uploaded)

		// After all remove backup
		backupClient.Delete(name, &metav1.DeleteOptions{})
		_, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotFound)
		require.NoError(t, err, "Backup test failed: %s", err)

		// check that the actual backup has been deleted
		found, _, err = statBackupMeta(databaseClient, id)
		require.False(t, found)

		// create backup with download operation
		backup, name, _ = ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Download: newDownload(string(id))})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup becomes ready
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsReady)
		require.NoError(t, err, "backup did not become ready: %s", err)

		// check that the backup is actually available
		found, meta, err = statBackupMeta(databaseClient, id)
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		require.Equal(t, meta.Version, backup.Status.Backup.Version)

		require.NotNil(t, backup.Status.Backup)
		require.Nil(t, backup.Status.Backup.Uploaded)
		require.NotNil(t, backup.Status.Backup.Downloaded)

		require.True(t, *backup.Status.Backup.Downloaded)
	})

	t.Run("upload-download-upload-cycle", func(t *testing.T) {
		skipOrRemotePath(t)

		// create backup with upload operation
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Upload: newOperation()})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup will be uploaded
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

		// check that the backup is actually available
		found, meta, err := statBackupMeta(databaseClient, id)
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		require.Equal(t, meta.Version, backup.Status.Backup.Version)

		require.NotNil(t, backup.Status.Backup)
		require.NotNil(t, backup.Status.Backup.Uploaded)
		require.Nil(t, backup.Status.Backup.Downloaded)

		require.True(t, *backup.Status.Backup.Uploaded)

		// After all remove backup
		backupClient.Delete(name, &metav1.DeleteOptions{})
		_, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotFound)
		require.NoError(t, err, "Backup test failed: %s", err)

		// check that the actual backup has been deleted
		found, _, err = statBackupMeta(databaseClient, id)
		require.False(t, found)

		// create backup with download operation
		backup, name, _ = ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Download: newDownload(string(id))})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup becomes ready
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsReady)
		require.NoError(t, err, "backup did not become ready: %s", err)

		// check that the backup is actually available
		found, meta, err = statBackupMeta(databaseClient, id)
		require.NoError(t, err, "Backup test failed: %s", err)
		require.True(t, found)
		require.Equal(t, meta.Version, backup.Status.Backup.Version)

		require.NotNil(t, backup.Status.Backup)
		require.Nil(t, backup.Status.Backup.Uploaded)
		require.NotNil(t, backup.Status.Backup.Downloaded)

		require.True(t, *backup.Status.Backup.Downloaded)

		// Add again upload flag
		currentBackup, err := backupClient.Get(backup.Name, metav1.GetOptions{})
		require.NoError(t, err)

		currentBackup.Spec.Upload = newOperation()

		_, err = backupClient.Update(currentBackup)
		require.NoError(t, err)

		// Wait for uploaded flag to appear
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)
	})

	t.Run("create-upload-download-restore-cycle", func(t *testing.T) {
		skipOrRemotePath(t)

		type Book struct {
			Title  string
			Author string
		}

		ctx := context.Background()
		// first add collections, insert data into the cluster
		dbname := "backup-test-db-up-down"
		db, err := databaseClient.CreateDatabase(ctx, dbname, nil)
		require.NoError(t, err, "failed to create database: %s", err)

		colname := "backup-test-col"
		col, err := db.CreateCollection(ctx, colname, nil)
		require.NoError(t, err, "failed to create collection: %s", err)

		meta1, err := col.CreateDocument(ctx, &Book{Title: "My first Go-Program", Author: "Adam"})
		require.NoError(t, err, "failed to create document: %s", err)

		// Now create a backup
		backup, name, id := ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Upload: newOperation()})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup becomes ready
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsUploaded)
		require.NoError(t, err, "backup did not become ready: %s", err)

		// insert yet another document
		meta2, err := col.CreateDocument(ctx, &Book{Title: "Bad book title", Author: "Lars"})
		require.NoError(t, err, "failed to create document: %s", err)

		// now remove the backup locally
		err = databaseClient.Backup().Delete(nil, id)
		require.NoError(t, err, "Failed to delete backup: %s", err)

		// wait for the backup to become unavailable
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotAvailable, 30*time.Second)
		require.NoError(t, err, "Backup test failed: %s", err)
		require.Equal(t, backupApi.ArangoBackupStateDeleted, backup.Status.State)

		// now remove the backup
		backupClient.Delete(name, &metav1.DeleteOptions{})
		_, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsNotFound)
		require.NoError(t, err, "Backup test failed: %s", err)

		// create backup with download operation
		backup, name, _ = ensureBackup(t, depl.GetName(), ns, deploymentClient, backupIsAvailable, &EnsureBackupOptions{Download: newDownload(string(id))})
		defer backupClient.Delete(name, &metav1.DeleteOptions{})

		// wait until the backup becomes ready
		backup, err = waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsReady)
		require.NoError(t, err, "backup did not become ready: %s", err)

		// now restore the backup
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = util.NewString(string(name))
		})
		require.NoError(t, err, "Failed to update deployment: %s", err)

		_, err = waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore != nil {
				result := status.Restore

				if result.RequestedFrom != name {
					return fmt.Errorf("Wrong backup in RequestedFrom: %s, expected %s", result.RequestedFrom, name)
				}

				if result.State == api.DeploymentRestoreStateRestoreFailed {
					t.Fatalf("Failed to restore backup: %s", result.Message)
				}

				if result.State == api.DeploymentRestoreStateRestored {
					return nil
				}

				return fmt.Errorf("Not yet restored - staate %s", result.State)
			}

			return fmt.Errorf("Restore is not set on deployment")
		})
		require.NoError(t, err, "Deployment did not restore in time: %s", err)

		// restore was completed, check if documents are there
		found, err := col.DocumentExists(ctx, meta1.Key)
		require.NoError(t, err, "Failed to check if document exists: %s", err)
		require.True(t, found)

		// second document should not exist
		found, err = col.DocumentExists(ctx, meta2.Key)
		require.NoError(t, err, "Failed to check if document exists: %s", err)
		require.False(t, found)

		// delete the RestoreFrom entry
		_, err = updateDeployment(deploymentClient, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
			spec.RestoreFrom = nil
		})
		require.NoError(t, err, "Failed to update deployment: %s", err)

		// wait for it to be deleted in the status
		waitUntilDeployment(deploymentClient, depl.GetName(), ns, func(depl *api.ArangoDeployment) error {
			status := depl.Status
			if status.Restore == nil {
				return nil
			}

			return fmt.Errorf("Restore is not set to nil")
		})
	})

	t.Run("create-backup-policy", func(t *testing.T) {
		skipOrRemotePath(t)

		selector := metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: deplLabels,
		})

		policy := newBackupPolicy(depl.GetName(), "*/1 * * * *", deplLabels, nil)
		list, err := backupClient.List(metav1.ListOptions{LabelSelector: selector})
		require.NoError(t, err)
		require.Len(t, list.Items, 0, "unexpected matching ArangoBackup objects")

		_, err = backupPolicyClient.Create(policy)
		require.NoError(t, err)
		defer backupPolicyClient.Delete(policy.Name, &metav1.DeleteOptions{})

		// Wait until 2 backups are created
		err = timeout(5*time.Second, 5*time.Minute, func() error {
			list, err := backupClient.List(metav1.ListOptions{LabelSelector: selector})

			if err != nil {
				return err
			}

			t.Logf("Received %d ArangoBackups from label selector %s", len(list.Items), selector)

			if len(list.Items) < 2 {
				return nil
			}

			return interrupt{}
		})
		require.NoError(t, err)

		// Cleanup scheduler
		backupPolicyClient.Delete(policy.Name, &metav1.DeleteOptions{})

		backups, err := backupClient.List(metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: deplLabels,
		})})
		require.NoError(t, err)

		for _, backup := range backups.Items {
			t.Run(fmt.Sprintf("deleting - %s", backup.Name), func(t *testing.T) {
				defer backupClient.Delete(backup.Name, &metav1.DeleteOptions{})

				currentBackup, err := waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsAvailable)
				require.NoError(t, err, "backup did not become available: %s", err)
				backupID := currentBackup.Status.Backup.ID

				// check that the backup is actually available
				found, meta, err := statBackupMeta(databaseClient, driver.BackupID(backupID))
				require.NoError(t, err, "Backup test failed: %s", err)
				require.True(t, found)
				require.Equal(t, meta.Version, currentBackup.Status.Backup.Version)
				require.Equal(t, depl.GetName(), currentBackup.Spec.Deployment.Name)
			})
		}

		// Cleanup
		err = timeout(time.Second, 2*time.Minute, func() error {
			list, err := backupClient.List(metav1.ListOptions{LabelSelector: selector})
			if err != nil {
				return err
			}

			if len(list.Items) != 0 {
				return nil
			}

			return interrupt{}
		})
		require.NoError(t, err)
	})

	t.Run("create-backup-policy-multiple", func(t *testing.T) {
		skipOrRemotePath(t)

		labels := map[string]string{
			"COMMON": "1",
		}
		selector := metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: labels,
		})

		policy := newBackupPolicy(depl.GetName(), "*/1 * * * *", labels, nil)
		list, err := backupClient.List(metav1.ListOptions{LabelSelector: selector})
		require.NoError(t, err)
		require.Len(t, list.Items, 0, "unexpected matching ArangoBackup objects")

		_, err = backupPolicyClient.Create(policy)
		require.NoError(t, err)
		defer backupPolicyClient.Delete(policy.Name, &metav1.DeleteOptions{})

		// Wait until 2 backups are created
		err = timeout(5*time.Second, 5*time.Minute, func() error {
			list, err := backupClient.List(metav1.ListOptions{LabelSelector: selector})

			if err != nil {
				return err
			}

			t.Logf("Received %d ArangoBackups from label selector %s", len(list.Items), selector)

			if len(list.Items) < 4 {
				return nil
			}

			return interrupt{}
		})
		require.NoError(t, err)

		// Cleanup scheduler
		backupPolicyClient.Delete(policy.Name, &metav1.DeleteOptions{})

		for _, deployment := range deployments {
			t.Run(fmt.Sprintf("deployment %s", deployment.Name), func(t *testing.T) {
				backups, err := backupClient.List(metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
					MatchLabels: deployment.Labels,
				})})
				require.NoError(t, err)

				require.Len(t, backups.Items, 2)

				for _, backup := range backups.Items {
					t.Run(fmt.Sprintf("deleting - %s", backup.Name), func(t *testing.T) {
						defer backupClient.Delete(backup.Name, &metav1.DeleteOptions{})

						currentBackup, err := waitUntilBackup(deploymentClient, backup.GetName(), ns, backupIsAvailable)
						require.NoError(t, err, "backup did not become available: %s", err)
						backupID := currentBackup.Status.Backup.ID

						// check that the backup is actually available
						found, meta, err := statBackupMeta(databaseClients[deployment], driver.BackupID(backupID))
						require.NoError(t, err, "Backup test failed: %s", err)
						require.True(t, found)
						require.Equal(t, meta.Version, currentBackup.Status.Backup.Version)
						require.Equal(t, deployment.GetName(), currentBackup.Spec.Deployment.Name)
					})
				}
			})
		}

		// Cleanup
		err = timeout(time.Second, 2*time.Minute, func() error {
			list, err := backupClient.List(metav1.ListOptions{LabelSelector: selector})
			if err != nil {
				return err
			}

			if len(list.Items) != 0 {
				return nil
			}

			return interrupt{}
		})
		require.NoError(t, err)
	})
}
