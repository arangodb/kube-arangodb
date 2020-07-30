//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
//

package deployment

import (
	"context"
	"time"

	operatorErrors "github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"

	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	inspectDeploymentDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_deployment_duration", "Amount of time taken by a single inspection of a deployment (in sec)", metrics.DeploymentName)
)

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
	defer func() {
		d.deps.Log.Info().Msgf("Inspect loop took %s", time.Since(start))
	}()

	nextInterval := lastInterval
	hasError := false
	ctx := context.Background()
	deploymentName := d.apiObject.GetName()
	defer metrics.SetDuration(inspectDeploymentDurationGauges.WithLabelValues(deploymentName), start)

	cachedStatus, err := inspector.NewInspector(d.GetKubeCli(), d.GetMonitoringV1Cli(), d.GetNamespace())
	if err != nil {
		log.Error().Err(err).Msg("Unable to get resources")
		return minInspectionInterval // Retry ASAP
	}

	// Check deployment still exists
	updated, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(d.apiObject.GetNamespace()).Get(deploymentName, metav1.GetOptions{})
	if k8sutil.IsNotFound(err) {
		// Deployment is gone
		log.Info().Msg("Deployment is gone")
		d.Delete()
		return nextInterval
	} else if updated != nil && updated.GetDeletionTimestamp() != nil {
		// Deployment is marked for deletion
		if err := d.runDeploymentFinalizers(ctx, cachedStatus); err != nil {
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

		if inspectNextInterval, err := d.inspectDeploymentWithError(ctx, nextInterval, cachedStatus); err != nil {
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

func (d *Deployment) inspectDeploymentWithError(ctx context.Context, lastInterval util.Interval, cachedStatus inspector.Inspector) (nextInterval util.Interval, inspectError error) {
	t := time.Now()
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
			if err = d.updateCondition(api.ConditionTypeUpToDate, false, "Spec Changed", "Spec Object changed. Waiting until plan will be applied"); err != nil {
				return minInspectionInterval, errors.Wrapf(err, "Unable to update UpToDate condition")
			}

			return minInspectionInterval, nil // Retry ASAP
		}
	}

	if err := d.resources.EnsureSecrets(d.deps.Log, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Secret creation failed")
	}

	if err := d.resources.EnsureServices(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Service creation failed")
	}

	// Inspect secret hashes
	if err := d.resources.ValidateSecretHashes(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Secret hash validation failed")
	}

	// Check for LicenseKeySecret
	if err := d.resources.ValidateLicenseKeySecret(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "License Key Secret invalid")
	}

	// Is the deployment in a good state?
	if status.Conditions.IsTrue(api.ConditionTypeSecretsChanged) {
		return minInspectionInterval, errors.Errorf("Secrets changed")
	}

	// Ensure we have image info
	if retrySoon, exists, err := d.ensureImages(d.apiObject); err != nil {
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
	if err := d.resilience.CheckMemberFailure(); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Member failure detection failed")
	}

	// Immediate actions
	if err := d.reconciler.CheckDeployment(); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Reconciler immediate actions failed")
	}

	if interval, err := d.ensureResources(nextInterval, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Reconciler resource recreation failed")
	} else {
		nextInterval = interval
	}

	// Create scale/update plan
	if err, updated := d.reconciler.CreatePlan(ctx, cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Plan creation failed")
	} else if updated {
		return minInspectionInterval, nil
	}

	if d.apiObject.Status.Plan.IsEmpty() && status.AppliedVersion != checksum {
		if err := d.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
			s.AppliedVersion = checksum
			return true
		}); err != nil {
			return minInspectionInterval, errors.Wrapf(err, "Unable to update UpToDate condition")
		}

		return minInspectionInterval, nil
	} else if status.AppliedVersion == checksum {
		if !status.Plan.IsEmpty() && status.Conditions.IsTrue(api.ConditionTypeUpToDate) {
			if err = d.updateCondition(api.ConditionTypeUpToDate, false, "Plan is not empty", "There are pending operations in plan"); err != nil {
				return minInspectionInterval, errors.Wrapf(err, "Unable to update UpToDate condition")
			}

			return minInspectionInterval, nil
		}

		if status.Plan.IsEmpty() && !status.Conditions.IsTrue(api.ConditionTypeUpToDate) {
			if err = d.updateCondition(api.ConditionTypeUpToDate, true, "Spec is Up To Date", "Spec is Up To Date"); err != nil {
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
	if err := d.createAccessPackages(); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "AccessPackage creation failed")
	}

	// Ensure deployment bootstrap
	if err := d.EnsureBootstrap(); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Bootstrap failed")
	}

	// Inspect deployment for obsolete members
	if err := d.resources.CleanupRemovedMembers(); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Removed member cleanup failed")
	}

	// At the end of the inspect, we cleanup terminated pods.
	if x, err := d.resources.CleanupTerminatedPods(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Pod cleanup failed")
	} else {
		nextInterval = nextInterval.ReduceTo(x)
	}

	return
}

func (d *Deployment) ensureResources(lastInterval util.Interval, cachedStatus inspector.Inspector) (util.Interval, error) {
	// Ensure all resources are created
	if d.haveServiceMonitorCRD {
		if err := d.resources.EnsureServiceMonitor(); err != nil {
			return minInspectionInterval, errors.Wrapf(err, "Service monitor creation failed")
		}
	}

	if err := d.resources.EnsurePVCs(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "PVC creation failed")
	}

	if err := d.resources.EnsurePods(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Pod creation failed")
	}

	if err := d.resources.EnsurePDBs(); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "PDB creation failed")
	}

	if err := d.resources.EnsureAnnotations(cachedStatus); err != nil {
		return minInspectionInterval, errors.Wrapf(err, "Annotation update failed")
	}

	if err := d.resources.EnsureLabels(cachedStatus); err != nil {
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

func (d *Deployment) updateCondition(conditionType api.ConditionType, status bool, reason, message string) error {
	d.deps.Log.Info().Str("condition", string(conditionType)).Bool("status", status).Str("reason", reason).Str("message", message).Msg("Updated condition")
	if err := d.WithStatusUpdate(func(s *api.DeploymentStatus) bool {
		return s.Conditions.Update(conditionType, status, reason, message)
	}); err != nil {
		return errors.Wrapf(err, "Unable to update condition")
	}

	return nil
}
