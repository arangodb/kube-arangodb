//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
// Author Tomasz Mielech
//

package deployment

import (
	"context"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"

	operatorErrors "github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeoutReconciliationPerNode = time.Second * 20
)

var (
	inspectDeploymentDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_deployment_duration", "Amount of time taken by a single inspection of a deployment (in sec)", metrics.DeploymentName)
)

// getReconciliationTimeout gets timeout for the reconciliation loop.
// The whole reconciliation loop timeout depends on the number of nodes but not less then one minute.
func (d *Deployment) getReconciliationTimeout() time.Duration {
	if nodes, ok := d.GetCachedStatus().GetNodes(); ok {
		if timeout := timeoutReconciliationPerNode * time.Duration(len(nodes.Nodes())); timeout > time.Minute {
			return timeout
		}
	}

	// The minimum timeout for the reconciliation loop.
	return time.Minute
}

// inspectDeployment inspects the entire deployment, creates
// a plan to update if needed and inspects underlying resources.
// This function should be called when:
// - the deployment has changed
// - any of the underlying resources has changed
// - once in a while
// Returns the delay until this function should be called again.
func (d *Deployment) inspectDeployment(lastInterval util.Interval) util.Interval {
	log := d.deps.Log
	start := time.Now()

	timeout := d.getReconciliationTimeout()

	ctxReconciliation, cancelReconciliation := context.WithTimeout(context.Background(), timeout)
	defer cancelReconciliation()
	defer func() {
		d.deps.Log.Info().Msgf("Inspect loop took %s", time.Since(start))
	}()

	nextInterval := lastInterval
	hasError := false

	deploymentName := d.GetName()
	defer metrics.SetDuration(inspectDeploymentDurationGauges.WithLabelValues(deploymentName), start)

	cachedStatus, err := inspector.NewInspector(context.Background(), d.getKubeCli(), d.getMonitoringV1Cli(), d.getArangoCli(), d.GetNamespace())
	if err != nil {
		log.Error().Err(err).Msg("Unable to get resources")
		return minInspectionInterval // Retry ASAP
	}

	// Check deployment still exists
	var updated *api.ArangoDeployment
	err = k8sutil.RunWithTimeout(ctxReconciliation, func(ctxChild context.Context) error {
		var err error
		updated, err = d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(d.GetNamespace()).Get(ctxChild, deploymentName, metav1.GetOptions{})
		return err
	})
	if k8sutil.IsNotFound(err) {
		// Deployment is gone
		log.Info().Msg("Deployment is gone")
		d.Delete()
		return nextInterval
	} else if updated != nil && updated.GetDeletionTimestamp() != nil {
		// Deployment is marked for deletion
		if err := d.runDeploymentFinalizers(ctxReconciliation, cachedStatus); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("ArangoDeployment finalizer inspection failed", err, d.apiObject))
		}
	} else {
		// Check if maintenance annotation is set
		if updated != nil && updated.Annotations != nil {
			if v, ok := updated.Annotations[deployment.ArangoDeploymentPodMaintenanceAnnotation]; ok && v == "true" {
				// Disable checks if we will enter maintenance mode
				log.Info().Str("deployment", deploymentName).Msg("Deployment in maintenance mode")
				return nextInterval
			}
		}
		// Is the deployment in failed state, if so, give up.
		if d.GetPhase() == api.DeploymentPhaseFailed {
			log.Debug().Msg("Deployment is in Failed state.")
			return nextInterval
		}

		d.apiObject = updated

		inspectNextInterval, err := d.inspectDeploymentWithError(ctxReconciliation, nextInterval, cachedStatus)
		if err != nil {
			if !operatorErrors.IsReconcile(err) {
				nextInterval = inspectNextInterval
				hasError = true

				d.CreateEvent(k8sutil.NewErrorEvent("Reconcilation failed", err, d.apiObject))
			} else {
				nextInterval = minInspectionInterval
			}
		}
	}

	// Update next interval (on errors)
	if hasError {
		if d.recentInspectionErrors == 0 {
			nextInterval = minInspectionInterval
			d.recentInspectionErrors++
		}
	} else {
		d.recentInspectionErrors = 0
	}
	return nextInterval.ReduceTo(maxInspectionInterval)
}

