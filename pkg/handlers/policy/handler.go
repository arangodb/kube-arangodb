//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package policy

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/robfig/cron"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	deployment "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

const (
	backupCreated       = "ArangoBackupCreated"
	policyError         = "Error"
	rescheduled         = "Rescheduled"
	scheduleSkipped     = "ScheduleSkipped"
	cleanedUpOldBackups = "CleanedUpOldBackups"
)

type handler struct {
	client        arangoClientSet.Interface
	kubeClient    kubernetes.Interface
	eventRecorder event.RecorderInstance

	operator operator.Operator
}

func (*handler) Name() string {
	return backup.ArangoBackupPolicyResourceKind
}

func (h *handler) Handle(_ context.Context, item operation.Item) error {
	// Do not act on delete event, finalizers are used
	if item.Operation == operation.Delete {
		return nil
	}

	// Get Backup object. It also cover NotFound case
	policy, err := h.client.BackupV1().ArangoBackupPolicies(item.Namespace).Get(context.Background(), item.Name, meta.GetOptions{})
	if err != nil {
		return err
	}

	status := h.processBackupPolicy(policy.DeepCopy())
	// Nothing to update, objects are equal
	if reflect.DeepEqual(policy.Status, status) {
		return nil
	}

	policy.Status = status

	// Update status on object
	if _, err = h.client.BackupV1().ArangoBackupPolicies(item.Namespace).UpdateStatus(context.Background(), policy, meta.UpdateOptions{}); err != nil {
		return err
	}

	return nil
}

func (h *handler) processBackupPolicy(policy *backupApi.ArangoBackupPolicy) backupApi.ArangoBackupPolicyStatus {
	if err := policy.Validate(); err != nil {
		h.eventRecorder.Warning(policy, policyError, "Policy Error: %s", err.Error())

		return backupApi.ArangoBackupPolicyStatus{
			Message: fmt.Sprintf("Validation error: %s", err.Error()),
		}
	}

	now := time.Now()

	expr, err := cron.ParseStandard(policy.Spec.Schedule)
	if err != nil {
		h.eventRecorder.Warning(policy, policyError, "Policy Error: %s", err.Error())

		return backupApi.ArangoBackupPolicyStatus{
			Message: fmt.Sprintf("error while parsing expr: %s", err.Error()),
		}
	}

	if policy.Status.Scheduled.IsZero() {
		next := expr.Next(now)

		return backupApi.ArangoBackupPolicyStatus{
			Scheduled: meta.Time{
				Time: next,
			},
		}
	}

	// Check if schedule is required
	if policy.Status.Scheduled.Unix() > now.Unix() {
		// check if we need to update schedule in case that string changed
		// in other case schedule string will be updated after scheduling objects
		next := expr.Next(now)

		if next != policy.Status.Scheduled.Time {
			return backupApi.ArangoBackupPolicyStatus{
				Scheduled: meta.Time{
					Time: next,
				},
			}
		}

		return policy.Status
	}

	// Schedule new deployments
	listOptions := meta.ListOptions{}
	if policy.Spec.DeploymentSelector != nil &&
		(policy.Spec.DeploymentSelector.MatchLabels != nil &&
			len(policy.Spec.DeploymentSelector.MatchLabels) > 0 ||
			policy.Spec.DeploymentSelector.MatchExpressions != nil) {
		listOptions.LabelSelector = meta.FormatLabelSelector(policy.Spec.DeploymentSelector)
	}

	deployments, err := h.client.DatabaseV1().ArangoDeployments(policy.Namespace).List(context.Background(), listOptions)
	if err != nil {
		h.eventRecorder.Warning(policy, policyError, "Policy Error: %s", err.Error())

		return backupApi.ArangoBackupPolicyStatus{
			Scheduled: policy.Status.Scheduled,
			Message:   fmt.Sprintf("deployments listing failed: %s", err.Error()),
		}
	}

	needToListBackups := !policy.Spec.GetAllowConcurrent() || policy.Spec.MaxBackups > 0
	for _, deployment := range deployments.Items {
		depl := deployment.DeepCopy()
		ctx := context.Background()

		if needToListBackups {
			backups, err := h.listAllBackupsForPolicy(ctx, depl, policy.Name)
			if err != nil {
				h.eventRecorder.Warning(policy, policyError, "Policy Error: %s", err.Error())
				return backupApi.ArangoBackupPolicyStatus{
					Scheduled: policy.Status.Scheduled,
					Message:   fmt.Sprintf("backup creation failed: %s", err.Error()),
				}
			}
			if numRemoved, err := h.removeOldHealthyBackups(ctx, policy.Spec.MaxBackups, backups); err != nil {
				h.eventRecorder.Warning(policy, policyError, "Policy Error: %s", err.Error())
				return backupApi.ArangoBackupPolicyStatus{
					Scheduled: policy.Status.Scheduled,
					Message:   fmt.Sprintf("automatic backup cleanup failed: %s", err.Error()),
				}
			} else if numRemoved > 0 {
				eventMsg := fmt.Sprintf("Cleaned up %d old backups due to maxBackups setting %s/%s", numRemoved, deployment.Namespace, deployment.Name)
				h.eventRecorder.Normal(policy, cleanedUpOldBackups, eventMsg)
			}
			if !policy.Spec.GetAllowConcurrent() && h.isPreviousBackupInProgress(backups) {
				eventMsg := fmt.Sprintf("Skipping ArangoBackup creation because earlier backup still running %s/%s", deployment.Namespace, deployment.Name)
				h.eventRecorder.Normal(policy, scheduleSkipped, eventMsg)
				continue
			}
		}

		b := policy.NewBackup(depl)
		if _, err := h.client.BackupV1().ArangoBackups(b.Namespace).Create(ctx, b, meta.CreateOptions{}); err != nil {
			h.eventRecorder.Warning(policy, policyError, "Policy Error: %s", err.Error())

			return backupApi.ArangoBackupPolicyStatus{
				Scheduled: policy.Status.Scheduled,
				Message:   fmt.Sprintf("backup creation failed: %s", err.Error()),
			}
		}

		h.eventRecorder.Normal(policy, backupCreated, "Created ArangoBackup: %s/%s", b.Namespace, b.Name)
	}

	next := expr.Next(time.Now())

	h.eventRecorder.Normal(policy, rescheduled, "Rescheduled for: %s", next.String())

	return backupApi.ArangoBackupPolicyStatus{
		Scheduled: meta.Time{
			Time: next,
		},
	}
}

