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
	"crypto/x509"
	"encoding/pem"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// upgradeDecision is the result of an upgrade check.
type upgradeDecision struct {
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
		log.Debug().Err(err).Msg("Failed to get owned pods")
		return maskAny(err)
	}

	// Create plan
	apiObject := d.context.GetAPIObject()
	spec := d.context.GetSpec()
	status := d.context.GetStatus()
	newPlan, changed := createPlan(d.log, apiObject, status.Plan, spec, status, pods, d.context.GetTLSKeyfile)

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
	if err := d.context.UpdateStatus(status); err != nil {
		return maskAny(err)
	}
	return nil
}

// createPlan considers the given specification & status and creates a plan to get the status in line with the specification.
// If a plan already exists, the given plan is returned with false.
// Otherwise the new plan is returned with a boolean true.
func createPlan(log zerolog.Logger, apiObject metav1.Object,
	currentPlan api.Plan, spec api.DeploymentSpec,
	status api.DeploymentStatus, pods []v1.Pod,
	getTLSKeyfile func(group api.ServerGroup, member api.MemberStatus) (string, error)) (api.Plan, bool) {
	if len(currentPlan) > 0 {
		// Plan already exists, complete that first
		return currentPlan, false
	}

	// Check for various scenario's
	var plan api.Plan

	// Check for members in failed state
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members *api.MemberStatusList) error {
		for _, m := range *members {
			if m.Phase == api.MemberPhaseFailed && len(plan) == 0 {
				newID := ""
				if group == api.ServerGroupAgents {
					newID = m.ID // Agents cannot (yet) be replaced with new IDs
				}
				plan = append(plan,
					api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
					api.NewAction(api.ActionTypeAddMember, group, newID),
				)
			}
		}
		return nil
	})

	// Check for cleaned out dbserver in created state
	for _, m := range status.Members.DBServers {
		if len(plan) == 0 && m.Phase == api.MemberPhaseCreated && m.Conditions.IsTrue(api.ConditionTypeCleanedOut) {
			plan = append(plan,
				api.NewAction(api.ActionTypeRemoveMember, api.ServerGroupDBServers, m.ID),
				api.NewAction(api.ActionTypeAddMember, api.ServerGroupDBServers, ""),
			)
		}
	}

	// Check for scale up/down
	if len(plan) == 0 {
		switch spec.GetMode() {
		case api.DeploymentModeSingle:
			// Never scale down
		case api.DeploymentModeActiveFailover:
			// Only scale singles
			plan = append(plan, createScalePlan(log, status.Members.Single, api.ServerGroupSingle, spec.Single.GetCount())...)
		case api.DeploymentModeCluster:
			// Scale dbservers, coordinators
			plan = append(plan, createScalePlan(log, status.Members.DBServers, api.ServerGroupDBServers, spec.DBServers.GetCount())...)
			plan = append(plan, createScalePlan(log, status.Members.Coordinators, api.ServerGroupCoordinators, spec.Coordinators.GetCount())...)
		}
		if spec.GetMode().SupportsSync() {
			// Scale syncmasters & syncworkers
			plan = append(plan, createScalePlan(log, status.Members.SyncMasters, api.ServerGroupSyncMasters, spec.SyncMasters.GetCount())...)
			plan = append(plan, createScalePlan(log, status.Members.SyncWorkers, api.ServerGroupSyncWorkers, spec.SyncWorkers.GetCount())...)
		}
	}

	// Check for the need to rotate one or more members
	if len(plan) == 0 {
		getPod := func(podName string) *v1.Pod {
			for _, p := range pods {
				if p.GetName() == podName {
					return &p
				}
			}
			return nil
		}
		status.Members.ForeachServerGroup(func(group api.ServerGroup, members *api.MemberStatusList) error {
			for _, m := range *members {
				if len(plan) > 0 {
					// Only 1 change at a time
					continue
				}
				if m.Phase != api.MemberPhaseCreated {
					// Only rotate when phase is created
					continue
				}
				if podName := m.PodName; podName != "" {
					if p := getPod(podName); p != nil {
						// Got pod, compare it with what it should be
						decision := podNeedsUpgrading(*p, spec, status.Images)
						if decision.UpgradeNeeded && decision.UpgradeAllowed {
							plan = append(plan, createUpgradeMemberPlan(log, m, group, "Version upgrade")...)
						} else {
							rotNeeded, reason := podNeedsRotation(*p, apiObject, spec, group, status.Members.Agents, m.ID)
							if rotNeeded {
								plan = append(plan, createRotateMemberPlan(log, m, group, reason)...)
							}
						}
					}
				}
			}
			return nil
		})
	}

	// Check for the need to rotate TLS certificate of a members
	if len(plan) == 0 && spec.TLS.IsSecure() {
		status.Members.ForeachServerGroup(func(group api.ServerGroup, members *api.MemberStatusList) error {
			for _, m := range *members {
				if len(plan) > 0 {
					// Only 1 change at a time
					continue
				}
				if m.Phase != api.MemberPhaseCreated {
					// Only make changes when phase is created
					continue
				}
				if group == api.ServerGroupSyncWorkers {
					// SyncWorkers have no externally created TLS keyfile
					continue
				}
				// Load keyfile
				keyfile, err := getTLSKeyfile(group, m)
				if err != nil {
					log.Warn().Err(err).
						Str("role", group.AsRole()).
						Str("id", m.ID).
						Msg("Failed to get TLS secret")
					continue
				}
				renewalNeeded := tlsKeyfileNeedsRenewal(log, keyfile)
				if renewalNeeded {
					plan = append(append(plan,
						api.NewAction(api.ActionTypeRenewTLSCertificate, group, m.ID)),
						createRotateMemberPlan(log, m, group, "TLS certificate renewal")...,
					)
				}
			}
			return nil
		})
	}

	// Return plan
	return plan, true
}