func (d *Deployment) inspectDeploymentWithError(ctx context.Context, lastInterval util.Interval,
	cachedStatus inspectorInterface.Inspector) (nextInterval util.Interval, inspectError error) {
	t := time.Now()

	d.SetCachedStatus(cachedStatus)

	defer func() {
		d.deps.Log.Info().Msgf("Reconciliation loop took %s", time.Since(t))
	}()

	// Ensure that spec and status checksum are same
	spec := d.GetSpec()
	status, _ := d.getStatus()

	nextInterval = lastInterval
	inspectError = nil

	checksum, err := spec.Checksum()
	if err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Calculation of spec failed")
	} else {
		condition, exists := status.Conditions.Get(api.ConditionTypeUpToDate)
		if checksum != status.AppliedVersion && (!exists || condition.IsTrue()) {
			if err = d.updateCondition(ctx, api.ConditionTypeUpToDate, false, "Spec Changed", "Spec Object changed. Waiting until plan will be applied"); err != nil {
				return minInspectionInterval, errors.Wrapf(err, "Unable to update UpToDate condition")
			}

			return minInspectionInterval, nil // Retry ASAP
		}
	}

	// Cleanup terminated pods on the beginning of loop
	if x, err := d.resources.CleanupTerminatedPods(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Pod cleanup failed")
	} else {
		nextInterval = nextInterval.ReduceTo(x)
	}

	if err := d.resources.EnsureArangoMembers(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "ArangoMember creation failed")
	}

	if err := d.resources.EnsureServices(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Service creation failed")
	}

	if err := d.resources.EnsureSecrets(ctx, d.deps.Log, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Secret creation failed")
	}

	// Inspect secret hashes
	if err := d.resources.ValidateSecretHashes(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Secret hash validation failed")
	}

	// Check for LicenseKeySecret
	if err := d.resources.ValidateLicenseKeySecret(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "License Key Secret invalid")
	}

	// Is the deployment in a good state?
	if status.Conditions.IsTrue(api.ConditionTypeSecretsChanged) {
		return minInspectionInterval, errors.Newf("Secrets changed")
	}

	// Ensure we have image info
	if retrySoon, exists, err := d.ensureImages(ctx, d.apiObject, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Image detection failed")
	} else if retrySoon || !exists {
		return minInspectionInterval, nil
	}

	// Inspection of generated resources needed
	if x, err := d.resources.InspectPods(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Pod inspection failed")
	} else {
		nextInterval = nextInterval.ReduceTo(x)
	}

	if x, err := d.resources.InspectPVCs(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "PVC inspection failed")
	} else {
		nextInterval = nextInterval.ReduceTo(x)
	}

	// Check members for resilience
	if err := d.resilience.CheckMemberFailure(ctx); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Member failure detection failed")
	}

	// Immediate actions
	if err := d.reconciler.CheckDeployment(ctx); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Reconciler immediate actions failed")
	}

	if interval, err := d.ensureResources(ctx, nextInterval, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Reconciler resource recreation failed")
	} else {
		nextInterval = interval
	}

	inspectDeploymentAgencyFetches.WithLabelValues(d.GetName()).Inc()
	if offset, err := d.RefreshAgencyCache(ctx); err != nil {
		inspectDeploymentAgencyErrors.WithLabelValues(d.GetName()).Inc()
		d.deps.Log.Err(err).Msgf("Unable to refresh agency")
	} else {
		inspectDeploymentAgencyIndex.WithLabelValues(d.GetName()).Set(float64(offset))
	}

	// Refresh maintenance lock
	d.refreshMaintenanceTTL(ctx)

	// Create scale/update plan
	if _, ok := d.apiObject.Annotations[deployment.ArangoDeploymentPlanCleanAnnotation]; ok {
		if err := d.ApplyPatch(ctx, patch.ItemRemove(patch.NewPath("metadata", "annotations", deployment.ArangoDeploymentPlanCleanAnnotation))); err != nil {
			return minInspectionInterval, errors.Wrapf(err, "Unable to create remove annotation patch")
		}

		if err := d.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
			s.Plan = nil
			return true
		}, true); err != nil {
			return minInspectionInterval, errors.Wrapf(err, "Unable clean plan")
		}
	} else if err, updated := d.reconciler.CreatePlan(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Plan creation failed")
	} else if updated {
		return minInspectionInterval, nil
	}

	if d.apiObject.Status.IsPlanEmpty() && status.AppliedVersion != checksum {
		if err := d.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
			s.AppliedVersion = checksum
			return true
		}); err != nil {
			return minInspectionInterval, errors.Wrapf(err, "Unable to update UpToDate condition")
		}

		return minInspectionInterval, nil
	} else if status.AppliedVersion == checksum {
		isUpToDate, reason := d.isUpToDateStatus()

		if !isUpToDate && status.Conditions.IsTrue(api.ConditionTypeUpToDate) {
			if err = d.updateCondition(ctx, api.ConditionTypeUpToDate, false, reason, "There are pending operations in plan or members are in restart process"); err != nil {
				return minInspectionInterval, errors.Wrapf(err, "Unable to update UpToDate condition")
			}

			return minInspectionInterval, nil
		}

		if isUpToDate && !status.Conditions.IsTrue(api.ConditionTypeUpToDate) {
			if err = d.updateCondition(ctx, api.ConditionTypeUpToDate, true, "Spec is Up To Date", "Spec is Up To Date"); err != nil {
				return minInspectionInterval, errors.Wrapf(err, "Unable to update UpToDate condition")
			}

			return minInspectionInterval, nil
		}
	}

	// Execute current step of scale/update plan
	retrySoon, err := d.reconciler.ExecutePlan(ctx, cachedStatus)
	if err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Plan execution failed")
	}
	if retrySoon {
		nextInterval = minInspectionInterval
	}

	// Create access packages
	if err := d.createAccessPackages(ctx); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "AccessPackage creation failed")
	}

	// Inspect deployment for obsolete members
	if err := d.resources.CleanupRemovedMembers(ctx); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Removed member cleanup failed")
	}

	// At the end of the inspect, we cleanup terminated pods.
	if x, err := d.resources.CleanupTerminatedPods(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Pod cleanup failed")
	} else {
		nextInterval = nextInterval.ReduceTo(x)
	}

	return
}

