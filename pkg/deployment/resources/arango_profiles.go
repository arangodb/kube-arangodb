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

	"k8s.io/apimachinery/pkg/api/equality"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/integrations/sidecar"
	"github.com/arangodb/kube-arangodb/pkg/metrics"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/patcher"
)

var (
	inspectedArangoProfilesCounters     = metrics.MustRegisterCounterVec(metricsComponent, "inspected_arango_profiles", "Number of ArangoProfiles inspections per deployment", metrics.DeploymentName)
	inspectArangoProfilesDurationGauges = metrics.MustRegisterGaugeVec(metricsComponent, "inspect_arango_profiles_duration", "Amount of time taken by a single inspection of all ArangoProfiles for a deployment (in sec)", metrics.DeploymentName)
)

func matchArangoProfilesLabels(labels map[string]string) *schedulerApi.ProfileSelectors {
	if labels == nil {
		return nil
	}

	return &schedulerApi.ProfileSelectors{
		Label: &meta.LabelSelector{
			MatchLabels: labels,
		},
	}
}

// EnsureArangoProfiles creates all ArangoProfiles needed to run the given deployment
func (r *Resources) EnsureArangoProfiles(ctx context.Context, cachedStatus inspectorInterface.Inspector) error {
	start := time.Now()
	spec := r.context.GetSpec()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()

	defer metrics.SetDuration(inspectArangoProfilesDurationGauges.WithLabelValues(deploymentName), start)
	counterMetric := inspectedArangoProfilesCounters.WithLabelValues(deploymentName)

	reconcileRequired := k8sutil.NewReconcile(cachedStatus)

	gen := func(name, version string, integrations ...sidecar.Integration) func() (string, *schedulerApi.ArangoProfile, error) {
		return func() (string, *schedulerApi.ArangoProfile, error) {
			counterMetric.Inc()
			fullName := fmt.Sprintf("%s-int-%s-%s", deploymentName, name, version)

			integration, err := sidecar.NewIntegrationEnablement(integrations...)
			if err != nil {
				return "", nil, err
			}

			return fullName, &schedulerApi.ArangoProfile{
				ObjectMeta: meta.ObjectMeta{
					Name:      fullName,
					Namespace: apiObject.GetNamespace(),
					OwnerReferences: []meta.OwnerReference{
						apiObject.AsOwner(),
					},
				},
				Spec: schedulerApi.ProfileSpec{
					Selectors: matchArangoProfilesLabels(map[string]string{
						constants.ProfilesDeployment:                                    deploymentName,
						fmt.Sprintf("%s/%s", constants.ProfilesIntegrationPrefix, name): version,
					}),
					Template: integration,
				},
			}, nil
		}
	}

	if changed, err := r.ensureArangoProfilesFactory(ctx, cachedStatus,
		func() (string, *schedulerApi.ArangoProfile, error) {
			counterMetric.Inc()
			name := fmt.Sprintf("%s-int", deploymentName)

			integration, err := sidecar.NewIntegration(&schedulerContainerResourcesApi.Image{
				Image: util.NewType(r.context.GetOperatorImage()),
			}, spec.Integration.GetSidecar())
			if err != nil {
				return "", nil, err
			}

			return name, &schedulerApi.ArangoProfile{
				ObjectMeta: meta.ObjectMeta{
					Name:      name,
					Namespace: apiObject.GetNamespace(),
					OwnerReferences: []meta.OwnerReference{
						apiObject.AsOwner(),
					},
				},
				Spec: schedulerApi.ProfileSpec{
					Selectors: matchArangoProfilesLabels(map[string]string{
						constants.ProfilesDeployment: deploymentName,
					}),
					Template: integration,
				},
			}, nil
		},
		gen("authz", "v0", sidecar.IntegrationAuthorizationV0{}),
		gen("authn", "v1", sidecar.IntegrationAuthenticationV1{Spec: spec, DeploymentName: apiObject.GetName()}),
	); err != nil {
		return err
	} else if changed {
		reconcileRequired.Required()
	}

	return reconcileRequired.Reconcile(ctx)
}

func (r *Resources) ensureArangoProfilesFactory(ctx context.Context, cachedStatus inspectorInterface.Inspector, expected ...func() (string, *schedulerApi.ArangoProfile, error)) (bool, error) {
	var changed bool

	for _, e := range expected {
		name, profile, err := e()
		if err != nil {
			return false, err
		}
		if c, err := r.ensureArangoProfile(ctx, cachedStatus, name, profile); err != nil {
			return false, err
		} else if c {
			changed = true
		}
	}

	return changed, nil
}

func (r *Resources) ensureArangoProfile(ctx context.Context, cachedStatus inspectorInterface.Inspector, name string, expected *schedulerApi.ArangoProfile) (bool, error) {
	arangoProfiles := cachedStatus.ArangoProfileModInterface().V1Beta1()

	if expected.GetName() != name {
		return false, errors.Errorf("Name mismatch")
	}

	if c, err := cachedStatus.ArangoProfile().V1Beta1(); err == nil {
		if s, ok := c.GetSimple(name); !ok {
			if expected != nil {
				if _, err := arangoProfiles.Create(ctx, expected, meta.CreateOptions{}); err != nil {
					return false, err
				}

				return true, nil
			}
		} else {
			if expected == nil {
				if err := arangoProfiles.Delete(ctx, s.GetName(), meta.DeleteOptions{}); err != nil {
					if !kerrors.IsNotFound(err) {
						return false, err
					}
				}

				return true, nil
			}
			expectedChecksum, err := expected.Spec.Template.Checksum()
			if err != nil {
				return false, err
			}

			currChecksum, err := s.Spec.Template.Checksum()
			if err != nil {
				return false, err
			}

			if expected.Spec.Selectors == nil && s.Spec.Selectors != nil {
				// Remove
				if _, changed, err := patcher.Patcher[*schedulerApi.ArangoProfile](ctx, arangoProfiles, s, meta.PatchOptions{},
					func(in *schedulerApi.ArangoProfile) []patch.Item {
						return []patch.Item{
							patch.ItemRemove(patch.NewPath("spec", "selectors")),
						}
					}); err != nil {
					return false, err
				} else if changed {
					return true, nil
				}
			} else if !equality.Semantic.DeepEqual(expected.Spec.Selectors, s.Spec.Selectors) {
				if _, changed, err := patcher.Patcher[*schedulerApi.ArangoProfile](ctx, arangoProfiles, s, meta.PatchOptions{},
					func(in *schedulerApi.ArangoProfile) []patch.Item {
						return []patch.Item{
							patch.ItemReplace(patch.NewPath("spec", "selectors"), expected.Spec.Selectors),
						}
					}); err != nil {
					return false, err
				} else if changed {
					return true, nil
				}
			}

			if currChecksum != expectedChecksum {
				if _, changed, err := patcher.Patcher[*schedulerApi.ArangoProfile](ctx, arangoProfiles, s, meta.PatchOptions{},
					func(in *schedulerApi.ArangoProfile) []patch.Item {
						return []patch.Item{
							patch.ItemReplace(patch.NewPath("spec", "template"), util.TypeOrDefault(expected.Spec.Template)),
						}
					}); err != nil {
					return false, err
				} else if changed {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
