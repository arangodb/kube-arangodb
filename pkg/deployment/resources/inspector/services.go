//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package inspector

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (i *inspector) IterateServices(action service.ServiceAction, filters ...service.ServiceFilter) error {
	for _, service := range i.Services() {
		if err := i.iterateServices(service, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateServices(service *core.Service, action service.ServiceAction, filters ...service.ServiceFilter) error {
	for _, filter := range filters {
		if !filter(service) {
			return nil
		}
	}

	return action(service)
}

func (i *inspector) Services() []*core.Service {
	i.lock.Lock()
	defer i.lock.Unlock()

	var r []*core.Service
	for _, service := range i.services {
		r = append(r, service)
	}

	return r
}

func (i *inspector) Service(name string) (*core.Service, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	service, ok := i.services[name]
	if !ok {
		return nil, false
	}

	return service, true
}

func servicesToMap(ctx context.Context, k kubernetes.Interface, namespace string) (map[string]*core.Service, error) {
	services, err := getServices(ctx, k, namespace, "")
	if err != nil {
		return nil, err
	}

	serviceMap := map[string]*core.Service{}

	for _, service := range services {
		_, exists := serviceMap[service.GetName()]
		if exists {
			return nil, errors.Newf("Service %s already exists in map, error received", service.GetName())
		}

		serviceMap[service.GetName()] = servicePointer(service)
	}

	return serviceMap, nil
}

func servicePointer(pod core.Service) *core.Service {
	return &pod
}

func getServices(ctx context.Context, k kubernetes.Interface, namespace, cont string) ([]core.Service, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	services, err := k.CoreV1().Services(namespace).List(ctxChild, meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})
	cancel()

	if err != nil {
		return nil, err
	}

	if services.Continue != "" {
		nextServicesLayer, err := getServices(ctx, k, namespace, services.Continue)
		if err != nil {
			return nil, err
		}

		return append(services.Items, nextServicesLayer...), nil
	}

	return services.Items, nil
}
