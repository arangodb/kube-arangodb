//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	schedulerContainerResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container/resources"
	schedulerPodApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	schedulerPodResourcesApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod/resources"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
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
	divisor1m  = resource.MustParse("1m")
	divisor1Mi = resource.MustParse("1Mi")
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

	gen := func(name, version string, generator func() (sidecar.Integration, bool)) func() (string, *schedulerApi.ArangoProfile, error) {
		return func() (string, *schedulerApi.ArangoProfile, error) {
			counterMetric.Inc()
			fullName := fmt.Sprintf("%s-int-%s-%s", deploymentName, name, version)

			intgr, exists := generator()
			if !exists {
				return fullName, nil, nil
			}

			integration, err := sidecar.NewIntegrationEnablement(intgr)
			if err != nil {
				return "", nil, err
			}

			key, v := constants.NewProfileIntegration(name, version)

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
						constants.ProfilesDeployment: deploymentName,
						key:                          v,
					}),
					Template: integration,
				},
			}, nil
		}
	}

	always := func(in sidecar.Integration) func() (sidecar.Integration, bool) {
		return func() (sidecar.Integration, bool) {
			return in, true
		}
	}

	if changed, err := r.ensureArangoProfilesFactory(ctx, cachedStatus,
		func() (string, *schedulerApi.ArangoProfile, error) {
			counterMetric.Inc()
			name := fmt.Sprintf("%s-int", deploymentName)

			integration, err := sidecar.NewIntegration(&schedulerContainerResourcesApi.Image{
				Image: util.NewType(r.context.GetOperatorImage()),
			}, spec.Integration.GetSidecar(),
				r.arangoDeploymentProfileTemplate(cachedStatus),
				r.arangoDeploymentCATemplate(),
			)
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
		gen(constants.ProfilesIntegrationAuthz, constants.ProfilesIntegrationV0, always(sidecar.IntegrationAuthorizationV0{})),
		gen(constants.ProfilesIntegrationAuthn, constants.ProfilesIntegrationV1, always(sidecar.IntegrationAuthenticationV1{Spec: spec, DeploymentName: apiObject.GetName()})),
		gen(constants.ProfilesIntegrationSched, constants.ProfilesIntegrationV1, always(sidecar.IntegrationSchedulerV1{})),
		gen(constants.ProfilesIntegrationSched, constants.ProfilesIntegrationV2, always(sidecar.IntegrationSchedulerV2{
			Spec:           spec,
			DeploymentName: apiObject.GetName(),
		})),
		gen(constants.ProfilesIntegrationShutdown, constants.ProfilesIntegrationV1, always(sidecar.IntegrationShutdownV1{})),
		gen(constants.ProfilesIntegrationEnvoy, constants.ProfilesIntegrationV3, always(sidecar.IntegrationEnvoyV3{Spec: spec})),
		gen(constants.ProfilesIntegrationStorage, constants.ProfilesIntegrationV1, func() (sidecar.Integration, bool) {
			if v, err := cachedStatus.ArangoPlatformStorage().V1Alpha1(); err == nil {
				if p, ok := v.GetSimple(deploymentName); ok {
					if p.Status.Conditions.IsTrue(platformApi.ReadyCondition) {
						return sidecar.IntegrationStorageV1{
							PlatformStorage: p,
						}, true
					}
				}
			}

			return nil, false
		}),
		gen(constants.ProfilesIntegrationStorage, constants.ProfilesIntegrationV2, func() (sidecar.Integration, bool) {
			if v, err := cachedStatus.ArangoPlatformStorage().V1Alpha1(); err == nil {
				if p, ok := v.GetSimple(deploymentName); ok {
					if p.Status.Conditions.IsTrue(platformApi.ReadyCondition) {
						return sidecar.IntegrationStorageV2{
							Storage: p,
						}, true
					}
				}
			}

			return nil, false
		})); err != nil {
		return err
	} else if changed {
		reconcileRequired.Required()
	}

	return reconcileRequired.Reconcile(ctx)
}

func (r *Resources) arangoDeploymentInternalAddress(cachedStatus inspectorInterface.Inspector) string {
	spec := r.context.GetSpec()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()

	proto := util.BoolSwitch(spec.IsSecure(), "https", "http")
	svc, ok := cachedStatus.Service().V1().GetSimple(deploymentName)
	if !ok {
		return ""
	}

	if spec.CommunicationMethod.Get() == api.DeploymentCommunicationMethodIP {
		if ip := svc.Spec.ClusterIP; ip != core.ClusterIPNone && ip != "" {
			return fmt.Sprintf("%s://%s:%d", proto, ip, shared.ArangoPort)
		}
	}

	return fmt.Sprintf("%s://%s:%d", proto, k8sutil.CreateDatabaseClientServiceDNSNameWithDomain(svc, spec.ClusterDomain), shared.ArangoPort)
}

