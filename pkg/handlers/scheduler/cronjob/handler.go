//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package cronjob

import (
	"context"
	"fmt"

	batch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

var logger = logging.Global().RegisterAndGetLogger("scheduler-cronjob-operator", logging.Info)

type handler struct {
	client     arangoClientSet.Interface
	kubeClient kubernetes.Interface

	eventRecorder event.RecorderInstance

	operator operator.Operator
}

func (h *handler) Name() string {
	return Kind()
}

func (h *handler) Handle(ctx context.Context, item operation.Item) error {
	// Get Backup object. It also covers NotFound case

	object, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.client.SchedulerV1beta1().ArangoSchedulerCronJobs(item.Namespace).Get, item.Name, meta.GetOptions{})
	if err != nil {
		if apiErrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	status := object.Status.DeepCopy()

	changed, reconcileErr := operator.HandleP3WithStop(ctx, item, object, status, h.handle)
	if reconcileErr != nil && !operator.IsReconcile(reconcileErr) {
		logger.Err(reconcileErr).Warn("Fail for %s %s/%s",
			item.Kind,
			item.Namespace,
			item.Name)

		return reconcileErr
	}

	if !changed {
		return reconcileErr
	}

	logger.Debug("Updating %s %s/%s",
		item.Kind,
		item.Namespace,
		item.Name)

	if _, err := operator.WithSchedulerCronJobUpdateStatusInterfaceRetry(context.Background(), h.client.SchedulerV1beta1().ArangoSchedulerCronJobs(object.GetNamespace()), object, *status, meta.UpdateOptions{}); err != nil {
		return err
	}

	return reconcileErr
}

func (h *handler) handle(ctx context.Context, item operation.Item, extension *schedulerApi.ArangoSchedulerCronJob, status *schedulerApi.ArangoSchedulerCronJobStatus) (bool, error) {
	return operator.HandleP3(ctx, item, extension, status, h.HandleObject)
}

func (h *handler) HandleObject(ctx context.Context, item operation.Item, extension *schedulerApi.ArangoSchedulerCronJob, status *schedulerApi.ArangoSchedulerCronJobStatus) (bool, error) {
	calculatedProfiles, profilesChecksum, err := scheduler.Profiles(ctx, h.client.SchedulerV1beta1().ArangoProfiles(extension.GetNamespace()), extension.GetLabels(), extension.Spec.Profiles...)
	if err != nil {
		return false, err
	}

	// Try to fetch status
	if profileNames := util.FormatList(calculatedProfiles, func(a util.KV[string, schedulerApi.ProfileAcceptedTemplate]) string {
		return a.K
	}); !equality.Semantic.DeepEqual(status.Profiles, profileNames) {
		status.Profiles = profileNames
		return true, operator.Reconcile("Status Changed")
	}

	var cronJobTemplate batch.CronJob
	cronJobTemplate.ObjectMeta = meta.ObjectMeta{
		Name:        extension.ObjectMeta.Name,
		Labels:      extension.ObjectMeta.Labels,
		Annotations: extension.ObjectMeta.Annotations,
	}

	extension.Spec.CronJobSpec.DeepCopyInto(&cronJobTemplate.Spec)

	deploymentSpecHash, err := util.SHA256FromJSON(cronJobTemplate)
	if err != nil {
		return false, err
	}

	hash := util.SHA256FromString(fmt.Sprintf("%s|%s", profilesChecksum, deploymentSpecHash))

	if err := schedulerApi.ProfileTemplates(util.FormatList(calculatedProfiles, func(a util.KV[string, schedulerApi.ProfileAcceptedTemplate]) *schedulerApi.ProfileTemplate {
		return a.V.Template
	})).RenderOnTemplate(&cronJobTemplate.Spec.JobTemplate.Spec.Template); err != nil {
		return false, err
	}

	if status.Object == nil {
		// Create

		obj := &batch.CronJob{}
		obj.ObjectMeta = meta.ObjectMeta{
			Name:        extension.ObjectMeta.Name,
			Labels:      extension.ObjectMeta.Labels,
			Annotations: extension.ObjectMeta.Annotations,
		}
		cronJobTemplate.Spec.DeepCopyInto(&obj.Spec)

		obj.OwnerReferences = append(obj.OwnerReferences, extension.AsOwner())

		newObj, err := h.kubeClient.BatchV1().CronJobs(extension.GetNamespace()).Create(ctx, obj, meta.CreateOptions{})
		if err != nil {
			h.eventRecorder.Warning(extension, "Create Failed", "Unable to create CronJob: %s", err.Error())
			return false, err
		}

		h.eventRecorder.Normal(extension, "Created", "CronJob %s created", newObj.GetName())

		status.Object = util.NewType(sharedApi.NewObjectWithChecksum(newObj, hash))
		return true, operator.Reconcile("Job Reference Changed")
	}

	// Find existing
	obj, err := h.kubeClient.BatchV1().CronJobs(status.Object.GetNamespace(extension)).Get(ctx, status.Object.GetName(), meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Object removed
			h.eventRecorder.Warning(extension, "Removed", "CronJob %s is gone", status.Object.GetName())
			status.Object = nil
			return true, operator.Reconcile("CronJob Reference Removed")
		}
		return false, err
	}

	// Try to fetch status
	if !equality.Semantic.DeepEqual(status.CronJobStatus, obj.Status) {
		obj.Status.DeepCopyInto(&status.CronJobStatus)
		return true, operator.Reconcile("Status Changed")
	}

	if obj.GetDeletionTimestamp() != nil {
		// Object is deleting, check later
		return false, operator.Reconcile("Job Deleting")
	}

	if !status.Object.Equals(obj) {
		// Object changed or was recreated
		h.eventRecorder.Warning(extension, "Removed", "CronJob %s reference is invalid", status.Object.GetName())
		if err := h.kubeClient.BatchV1().CronJobs(status.Object.GetNamespace(extension)).Delete(ctx, status.Object.GetName(), meta.DeleteOptions{}); err != nil {
			return false, err
		}

		return false, operator.Reconcile("CronJob Deleted")
	}

	// Object is equal, lets check if changed
	if hash != status.Object.GetChecksum() {
		// Checksum changed, lets apply changes
		_, _, err := patcher.Patcher[*batch.CronJob](ctx, h.kubeClient.BatchV1().CronJobs(status.Object.GetNamespace(extension)), obj, meta.PatchOptions{}, func(in *batch.CronJob) []patch.Item {
			return []patch.Item{
				patch.ItemReplace(patch.NewPath("spec"), cronJobTemplate.Spec),
			}
		}, patcher.PatchMetadata(obj))
		if err != nil {
			h.eventRecorder.Warning(extension, "Patch Failed", "Unable to patch CronJob: %s", err.Error())
			return false, err
		}
		h.eventRecorder.Normal(extension, "Updated", "CronJob %s patched", obj.GetName())
		status.Object.Checksum = util.NewType(hash)
		return true, nil
	}

	return false, nil
}

func (h *handler) CanBeHandled(item operation.Item) bool {
	return item.Group == Group() &&
		item.Version == Version() &&
		item.Kind == Kind()
}

func (h *handler) init() {}
