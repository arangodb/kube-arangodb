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

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"

	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/rs/zerolog"
)

func createRotateTLSServerSNIPlan(ctx context.Context,
	log zerolog.Logger, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	cachedStatus inspectorInterface.Inspector, context PlanBuilderContext) api.Plan {
	if !spec.TLS.IsSecure() {
		return nil
	}

	if i := status.CurrentImage; i == nil || !i.Enterprise {
		return nil
	}

	sni := spec.TLS.SNI
	if sni == nil {
		return nil
	}

	fetchedSecrets, err := mapTLSSNIConfig(log, *sni, cachedStatus)
	if err != nil {
		log.Warn().Err(err).Msg("Unable to get SNI desired state")
		return nil
	}

	var plan api.Plan
	status.Members.ForeachServerGroup(func(group api.ServerGroup, members api.MemberStatusList) error {
		if !pod.GroupSNISupported(spec.Mode.Get(), group) {
			return nil
		}

		for _, m := range members {
			if !plan.IsEmpty() {
				// Only 1 member at a time
				return nil
			}

			if m.Phase != api.MemberPhaseCreated {
				// Only make changes when phase is created
				continue
			}

			if i, ok := status.Images.GetByImageID(m.ImageID); !ok || !features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
				continue
			}

			c, err := context.GetServerClient(ctx, group, m.ID)
			if err != nil {
				log.Warn().Err(err).Msg("Unable to get client")
				continue
			}

			if ok, err := compareTLSSNIConfig(ctx, c.Connection(), fetchedSecrets, false); err != nil {
				log.Warn().Err(err).Msg("SNI compare failed")
				return nil

			} else if !ok {
				switch spec.TLS.Mode.Get() {
				case api.TLSRotateModeRecreate:
					plan = append(plan, createRotateMemberPlan(log, m, group, "SNI Secret needs update")...)
				case api.TLSRotateModeInPlace:
					plan = append(plan,
						api.NewAction(api.ActionTypeUpdateTLSSNI, group, m.ID, "SNI Secret needs update"))
				default:
					log.Warn().Msg("SNI mode rotation is unknown")
					continue
				}
			}
		}
		return nil
	})
	return plan
}
