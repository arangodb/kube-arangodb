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

func (h *handler) HandleArangoDestinationEndpoints(ctx context.Context, item operation.Item, extension *networkingApi.ArangoRoute, status *networkingApi.ArangoRouteStatus, deployment *api.ArangoDeployment, dest *networkingApi.ArangoRouteSpecDestination, endpoints *networkingApi.ArangoRouteSpecDestinationEndpoints) (*operator.Condition, bool, error) {
	port := endpoints.Port

	if port == nil {
		return &operator.Condition{
			Status:  false,
			Reason:  "Destination Not Found",
			Message: "Missing Port definition",
		}, false, nil
	}

	s, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.kubeClient.CoreV1().Services(endpoints.GetNamespace(extension)).Get, endpoints.GetName(), meta.GetOptions{})
	if err != nil {
		if api.IsNotFound(err) {
			return &operator.Condition{
				Status:  false,
				Reason:  "Destination Not Found",
				Message: fmt.Sprintf("Service `%s/%s` Not found", endpoints.GetNamespace(extension), endpoints.GetName()),
			}, false, nil
		}

		return &operator.Condition{
			Status:  false,
			Reason:  "Destination Not Found",
			Message: fmt.Sprintf("Unknown error for service `%s/%s`: %s", endpoints.GetNamespace(extension), endpoints.GetName(), err.Error()),
		}, false, nil
	}

	if !endpoints.Equals(s) {
		return &operator.Condition{
			Status:  false,
			Reason:  "Destination Not Found",
			Message: fmt.Sprintf("Service `%s/%s` Changed", endpoints.GetNamespace(extension), endpoints.GetName()),
		}, false, nil
	}

	e, err := util.WithKubernetesContextTimeoutP2A2(ctx, h.kubeClient.CoreV1().Endpoints(endpoints.GetNamespace(extension)).Get, endpoints.GetName(), meta.GetOptions{})
	if err != nil {
		if api.IsNotFound(err) {
			return &operator.Condition{
				Status:  false,
				Reason:  "Destination Not Found",
				Message: fmt.Sprintf("Endpoints `%s/%s` Not found", endpoints.GetNamespace(extension), endpoints.GetName()),
			}, false, nil
		}

		return &operator.Condition{
			Status:  false,
			Reason:  "Destination Not Found",
			Message: fmt.Sprintf("Unknown error for endpoints `%s/%s`: %s", endpoints.GetNamespace(extension), endpoints.GetName(), err.Error()),
		}, false, nil
	}

	// Discover port name - empty names are allowed
	var destPortName = "N/A"

	if port.Type == intstr.Int {
		p, ok := util.PickFromList(s.Spec.Ports, func(v core.ServicePort) bool {
			return v.Port == port.IntVal
		})
		if !ok {
			return &operator.Condition{
				Status:  false,
				Reason:  "Destination Not Found",
				Message: fmt.Sprintf("Port `%d` not defined on Service `%s/%s`", port.IntVal, endpoints.GetNamespace(extension), endpoints.GetName()),
			}, false, nil
		}

		destPortName = p.Name
	} else if port.Type == intstr.String {
		destPortName = port.StrVal
	}

	if destPortName == "N/A" {
		return &operator.Condition{
			Status:  false,
			Reason:  "Destination Not Found",
			Message: fmt.Sprintf("Unable to discover port on Service `%s/%s`", endpoints.GetNamespace(extension), endpoints.GetName()),
		}, false, nil
	}

	var target networkingApi.ArangoRouteStatusTarget

	target.Path = dest.GetPath()
	target.Type = networkingApi.ArangoRouteStatusTargetEndpointsType
	target.Protocol = dest.GetProtocol().Get()

	target.Options = extension.Spec.Options.AsStatus()

	// Render Auth Settings

	target.Authentication.Type = dest.GetAuthentication().GetType()
	target.Authentication.PassMode = dest.GetAuthentication().GetPassMode()

	if dest.Schema.Get() == networkingApi.ArangoRouteSpecDestinationSchemaHTTPS {
		target.TLS = &networkingApi.ArangoRouteStatusTargetTLS{
			Insecure: util.NewType(extension.Spec.Destination.GetTLS().GetInsecure()),
		}
	}

	for _, subset := range e.Subsets {
		p, ok := util.PickFromList(subset.Ports, func(v core.EndpointPort) bool {
			return v.Name == destPortName
		})
		if !ok {
			continue
		}

		for _, address := range subset.Addresses {
			target.Destinations = append(target.Destinations, networkingApi.ArangoRouteStatusTargetDestination{
				Host: address.IP,
				Port: p.Port,
			})
		}

		if s.Spec.PublishNotReadyAddresses {
			for _, address := range subset.NotReadyAddresses {
				target.Destinations = append(target.Destinations, networkingApi.ArangoRouteStatusTargetDestination{
					Host: address.IP,
					Port: p.Port,
				})
			}
		}
	}

	target.Destinations = util.Sort(target.Destinations, func(i, j networkingApi.ArangoRouteStatusTargetDestination) bool {
		return i.Hash() < j.Hash()
	})

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
