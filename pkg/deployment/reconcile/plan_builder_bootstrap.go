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
	"context"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/rs/zerolog"
)

func createBootstrapPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspector.Inspector, context PlanBuilderContext) api.Plan {

	if !status.Conditions.IsTrue(api.ConditionTypeReady) {
		return nil
	}

	if condition, hasBootstrap := status.Conditions.Get(api.ConditionTypeBootstrapCompleted); !hasBootstrap || condition.Status == core.ConditionTrue {
		return nil
	}

	for user, secret := range spec.Bootstrap.PasswordSecretNames {
		if secret.IsNone() {
			continue
		}

		if s := status.SecretHashes; s != nil {
			if u := s.Users; u != nil {
				if _, ok := u[user]; ok {
					continue
				}
			}
		}

		return api.Plan{api.NewAction(api.ActionTypeBootstrapSetPassword, api.ServerGroupUnknown, "", "Updating password").AddParam("user", user)}
	}

	return api.Plan{api.NewAction(api.ActionTypeBootstrapUpdate, api.ServerGroupUnknown, "", "Finalizing bootstrap")}
}
