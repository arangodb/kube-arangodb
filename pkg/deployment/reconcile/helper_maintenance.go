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
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
)

func withMaintenance(plan ...api.Action) api.Plan {
	if !features.Maintenance().Enabled() {
		return plan
	}

	var newPlan api.Plan

	newPlan = append(newPlan, api.NewAction(api.ActionTypeEnableMaintenance, api.ServerGroupUnknown, "", "Enable maintenance before actions"))

	newPlan = append(newPlan, plan...)

	newPlan = append(newPlan, api.NewAction(api.ActionTypeDisableMaintenance, api.ServerGroupUnknown, "", "Disable maintenance after actions"))

	return newPlan
}
