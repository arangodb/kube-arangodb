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

package reconcile

import (
	"context"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/client"
	"github.com/arangodb/kube-arangodb/pkg/deployment/reconcile/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) updateClusterLicense(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	context PlanBuilderContext) api.Plan {
	if !spec.License.HasSecretName() {
		return nil
	}

	l, err := k8sutil.GetLicenseFromSecret(context.ACS().CurrentClusterCache(), spec.License.GetSecretName())
	if err != nil {
		r.log.Err(err).Error("License secret error")
		return nil
	}

	if !l.V2.IsV2Set() {
		r.log.Str("secret", spec.License.GetSecretName()).Error("V2 License key is not set")
		return nil
	}

	members := status.Members.AsListInGroups(arangod.GroupsWithLicenseV2()...).Filter(func(a api.DeploymentStatusMemberElement) bool {
		i := a.Member.Image
		if i == nil {
			return false
		}

		return i.ArangoDBVersion.CompareTo("3.9.0") >= 0 && i.Enterprise
	})

	if len(members) == 0 {
		// No member found to take this action
		r.log.Trace("No enterprise member in version 3.9.0 or above")
		return nil
	}

	member := members[0]

	ctxChild, cancel := globals.GetGlobals().Timeouts().ArangoD().WithTimeout(ctx)
	defer cancel()

	c, err := context.GetMembersState().GetMemberClient(member.Member.ID)
	if err != nil {
		r.log.Err(err).Error("Unable to get client")
		return nil
	}

	internalClient := client.NewClient(c.Connection(), r.log)

	if ok, err := licenseV2Compare(ctxChild, internalClient, l.V2); err != nil {
		r.log.Err(err).Error("Unable to verify license")
		return nil
	} else if ok {
		if c, _ := status.Conditions.Get(api.ConditionTypeLicenseSet); !c.IsTrue() || c.Hash != l.V2.V2Hash() {
			return api.Plan{shared.UpdateConditionActionV2("License is set", api.ConditionTypeLicenseSet, true, "License UpToDate", "", l.V2.V2Hash())}
		}
		return nil
	}

	return api.Plan{shared.RemoveConditionActionV2("License is not set", api.ConditionTypeLicenseSet), actions.NewAction(api.ActionTypeLicenseSet, member.Group, member.Member, "Setting license")}
}
