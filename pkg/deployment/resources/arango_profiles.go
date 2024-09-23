//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"
	"fmt"
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/integrations/sidecar"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

var (
	inspectedArangoProfilesCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_arango_profiles", "Number of ArangoProfiles inspections per deployment", metrics.DeploymentName)
	inspectArangoProfilesDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_arango_profiles_duration", "Amount of time taken by a single inspection of all ArangoProfiles for a deployment (in sec)", metrics.DeploymentName)
)

// EnsureArangoProfiles creates all ArangoProfiles needed to run the given deployment
func (r *Resources) EnsureArangoProfiles(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	start := time.Now()
	spec := r.context.GetSpec()
	arangoProfiles := cachedStatus.ArangoProfileModInterface().V1Beta1()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()

	defer metrics.SetDuration(inspectArangoProfilesDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedArangoProfilesCounters.WithLabelValues(deploymentName)

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	intName := fmt.Sprintf("%s-int", deploymentName)

	integration, err := sidecar.NewIntegration(&schedulerContainerResourcesApi.Image{
		Image: util.NewType(r.context.GetOperatorImage()),
	}, spec.Integration.GetSidecar())
	if err != nil {
		return err
	}

	integrationChecksum, err := integration.Checksum()
	if err != nil {
		return err
	}

	if c, err := cachedStatus.ArangoProfile().V1Beta1(); err == nil {
		counterMetric.Inc()
		if s, ok := c.GetSimple(intName); !ok {
			s = &schedulerApi.ArangoProfile{
				ObjectMeta: meta.ObjectMeta{
					Name:      intName,
					Namespace: apiObject.GetNamespace(),
					OwnerReferences: []meta.OwnerReference{
						apiObject.AsOwner(),
					},
				},
				Spec: schedulerApi.ProfileSpec{
					Template: integration,
				},
			}

			if _, err := cachedStatus.ArangoProfileModInterface().V1Beta1().Create(ctx, s, meta.CreateOptions{}); err != nil {
				return err
			}

			reconcileRequired.Required()
		} else {
			currChecksum, err := s.Spec.Template.Checksum()
			if err != nil {
				return err
			}

			if s.Spec.Selectors != nil {
				if _, changed, err := patcher.Patcher[*schedulerApi.ArangoProfile](ctx, arangoProfiles, s, meta.PatchOptions{},
					func(in *schedulerApi.ArangoProfile) []patch.Item {
						return []patch.Item{
							patch.ItemRemove(patch.NewPath("spec", "selectors")),
						}
					}); err != nil {
					return err
				} else if changed {
					reconcileRequired.Required()
				}
			}

			if currChecksum != integrationChecksum {
				if _, changed, err := patcher.Patcher[*schedulerApi.ArangoProfile](ctx, arangoProfiles, s, meta.PatchOptions{},
					func(in *schedulerApi.ArangoProfile) []patch.Item {
						return []patch.Item{
							patch.ItemReplace(patch.NewPath("spec", "template"), integration),
						}
					}); err != nil {
					return err
				} else if changed {
					reconcileRequired.Required()
				}
			}
		}
	}

	return reconcileRequired.Reconcile(ctx)
}
