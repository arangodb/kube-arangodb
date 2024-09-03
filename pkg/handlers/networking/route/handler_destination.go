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

package route

import (
	"context"
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (h *handler) HandleArangoDestination(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus, deployment *api.ArangoDeployment) (*operator.Condition, bool, error) {
	if dest := extension.Spec.GetDestination(); dest != nil {
		if svc := dest.GetService(); svc != nil {
			port := svc.Port

			if port == nil {
				return &operator.Condition{
					Status:  false,
					Reason:  "Destination Not Found",
					Message: "Missing Port definition",
				}, false, nil
			}

			s, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.kubeClient.CoreV1().Services(svc.GetNamespace(extension)).Get, svc.GetName(), meta.GetOptions{})
			if err != nil {
				if api.IsNotFound(err) {
					return &operator.Condition{
						Status:  false,
						Reason:  "Destination Not Found",
						Message: fmt.Sprintf("Service `%s/%s` Not found", svc.GetNamespace(extension), svc.GetName()),
					}, false, nil
				}

				return &operator.Condition{
					Status:  false,
					Reason:  "Destination Not Found",
					Message: fmt.Sprintf("Unknown error for service `%s/%s`: %s", svc.GetNamespace(extension), svc.GetName(), err.Error()),
				}, false, nil
			}

			if !svc.Equals(s) {
				return &operator.Condition{
					Status:  false,
					Reason:  "Destination Not Found",
					Message: fmt.Sprintf("Service `%s/%s` Changed", svc.GetNamespace(extension), svc.GetName()),
				}, false, nil
			}

			var destPort int32

			if port.Type == intstr.Int {
				p, ok := util.PickFromList(s.Spec.Ports, func(v core.ServicePort) bool {
					return v.Port == port.IntVal
				})
				if !ok {
					return &operator.Condition{
						Status:  false,
						Reason:  "Destination Not Found",
						Message: fmt.Sprintf("Port `%d` not defined on Service `%s/%s`", port.IntVal, svc.GetNamespace(extension), svc.GetName()),
					}, false, nil
				}

				destPort = p.Port
			} else if port.Type == intstr.String && port.StrVal != "" {
				p, ok := util.PickFromList(s.Spec.Ports, func(v core.ServicePort) bool {
					return v.Name == port.StrVal
				})
				if !ok {
					return &operator.Condition{
						Status:  false,
						Reason:  "Destination Not Found",
						Message: fmt.Sprintf("Port `%s` not defined on Service `%s/%s`", port.StrVal, svc.GetNamespace(extension), svc.GetName()),
					}, false, nil
				}

				destPort = p.Port
			} else {
				return &operator.Condition{
					Status:  false,
					Reason:  "Destination Not Found",
					Message: "Unknown Port definition",
				}, false, nil
			}

			if destPort == -1 {
				return &operator.Condition{
					Status:  false,
					Reason:  "Destination Not Found",
					Message: fmt.Sprintf("Unable to discover port on Service `%s/%s`", svc.GetNamespace(extension), svc.GetName()),
				}, false, nil
			}

			var target networkingApi.ArangoRouteStatusTarget

			target.Path = dest.GetPath()

			// Render Auth Settings

			target.Authentication.Type = dest.GetAuthentication().GetType()

			if dest.Schema.Get() == networkingApi.ArangoRouteSpecDestinationSchemaHTTPS {
				target.TLS = &networkingApi.ArangoRouteStatusTargetTLS{
					Insecure: util.NewType(extension.Spec.Destination.GetTLS().GetInsecure()),
				}
			}

			if ip := s.Spec.ClusterIP; ip != "" {
				target.Destinations = networkingApi.ArangoRouteStatusTargetDestinations{
					networkingApi.ArangoRouteStatusTargetDestination{
						Host: ip,
						Port: destPort,
					},
				}
			} else {
				if domain := deployment.Spec.ClusterDomain; domain != nil {
					target.Destinations = networkingApi.ArangoRouteStatusTargetDestinations{
						networkingApi.ArangoRouteStatusTargetDestination{
							Host: fmt.Sprintf("%s.%s.svc.%s", s.GetName(), s.GetNamespace(), *domain),
							Port: destPort,
						},
					}
				} else {
					target.Destinations = networkingApi.ArangoRouteStatusTargetDestinations{
						networkingApi.ArangoRouteStatusTargetDestination{
							Host: fmt.Sprintf("%s.%s.svc", s.GetName(), s.GetNamespace()),
							Port: destPort,
						},
					}
				}
			}

			if status.Target.Hash() == target.Hash() {
				return &operator.Condition{
					Status:  true,
					Reason:  "Destination Found",
					Message: "Destination Found",
					Hash:    target.Hash(),
				}, false, nil
			}

			status.Target = &target
			return &operator.Condition{
				Status:  true,
				Reason:  "Destination Found",
				Message: "Destination Found",
				Hash:    target.Hash(),
			}, true, nil
		}
	}

	return &operator.Condition{
		Status:  false,
		Reason:  "Destination Not Found",
		Message: "Destination Not Found",
	}, false, nil
}

func (h *handler) HandleArangoDestinationWithTargets(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus, depl *api.ArangoDeployment) (*operator.Condition, bool, error) {
	c, changed, err := h.HandleArangoDestination(ctx, item, extension, status, depl)
	if c == nil && !c.Status && status.Target != nil {
		status.Target = nil
		changed = true
	}

	return c, changed, err
}

func (h *handler) HandleDestinationRequired(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus, _ *api.ArangoDeployment) (bool, error) {
	if !status.Conditions.IsTrue(networkingApi.DestinationValidCondition) {
		return false, operator.Stop("Destination is not ready")
	}

	return false, nil
}
