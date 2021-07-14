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
// Author Adam Janikowski
//

package reconcile

import (
	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

func withMaintenance(spec api.DeploymentSpec, plan ...api.Action) api.Plan {
	if !features.Maintenance().Enabled() {
		return plan
	}

	if spec.Database.GetMaintenance() {
		// If maintenance is enabled skip
		return plan
	}

	return api.AsPlan(plan).Before(api.NewAction(api.ActionTypeEnableMaintenance, api.ServerGroupUnknown, "", "Enable maintenance before actions"))
}

func skipResignLeadership(v driver.Version) bool {
	return (v.CompareTo("3.6.0") >= 0 && v.CompareTo("3.6.14") <= 0) &&
		(v.CompareTo("3.7.0") >= 0 && v.CompareTo("3.7.13") <= 0)
}

func withResignLeadership(group api.ServerGroup, member api.MemberStatus, reason string, plan ...api.Action) api.Plan {
	if member.Image == nil {
		return plan
	}

	if skipResignLeadership(member.Image.ArangoDBVersion) {
		return plan
	}

	return api.AsPlan(plan).After(api.NewAction(api.ActionTypeResignLeadership, group, member.ID, reason))
}
