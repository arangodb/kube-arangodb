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
// Author Ewout Prangsma
//

package deployment

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
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

	nextInterval := lastInterval
	hasError := false
	ctx := context.Background()
	deploymentName := d.apiObject.GetName()
	defer metrics.SetDuration(inspectDeploymentDurationGauges.WithLabelValues(deploymentName), start)

	// Check deployment still exists
	updated, err := d.deps.DatabaseCRCli.DatabaseV1alpha().ArangoDeployments(d.apiObject.GetNamespace()).Get(deploymentName, metav1.GetOptions{})
	if k8sutil.IsNotFound(err) {
		// Deployment is gone
		log.Info().Msg("Deployment is gone")
		d.Delete()
		return nextInterval
	} else if updated != nil && updated.GetDeletionTimestamp() != nil {
		// Deployment is marked for deletion
		if err := d.runDeploymentFinalizers(ctx); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("ArangoDeployment finalizer inspection failed", err, d.apiObject))
		}
	} else {
		// Is the deployment in failed state, if so, give up.
		if d.GetPhase() == api.DeploymentPhaseFailed {
			log.Debug().Msg("Deployment is in Failed state.")
			return nextInterval
		}

		// Inspect secret hashes
		if err := d.resources.ValidateSecretHashes(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Secret hash validation failed", err, d.apiObject))
		}

		// Check for LicenseKeySecret
		if err := d.resources.ValidateLicenseKeySecret(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("License Key Secret invalid", err, d.apiObject))
		}

		// Is the deployment in a good state?
		status, _ := d.GetStatus()
		if status.Conditions.IsTrue(api.ConditionTypeSecretsChanged) {
			log.Debug().Msg("Condition SecretsChanged is true. Revert secrets before we can continue")
			return nextInterval
		}

		// Ensure we have image info
		if retrySoon, err := d.ensureImages(d.apiObject); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Image detection failed", err, d.apiObject))
		} else if retrySoon {
			nextInterval = minInspectionInterval
		}

		// Inspection of generated resources needed
		if x, err := d.resources.InspectPods(ctx); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Pod inspection failed", err, d.apiObject))
		} else {
			nextInterval = nextInterval.ReduceTo(x)
		}
		if x, err := d.resources.InspectPVCs(ctx); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("PVC inspection failed", err, d.apiObject))
		} else {
			nextInterval = nextInterval.ReduceTo(x)
		}

		// Check members for resilience
		if err := d.resilience.CheckMemberFailure(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Member failure detection failed", err, d.apiObject))
		}

		// Immediate actions
		if err := d.reconciler.CheckDeployment(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Reconciler immediate actions failed", err, d.apiObject))
		}

		// Create scale/update plan
		if err := d.reconciler.CreatePlan(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Plan creation failed", err, d.apiObject))
		}

		// Execute current step of scale/update plan
		retrySoon, err := d.reconciler.ExecutePlan(ctx)
		if err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Plan execution failed", err, d.apiObject))
		}
		if retrySoon {
			nextInterval = minInspectionInterval
		}

		// Ensure all resources are created
		if err := d.resources.EnsureSecrets(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Secret creation failed", err, d.apiObject))
		}
		if err := d.resources.EnsureServices(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Service creation failed", err, d.apiObject))
		}
		if d.haveServiceMonitorCRD {
			if err := d.resources.EnsureServiceMonitor(); err != nil {
				hasError = true
				d.CreateEvent(k8sutil.NewErrorEvent("Service monitor creation failed", err, d.apiObject))
			}
		}

		if err := d.resources.EnsurePVCs(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("PVC creation failed", err, d.apiObject))
		}
		if err := d.resources.EnsurePods(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Pod creation failed", err, d.apiObject))
		}
		if err := d.resources.EnsurePDBs(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("PDB creation failed", err, d.apiObject))
		}

		// Create access packages
		if err := d.createAccessPackages(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("AccessPackage creation failed", err, d.apiObject))
		}

		// Ensure deployment bootstrap
		if err := d.EnsureBootstrap(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Bootstrap failed", err, d.apiObject))
		}

		// Inspect deployment for obsolete members
		if err := d.resources.CleanupRemovedMembers(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Removed member cleanup failed", err, d.apiObject))
		}

		// At the end of the inspect, we cleanup terminated pods.
		if x, err := d.resources.CleanupTerminatedPods(); err != nil {
			hasError = true
			d.CreateEvent(k8sutil.NewErrorEvent("Pod cleanup failed", err, d.apiObject))
		} else {
			nextInterval = nextInterval.ReduceTo(x)
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

// triggerInspection ensures that an inspection is run soon.
func (d *Deployment) triggerInspection() {
	d.inspectTrigger.Trigger()
}

// triggerCRDInspection ensures that an inspection is run soon.
func (d *Deployment) triggerCRDInspection() {
	d.inspectCRDTrigger.Trigger()
}