func (d *Deployment) isUpToDateStatus() (upToDate bool, reason string) {
	if !d.apiObject.Status.IsPlanEmpty() {
		return false, "Plan is not empty"
	}

	upToDate = true

	d.apiObject.Status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, member := range list {
			if member.Conditions.IsTrue(api.ConditionTypeRestart) || member.Conditions.IsTrue(api.ConditionTypePendingRestart) {
				upToDate = false
				reason = "Pending restarts on members"
			}
		}
		return nil
	})

	return
}

func (d *Deployment) refreshMaintenanceTTL(ctx context.Context) {
	if d.apiObject.Spec.Mode.Get() == api.DeploymentModeSingle {
		return
	}

	if !features.Maintenance().Enabled() {
		// Maintenance feature is not enabled
		return
	}

	condition, ok := d.status.last.Conditions.Get(api.ConditionTypeMaintenanceMode)

	if !ok || !condition.IsTrue() {
		return
	}

	// Check GracePeriod
	if condition.LastUpdateTime.Add(d.apiObject.Spec.Timeouts.GetMaintenanceGracePeriod()).Before(time.Now()) {
		if err := d.SetAgencyMaintenanceMode(ctx, true); err != nil {
			return
		}
		if err := d.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
			return s.Conditions.Touch(api.ConditionTypeMaintenanceMode)
		}); err != nil {
			return
		}
		d.deps.Log.Info().Msgf("Refreshed maintenance lock")
	}
}

func (d *Deployment) ensureResources(ctx context.Context, lastInterval util.Interval, cachedStatus inspectorInterface.Inspector) (util.Interval, error) {
	// Ensure all resources are created
	if d.haveServiceMonitorCRD {
		if err := d.resources.EnsureServiceMonitor(ctx); err != nil {
			return minInspectionInterval, errors.Wrapf(err, "Service monitor creation failed")
		}
	}

	if err := d.resources.EnsurePVCs(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "PVC creation failed")
	}

	if err := d.resources.EnsurePods(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Pod creation failed")
	}

	if err := d.resources.EnsurePDBs(ctx); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "PDB creation failed")
	}

	if err := d.resources.EnsureAnnotations(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Annotation update failed")
	}

	if err := d.resources.EnsureLabels(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Labels update failed")
	}

	return lastInterval, nil
}

// triggerInspection ensures that an inspection is run soon.
func (d *Deployment) triggerInspection() {
	d.inspectTrigger.Trigger()
}

// triggerCRDInspection ensures that an inspection is run soon.
func (d *Deployment) triggerCRDInspection() {
	d.inspectCRDTrigger.Trigger()
}

func (d *Deployment) updateCondition(ctx context.Context, conditionType api.ConditionType, status bool, reason, message string) error {
	d.deps.Log.Info().Str("condition", string(conditionType)).Bool("status", status).Str("reason", reason).Str("message", message).Msg("Updated condition")
	if err := d.WithStatusUpdate(ctx, func(s *api.DeploymentStatus) bool {
		return s.Conditions.Update(conditionType, status, reason, message)
	}); err != nil {
		return errors.Wrapf(err, "Unable to update condition")
	}

	return nil
}
