//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package reconcile

import (
	"context"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	sharedReconcile "github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) updateClusterLicense(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if l := status.License; l == nil {
		// Cleanup the Condition
		if status.Conditions.IsTrue(api.ConditionTypeLicenseSet) {
			// Cleanup the old condition if we do not expect license
			if !spec.License.HasSecretName() {
				return api.Plan{sharedReconcile.RemoveConditionActionV2("License is not set", api.ConditionTypeLicenseSet)}
			} else {
				// Cleanup the old condition if we do not expect license
				return api.Plan{sharedReconcile.UpdateConditionActionV2("License is not set", api.ConditionTypeLicenseSet, false, "License Pending", "", "")}

			}
		}
	} else {
		// Set the Condition
		if !status.Conditions.IsTrue(api.ConditionTypeLicenseSet) {
			// Cleanup the old condition
			return api.Plan{sharedReconcile.UpdateConditionActionV2("License is set", api.ConditionTypeLicenseSet, true, "License UpToDate", "", l.Hash)}
		}
	}

	if !spec.License.HasSecretName() {
		if status.License != nil {
			return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference")}
		}
		return nil
	}

	mode, err := r.updateClusterLicenseDiscover(spec, context)
	if err != nil {
		r.log.Err(err).Warn("Unable to discover license mode")
	}

	if l := status.License; l != nil {
		if mode != l.Mode {
			return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference - invalid mode")}
		}
	}

	switch mode {
	case api.LicenseModeKey:
		if p := r.updateClusterLicenseKey(ctx, spec, status, context); len(p) > 0 {
			return p
		}
	case api.LicenseModeAPI:
		if p := r.updateClusterLicenseAPI(ctx, spec, status, context); len(p) > 0 {
			return p
		}
	}

	return nil
}

func (r *Reconciler) updateClusterLicenseDiscover(spec api.DeploymentSpec, context PlanBuilderContext) (api.LicenseMode, error) {
	switch spec.License.Mode.Get() {
	case api.LicenseModeKey:
		return api.LicenseModeKey, nil
	case api.LicenseModeAPI:
		return api.LicenseModeAPI, nil
	}

	// Run the discovery
	l, err := k8sutil.GetLicenseFromSecret(context.ACS().CurrentClusterCache(), spec.License.GetSecretName())
	if err != nil {
		return "", err
	}

	if l.V2.IsV2Set() {
		return api.LicenseModeKey, nil
	}

	if l.API != nil {
		return api.LicenseModeAPI, nil
	}

	return "", errors.Errorf("Unable to discover License mode")
}

func (r *Reconciler) updateClusterLicenseMember(spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) (api.DeploymentStatusMemberElement, bool) {
	members := status.Members.AsListInGroups(arangod.GroupsWithLicenseV2()...).Filter(func(a api.DeploymentStatusMemberElement) bool {
		i := a.Member.Image
		if i == nil {
			return false
		}

		return i.ArangoDBVersion.CompareTo("3.9.0") >= 0 && i.Enterprise
	})

	if spec.Mode.Get() == api.DeploymentModeActiveFailover {
		cache := context.ACS().CurrentClusterCache()

		// For AF is different
		members = members.Filter(func(a api.DeploymentStatusMemberElement) bool {
			pod, ok := cache.Pod().V1().GetSimple(a.Member.Pod.GetName())
			if !ok {
				return false
			}

			if _, ok := pod.Labels[k8sutil.LabelKeyArangoLeader]; ok {
				return true
			}

			return false
		})
	}

	if len(members) == 0 {
		return api.DeploymentStatusMemberElement{}, false
	}

	return members[0], true
}

func (r *Reconciler) updateClusterLicenseKey(ctx context.Context, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {
	l, err := k8sutil.GetLicenseFromSecret(context.ACS().CurrentClusterCache(), spec.License.GetSecretName())
	if err != nil {
		r.log.Err(err).Error("License secret error")
		return nil
	}

	if !l.V2.IsV2Set() {
		r.log.Str("secret", spec.License.GetSecretName()).Error("V2 License key is not set")
		return nil
	}

	member, ok := r.updateClusterLicenseMember(spec, status, context)

	if !ok {
		// No member found to take this action
		r.log.Trace("No enterprise member in version 3.9.0 or above")
		return nil
	}

	ctxChild, cancel := globals.GetGlobals().Timeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	c, err := context.GetMembersState().GetMemberClient(member.Member.ID)
	if err != nil {
		r.log.Err(err).Error("Unable to get client")
		return nil
	}

	if status.License == nil {
		// Run the set
		return api.Plan{actions.NewAction(api.ActionTypeLicenseSet, member.Group, member.Member, "Generating license")}
	}

	internalClient := client.NewClient(c.Connection(), r.log)

	license, err := internalClient.GetLicense(ctxChild)
	if err != nil {
		r.log.Err(err).Error("Unable to get client")
		return nil
	}

	if status.License.Hash != license.Hash || status.License.InputHash != l.V2.V2Hash() {
		return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference - Invalid Hash")}
	}

	return nil
}

func (r *Reconciler) updateClusterLicenseAPI(ctx context.Context, spec api.DeploymentSpec, status api.DeploymentStatus, context PlanBuilderContext) api.Plan {
	l, err := k8sutil.GetLicenseFromSecret(context.ACS().CurrentClusterCache(), spec.License.GetSecretName())
	if err != nil {
		r.log.Err(err).Error("License secret error")
		return nil
	}

	if l.API == nil {
		r.log.Str("secret", spec.License.GetSecretName()).Error("V2 License key is not set")
		return nil
	}

	member, ok := r.updateClusterLicenseMember(spec, status, context)

	if !ok {
		// No member found to take this action
		r.log.Trace("No enterprise member in version 3.9.0 or above")
		return nil
	}

	ctxChild, cancel := globals.GetGlobals().Timeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	c, err := context.GetMembersState().GetMemberClient(member.Member.ID)
	if err != nil {
		r.log.Err(err).Error("Unable to get client")
		return nil
	}

	if status.License == nil {
		// Run the generation
		return api.Plan{actions.NewAction(api.ActionTypeLicenseGenerate, member.Group, member.Member, "Generating license")}
	}

	internalClient := client.NewClient(c.Connection(), r.log)

	currentLicense, err := internalClient.GetLicense(ctxChild)
	if err != nil {
		r.log.Err(err).Error("Unable to get current license")
		return nil
	}

	if status.License.InputHash != l.API.Hash() {
		// Invalid hash, cleanup
		return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference - Invalid Input")}
	}

	if currentLicense.Hash != status.License.Hash {
		// Invalid hash, cleanup
		return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference - Invalid Hash")}
	}

	if status.License.Regenerate.Time.Before(time.Now()) {
		return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference - Regeneration Required")}
	}

	cache := r.context.ACS().CurrentClusterCache()

	if s, ok := cache.Secret().V1().GetSimple(pod.GetLicenseRegistryCredentialsSecretName(r.context.GetName())); ok {
		if string(util.Optional(s.Data, utilConstants.ChecksumKey, []byte{})) != l.API.Hash() {
			return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference - Registry Change Required")}
		}
	} else {
		return api.Plan{actions.NewClusterAction(api.ActionTypeLicenseClean, "Removing license reference - Registry Change Required")}
	}

	return nil
}