// podNeedsUpgrading decides if an upgrade of the pod is needed (to comply with
// the given spec) and if that is allowed.
func podNeedsUpgrading(p v1.Pod, spec api.DeploymentSpec, images api.ImageInfoList) upgradeDecision {
	if c, found := k8sutil.GetContainerByName(&p, k8sutil.ServerContainerName); found {
		specImageInfo, found := images.GetByImage(spec.GetImage())
		if !found {
			return upgradeDecision{UpgradeNeeded: false}
		}
		podImageInfo, found := images.GetByImageID(c.Image)
		if !found {
			return upgradeDecision{UpgradeNeeded: false}
		}
		if specImageInfo.ImageID == podImageInfo.ImageID {
			// No change
			return upgradeDecision{UpgradeNeeded: false}
		}
		// Image changed, check if change is allowed
		specVersion := specImageInfo.ArangoDBVersion
		podVersion := podImageInfo.ArangoDBVersion
		if specVersion.Major() != podVersion.Major() {
			// E.g. 3.x -> 4.x, we cannot allow automatically
			return upgradeDecision{UpgradeNeeded: true, UpgradeAllowed: false}
		}
		if specVersion.Minor() != podVersion.Minor() {
			// Is allowed, with `--database.auto-upgrade`
			return upgradeDecision{
				UpgradeNeeded:     true,
				UpgradeAllowed:    true,
				AutoUpgradeNeeded: true,
			}
		}
		// Patch version change, rotate only
		return upgradeDecision{
			UpgradeNeeded:     true,
			UpgradeAllowed:    true,
			AutoUpgradeNeeded: false,
		}
	}
	return upgradeDecision{UpgradeNeeded: false}
}

// podNeedsRotation returns true when the specification of the
// given pod differs from what it should be according to the
// given deployment spec.
// When true is returned, a reason for the rotation is already returned.
func podNeedsRotation(p v1.Pod, apiObject metav1.Object, spec api.DeploymentSpec,
	group api.ServerGroup, agents api.MemberStatusList, id string) (bool, string) {
	// Check image pull policy
	if c, found := k8sutil.GetContainerByName(&p, k8sutil.ServerContainerName); found {
		if c.ImagePullPolicy != spec.GetImagePullPolicy() {
			return true, "Image pull policy changed"
		}
	} else {
		return true, "Server container not found"
	}
	// Check arguments
	/*expectedArgs := createArangodArgs(apiObject, spec, group, agents, id)
	if len(expectedArgs) != len(c.Args) {
		return true, "Arguments changed"
	}
	for i, a := range expectedArgs {
		if c.Args[i] != a {
			return true, "Arguments changed"
		}
	}*/

	return false, ""
}

