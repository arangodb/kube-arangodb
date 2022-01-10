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

package policy

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"

	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"

	"k8s.io/client-go/kubernetes"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/robfig/cron"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	backupCreated = "ArangoBackupCreated"
	policyError   = "Error"
	rescheduled   = "Rescheduled"
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

func (h *handler) Handle(item operation.Item) error {
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

	for _, deployment := range deployments.Items {
		b := policy.NewBackup(deployment.DeepCopy())

		if _, err := h.client.BackupV1().ArangoBackups(b.Namespace).Create(context.Background(), b, meta.CreateOptions{}); err != nil {
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
