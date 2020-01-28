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

package reconcile

import (
	goContext "context"
	"fmt"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency"
	"time"

	driver "github.com/arangodb/go-driver"
	upgraderules "github.com/arangodb/go-upgrade-rules"
	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
)

// upgradeDecision is the result of an upgrade check.
type upgradeDecision struct {
	FromVersion       driver.Version
	FromLicense       upgraderules.License
	ToVersion         driver.Version
	ToLicense         upgraderules.License
	UpgradeNeeded     bool // If set, the image version has changed
	UpgradeAllowed    bool // If set, it is an allowed version change
	AutoUpgradeNeeded bool // If set, the database must be started with `--database.auto-upgrade` once
}

// CreatePlan considers the current specification & status of the deployment creates a plan to
// get the status in line with the specification.
// If a plan already exists, nothing is done.
func (d *Reconciler) CreatePlan() error {
	// Get all current pods
	pods, err := d.context.GetOwnedPods()
	if err != nil {
		d.log.Debug().Err(err).Msg("Failed to get owned pods")
		return maskAny(err)
	}

	// Create plan
	apiObject := d.context.GetAPIObject()
	spec := d.context.GetSpec()
	status, lastVersion := d.context.GetStatus()
	ctx := newPlanBuilderContext(d.context)
	newPlan, changed := createPlan(d.log, apiObject, status.Plan, spec, status, pods, ctx)

	// If not change, we're done
	if !changed {
		return nil
	}

	// Save plan
	if len(newPlan) == 0 {
		// Nothing to do
		return nil
	}
	status.Plan = newPlan
	if err := d.context.UpdateStatus(status, lastVersion); err != nil {
		return maskAny(err)
	}
	return nil
}

func fetchAgency(log zerolog.Logger,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) (*agency.ArangoPlanDatabases,  error) {
	if spec.GetMode() != api.DeploymentModeCluster && spec.GetMode() != api.DeploymentModeActiveFailover {
		return nil, nil
	} else if status.Members.Agents.MembersReady() > 0 {
		agencyCtx, agencyCancel := goContext.WithTimeout(goContext.Background(), time.Minute)
		defer agencyCancel()

		ret := &agency.ArangoPlanDatabases{}

		if err := context.GetAgencyData(agencyCtx, ret, agency.ArangoKey, agency.PlanKey, agency.PlanCollectionsKey); err != nil {
			return nil, err
		}

		return ret, nil
	} else {
		return nil, fmt.Errorf("not able to read from agency when agency is down")
	}
}

// createPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func createPlan(log zerolog.Logger, apiObject k8sutil.APIObject,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus, pods []v1.Pod,
	context PlanBuilderContext) (api.Plan, bool) {
	if !currentPlan.IsEmpty() {
		// Plan already exists, complete that first
		return currentPlan, false
	}

	// Fetch agency plan
	agencyPlan, agencyErr := fetchAgency(log, spec, status, context)

	// Check for various scenario's
	var plan api.Plan

	// Check for members in failed state
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		for _, m := range members {
			if m.Phase != api.MemberPhaseFailed || len(plan) > 0 {
				continue
			}

			memberLog := log.Info().Str("id", m.ID).Str("role", group.AsRole())

			if group == api.ServerGroupDBServers && spec.GetMode() == api.DeploymentModeCluster {
				// Do pre check for DBServers. If agency is down DBServers should not be touch
				if agencyErr != nil {
					memberLog.Msg("Error in agency")
					continue
				}

				if agencyPlan == nil {
					memberLog.Msg("AgencyPlan is nil")
					continue
				}

				if agencyPlan.IsDBServerInDatabases(m.ID) {
					// DBServer still exists in agency plan! Will not be removed, but needs to be recreated
					memberLog.Msg("Recreating DBServer - it cannot be removed gracefully")
					plan = append(plan,
						api.NewAction(api.ActionTypeRecreateMember, group, m.ID))
					continue
				}

				// Everything is fine, proceed
			}

			memberLog.Msg("Creating member replacement plan because member has failed")
			newID := ""
			if group == api.ServerGroupAgents {
				newID = m.ID // Agents cannot (yet) be replaced with new IDs
			}
			plan = append(plan,
				api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
				api.NewAction(api.ActionTypeAddMember, group, newID),
			)
		}
		return nil
	})

	// Ensure that we were able to get agency info
	if len(plan) == 0 && agencyErr != nil {
		log.Err(agencyErr).Msg("unable to build further plan without access to agency")
		return plan, false
	}

	// Check for cleaned out dbserver in created state
	for _, m := range status.Members.DBServers {
		if plan.IsEmpty() && m.Phase.IsCreatedOrDrain() && m.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			log.Debug().
				Str("id", m.ID).
				Str("role", api.ServerGroupDBServers.AsRole()).
				Msg("Creating dbserver replacement plan because server is cleanout in created phase")
			plan = append(plan,
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, m.ID),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupDBServers, ""),
			)
		}
	}

	// Check for scale up/down
	if plan.IsEmpty() {
		plan = createScaleMemeberPlan(log, spec, status)
	}

	// Check for the need to rotate one or more members
	if plan.IsEmpty() {
		plan = createRotateOrUpgradePlan(log, apiObject, spec, status, context, pods)
	}

	// Check for the need to rotate TLS certificate of a members
	if plan.IsEmpty() {
		plan = createRotateTLSServerCertificatePlan(log, spec, status, context.GetTLSKeyfile)
	}

	// Check for changes storage classes or requirements
	if plan.IsEmpty() {
		plan = createRotateServerStoragePlan(log, apiObject, spec, status, context.GetPvc, context.CreateEvent)
	}

	// Check for the need to rotate TLS CA certificate and all members
	if plan.IsEmpty() {
		plan = createRotateTLSCAPlan(log, apiObject, spec, status, context.GetTLSCA, context.CreateEvent)
	}

	// Return plan
	return plan, true
}

// createRotateMemberPlan creates a plan to rotate (stop-recreate-start) an existing
// member.
func createRotateMemberPlan(log zerolog.Logger, member api.MemberStatus,
	group api.ServerGroup, reason string) api.Plan {
	log.Debug().
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Str("reason", reason).
		Msg("Creating rotation plan")
	plan := api.Plan{
		api.NewAction(api.ActionTypeRotateMember, group, member.ID, reason),
		api.NewAction(api.ActionTypeWaitForMemberUp, group, member.ID),
	}
	return plan
}