func (*handler) CanBeHandled(item operation.Item) bool {
	return item.Group == backupApi.SchemeGroupVersion.Group &&
		item.Version == backupApi.SchemeGroupVersion.Version &&
		item.Kind == backup.ArangoBackupPolicyResourceKind
}

func (h *handler) listAllBackupsForPolicy(ctx context.Context, d *deployment.ArangoDeployment, policyName string) (util.List[*backupApi.ArangoBackup], error) {
	var r []*backupApi.ArangoBackup

	if err := k8sutil.APIList[*backupApi.ArangoBackupList](ctx, h.client.BackupV1().ArangoBackups(d.Namespace), meta.ListOptions{
		Limit: globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
	}, func(result *backupApi.ArangoBackupList, err error) error {
		if err != nil {
			return err
		}

		for _, b := range result.Items {
			if b.Spec.PolicyName == nil || *b.Spec.PolicyName != policyName {
				continue
			}
			if b.Spec.Deployment.Name != d.Name {
				continue
			}
			r = append(r, b.DeepCopy())
		}

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "Failed to list ArangoBackups")
	}

	return r, nil
}

func (h *handler) isPreviousBackupInProgress(backups util.List[*backupApi.ArangoBackup]) bool {
	inProgressBackups := backups.Count(func(b *backupApi.ArangoBackup) bool {
		switch b.Status.State {
		case backupApi.ArangoBackupStateFailed:
			return false
		}

		if b.Spec.Download != nil {
			return false
		}

		// Backup is not yet done
		if b.Status.Backup == nil {
			return true
		}
		return false
	})
	return inProgressBackups > 0
}

func (h *handler) removeOldHealthyBackups(ctx context.Context, limit int, backups util.List[*backupApi.ArangoBackup]) (int, error) {
	if limit <= 0 {
		// no limit set
		return 0, nil
	}

	healthyBackups := backups.Filter(func(b *backupApi.ArangoBackup) bool {
		return b.Status.State == backupApi.ArangoBackupStateReady
	}).Sort(func(a *backupApi.ArangoBackup, b *backupApi.ArangoBackup) bool {
		// newest first
		return a.CreationTimestamp.After(b.CreationTimestamp.Time)
	})
	if len(healthyBackups) < limit {
		return 0, nil
	}
	toDelete := healthyBackups[limit-1:]
	numDeleted := 0
	for _, b := range toDelete {
		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			return h.client.BackupV1().ArangoBackups(b.Namespace).Delete(ctx, b.Name, meta.DeleteOptions{})
		})
		if err != nil && !kerrors.IsNotFound(err) {
			return numDeleted, errors.Wrapf(err, "could not trigger deletion of backup %s", b.Name)
		}
		numDeleted++
	}
	return numDeleted, nil
}
