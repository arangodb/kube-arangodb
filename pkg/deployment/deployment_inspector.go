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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// inspectDeployment inspects the entire deployment, creates
// a plan to update if needed and inspects underlying resources.
// This function should be called when:
// - the deployment has changed
// - any of the underlying resources has changed
// - once in a while
// Returns the delay until this function should be called again.
func (d *Deployment) inspectDeployment(lastInterval time.Duration) time.Duration {
	log := d.deps.Log

	nextInterval := lastInterval
	hasError := false
	ctx := context.Background()

	// Check deployment still exists
	if _, err := d.deps.DatabaseCRCli.DatabaseV1alpha().ArangoDeployments(d.apiObject.GetNamespace()).Get(d.apiObject.GetName(), metav1.GetOptions{}); k8sutil.IsNotFound(err) {
		// Deployment is gone
		log.Info().Msg("Deployment is gone")
		d.Delete()
		return nextInterval
	}

	// Is the deployment in failed state, if so, give up.
	if d.status.Phase == api.DeploymentPhaseFailed {
		log.Debug().Msg("Deployment is in Failed state.")
		return nextInterval
	}

	// Inspect secret hashes
	if err := d.resources.ValidateSecretHashes(); err != nil {
		hasError = true
		d.CreateEvent(k8sutil.NewErrorEvent("Secret hash validation failed", err, d.apiObject))
	}

	// Is the deployment in a good state?
	if d.status.Conditions.IsTrue(api.ConditionTypeSecretsChanged) {
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
	if err := d.resources.InspectPods(ctx); err != nil {
		hasError = true
		d.CreateEvent(k8sutil.NewErrorEvent("Pod inspection failed", err, d.apiObject))
	}
	if err := d.resources.InspectPVCs(ctx); err != nil {
		hasError = true
		d.CreateEvent(k8sutil.NewErrorEvent("PVC inspection failed", err, d.apiObject))
	}

	// Check members for resilience
	if err := d.resilience.CheckMemberFailure(); err != nil {
		hasError = true
		d.CreateEvent(k8sutil.NewErrorEvent("Member failure detection failed", err, d.apiObject))
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
	if err := d.resources.EnsurePVCs(); err != nil {
		hasError = true
		d.CreateEvent(k8sutil.NewErrorEvent("PVC creation failed", err, d.apiObject))
	}
	if err := d.resources.EnsurePods(); err != nil {
		hasError = true
		d.CreateEvent(k8sutil.NewErrorEvent("Pod creation failed", err, d.apiObject))
	}

	// At the end of the inspect, we cleanup terminated pods.
	if d.resources.CleanupTerminatedPods(); err != nil {
		hasError = true
		d.CreateEvent(k8sutil.NewErrorEvent("Pod cleanup failed", err, d.apiObject))
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
	if nextInterval > maxInspectionInterval {
		nextInterval = maxInspectionInterval
	}
	return nextInterval
}

// triggerInspection ensures that an inspection is run soon.
func (d *Deployment) triggerInspection() {
	d.inspectTrigger.Trigger()
}