func (r *Resources) arangoDeploymentProfileTemplate(cachedStatus inspectorInterface.Inspector) *schedulerApi.ProfileTemplate {
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()

	var envs []core.EnvVar

	envs = append(envs,
		core.EnvVar{
			Name:  "ARANGO_DEPLOYMENT_NAME",
			Value: deploymentName,
		},
		core.EnvVar{
			Name:  "ARANGO_DEPLOYMENT_ENDPOINT",
			Value: r.arangoDeploymentInternalAddress(cachedStatus),
		},
		core.EnvVar{
			Name:  "ARANGODB_ENDPOINT",
			Value: r.arangoDeploymentInternalAddress(cachedStatus),
		},
	)

	if !r.context.GetSpec().IsAuthenticated() {
		envs = append(envs, core.EnvVar{
			Name:  "AUTHENTICATION_ENABLED",
			Value: r.arangoDeploymentInternalAddress(cachedStatus),
		})
	}

	return &schedulerApi.ProfileTemplate{
		Container: &schedulerApi.ProfileContainerTemplate{
			All: &schedulerContainerApi.Generic{
				Environments: &schedulerContainerResourcesApi.Environments{
					Env: []core.EnvVar{
						{
							Name:  "ARANGO_DEPLOYMENT_NAME",
							Value: deploymentName,
						},
						{
							Name:  "ARANGO_DEPLOYMENT_ENDPOINT",
							Value: r.arangoDeploymentInternalAddress(cachedStatus),
						},
						{
							Name:  "ARANGODB_ENDPOINT",
							Value: r.arangoDeploymentInternalAddress(cachedStatus),
						},
					},
				},
			},
		},
	}
}

func (r *Resources) tempalteKubernetesEnvs() *schedulerApi.ProfileTemplate {
	return &schedulerApi.ProfileTemplate{
		Container: &schedulerApi.ProfileContainerTemplate{
			All: &schedulerContainerApi.Generic{
				Environments: &schedulerContainerResourcesApi.Environments{
					Env: []core.EnvVar{
						{
							Name: "KUBERNETES_NAMESPACE",
							ValueFrom: &core.EnvVarSource{
								FieldRef: &core.ObjectFieldSelector{
									FieldPath: "metadata.namespace",
								},
							},
						},
						{
							Name: "KUBERNETES_POD_NAME",
							ValueFrom: &core.EnvVarSource{
								FieldRef: &core.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name: "KUBERNETES_POD_IP",
							ValueFrom: &core.EnvVarSource{
								FieldRef: &core.ObjectFieldSelector{
									FieldPath: "status.podIP",
								},
							},
						},
						{
							Name: "KUBERNETES_SERVICE_ACCOUNT",
							ValueFrom: &core.EnvVarSource{
								FieldRef: &core.ObjectFieldSelector{
									FieldPath: "spec.serviceAccountName",
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *Resources) templateResourceEnvs() *schedulerApi.ProfileTemplate {
	return &schedulerApi.ProfileTemplate{
		Container: &schedulerApi.ProfileContainerTemplate{
			All: &schedulerContainerApi.Generic{
				Environments: &schedulerContainerResourcesApi.Environments{
					Env: []core.EnvVar{

						{
							Name: "CONTAINER_CPU_REQUESTS",
							ValueFrom: &core.EnvVarSource{
								ResourceFieldRef: &core.ResourceFieldSelector{
									Resource: "requests.cpu",
									Divisor:  divisor1m,
								},
							},
						},
						{
							Name: "CONTAINER_MEMORY_REQUESTS",
							ValueFrom: &core.EnvVarSource{
								ResourceFieldRef: &core.ResourceFieldSelector{
									Resource: "requests.memory",
									Divisor:  divisor1Mi,
								},
							},
						},
						{
							Name: "CONTAINER_CPU_LIMITS",
							ValueFrom: &core.EnvVarSource{
								ResourceFieldRef: &core.ResourceFieldSelector{
									Resource: "limits.cpu",
									Divisor:  divisor1m,
								},
							},
						},
						{
							Name: "CONTAINER_MEMORY_LIMITS",
							ValueFrom: &core.EnvVarSource{
								ResourceFieldRef: &core.ResourceFieldSelector{
									Resource: "limits.memory",
									Divisor:  divisor1Mi,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *Resources) arangoDeploymentCATemplate() *schedulerApi.ProfileTemplate {
	t := r.context.GetSpec().TLS
	if !t.IsSecure() {
		return nil
	}

	return &schedulerApi.ProfileTemplate{
		Pod: &schedulerPodApi.Pod{
			Volumes: &schedulerPodResourcesApi.Volumes{
				Volumes: []core.Volume{
					{
						Name: "deployment-int-ca",
						VolumeSource: core.VolumeSource{
							Secret: &core.SecretVolumeSource{
								SecretName: GetCASecretName(r.context.GetAPIObject()),
							},
						},
					},
				},
			},
		},
		Container: &schedulerApi.ProfileContainerTemplate{
			All: &schedulerContainerApi.Generic{
				Environments: &schedulerContainerResourcesApi.Environments{
					Env: []core.EnvVar{
						{
							Name:  "ARANGO_DEPLOYMENT_CA",
							Value: fmt.Sprintf("/etc/deployment-int/ca/%s", CACertName),
						},
					},
				},
				VolumeMounts: &schedulerContainerResourcesApi.VolumeMounts{
					VolumeMounts: []core.VolumeMount{
						{
							Name:              "deployment-int-ca",
							ReadOnly:          true,
							RecursiveReadOnly: nil,
							MountPath:         "/etc/deployment-int/ca",
						},
					},
				},
			},
		},
	}
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

	if expected != nil && expected.GetName() != name {
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
