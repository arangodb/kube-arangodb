//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/actions"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

func (r *Reconciler) createRotateTLSServerSNIPlan(ctx context.Context, apiObject k8sutil.APIObject,
	spec api.DeploymentSpec, status api.DeploymentStatus,
	planCtx PlanBuilderContext) api.Plan {
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

	fetchedSecrets, err := mapTLSSNIConfig(*sni, planCtx.ACS().CurrentClusterCache())
	if err != nil {
		r.planLogger.Err(err).Warn("Unable to get SNI desired state")
		return nil
	}

	var plan api.Plan
	for _, group := range api.AllServerGroups {
		if !pod.GroupSNISupported(spec.Mode.Get(), group) {
			continue
		}
		for _, m := range status.Members.MembersOfGroup(group) {
			if !plan.IsEmpty() {
				// Only 1 member at a time
				break
			}

			if m.Phase != api.MemberPhaseCreated {
				// Only make changes when phase is created
				continue
			}

			if i, ok := status.Images.GetByImageID(m.ImageID); !ok || !features.EncryptionRotation().Supported(i.ArangoDBVersion, i.Enterprise) {
				continue
			}

			var c driver.Client
			err := globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				var err error
				c, err = planCtx.GetMembersState().GetMemberClient(m.ID)

				return err
			})
			if err != nil {
				r.planLogger.Err(err).Info("Unable to get client")
				continue
			}

			var ok bool
			err = globals.GetGlobalTimeouts().ArangoD().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				var err error
				ok, err = compareTLSSNIConfig(ctxChild, r.log, c.Connection(), fetchedSecrets, false)
				return err
			})
			if err != nil {
				r.planLogger.Err(err).Info("SNI compare failed")
				break
			} else if !ok {
				switch spec.TLS.Mode.Get() {
				case api.TLSRotateModeRecreate:
					plan = append(plan, tlsRotateConditionAction(group, m.ID, "SNI Secret needs update"))
				case api.TLSRotateModeInPlace:
					plan = append(plan,
						actions.NewAction(api.ActionTypeUpdateTLSSNI, group, m, "SNI Secret needs update"))
				default:
					r.planLogger.Warn("SNI mode rotation is unknown")
					continue
				}
			}
		}
	}
	return plan
}