// tlsKeyfileNeedsRenewal decides if the certificate in the given keyfile
// should be renewed.
func tlsKeyfileNeedsRenewal(log zerolog.Logger, keyfile string) bool {
	raw := []byte(keyfile)
	for {
		var derBlock *pem.Block
		derBlock, raw = pem.Decode(raw)
		if derBlock == nil {
			break
		}
		if derBlock.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(derBlock.Bytes)
			if err != nil {
				// We do not understand the certificate, let's renew it
				log.Warn().Err(err).Msg("Failed to parse x509 certificate. Renewing it")
				return true
			}
			if cert.IsCA {
				// Only look at the server certificate, not CA or intermediate
				continue
			}
			// Check expiration date. Renewal at 2/3 of lifetime.
			ttl := cert.NotAfter.Sub(cert.NotBefore)
			expirationDate := cert.NotBefore.Add((ttl / 3) * 2)
			if expirationDate.Before(time.Now()) {
				// We should renew now
				log.Debug().
					Str("not-before", cert.NotBefore.String()).
					Str("not-after", cert.NotAfter.String()).
					Str("expiration-date", expirationDate.String()).
					Msg("TLS certificate renewal needed")
				return true
			}
		}
	}
	return false
}

// createScalePlan creates a scaling plan for a single server group
func createScalePlan(log zerolog.Logger, members api.MemberStatusList, group api.ServerGroup, count int) api.Plan {
	var plan api.Plan
	if len(members) < count {
		// Scale up
		toAdd := count - len(members)
		for i := 0; i < toAdd; i++ {
			plan = append(plan, api.NewAction(api.ActionTypeAddMember, group, ""))
		}
		log.Debug().
			Int("delta", toAdd).
			Str("role", group.AsRole()).
			Msg("Creating scale-up plan")
	} else if len(members) > count {
		// Note, we scale down 1 member as a time
		if m, err := members.SelectMemberToRemove(); err == nil {
			if group == api.ServerGroupDBServers {
				plan = append(plan,
					api.NewAction(api.ActionTypeCleanOutMember, group, m.ID),
				)
			}
			plan = append(plan,
				api.NewAction(api.ActionTypeShutdownMember, group, m.ID),
				api.NewAction(api.ActionTypeRemoveMember, group, m.ID),
			)
			log.Debug().
				Str("role", group.AsRole()).
				Msg("Creating scale-down plan")
		}
	}
	return plan
}

// createRotateMemberPlan creates a plan to rotate (stop-recreate-start) an existing
// member.
func createRotateMemberPlan(log zerolog.Logger, member api.MemberStatus,
	group api.ServerGroup, reason string) api.Plan {
	log.Debug().
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Msg("Creating rotation plan")
	plan := api.Plan{
		api.NewAction(api.ActionTypeRotateMember, group, member.ID, reason),
		api.NewAction(api.ActionTypeWaitForMemberUp, group, member.ID),
	}
	return plan
}

// createUpgradeMemberPlan creates a plan to upgrade (stop-recreateWithAutoUpgrade-stop-start) an existing
// member.
func createUpgradeMemberPlan(log zerolog.Logger, member api.MemberStatus,
	group api.ServerGroup, reason string) api.Plan {
	log.Debug().
		Str("id", member.ID).
		Str("role", group.AsRole()).
		Msg("Creating upgrade plan")
	plan := api.Plan{
		api.NewAction(api.ActionTypeUpgradeMember, group, member.ID, reason),
		api.NewAction(api.ActionTypeWaitForMemberUp, group, member.ID),
	}
	return plan
}
