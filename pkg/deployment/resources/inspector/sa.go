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

package inspector

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (i *inspector) IterateServiceAccounts(action serviceaccount.Action, filters ...serviceaccount.Filter) error {
	for _, serviceAccount := range i.ServiceAccounts() {
		if err := i.iterateServiceAccount(serviceAccount, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateServiceAccount(serviceAccount *core.ServiceAccount, action serviceaccount.Action, filters ...serviceaccount.Filter) error {
	for _, filter := range filters {
		if !filter(serviceAccount) {
			return nil
		}
	}

	return action(serviceAccount)
}

func (i *inspector) ServiceAccounts() []*core.ServiceAccount {
	i.lock.Lock()
	defer i.lock.Unlock()

	var r []*core.ServiceAccount
	for _, serviceAccount := range i.serviceAccounts {
		r = append(r, serviceAccount)
	}

	return r
}

func (i *inspector) ServiceAccount(name string) (*core.ServiceAccount, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	serviceAccount, ok := i.serviceAccounts[name]
	if !ok {
		return nil, false
	}

	return serviceAccount, true
}

func serviceAccountsToMap(k kubernetes.Interface, namespace string) (map[string]*core.ServiceAccount, error) {
	serviceAccounts, err := getServiceAccounts(k, namespace, "")
	if err != nil {
		return nil, err
	}

	serviceAccountMap := map[string]*core.ServiceAccount{}

	for _, serviceAccount := range serviceAccounts {
		_, exists := serviceAccountMap[serviceAccount.GetName()]
		if exists {
			return nil, errors.Newf("ServiceAccount %s already exists in map, error received", serviceAccount.GetName())
		}

		serviceAccountMap[serviceAccount.GetName()] = serviceAccountPointer(serviceAccount)
	}

	return serviceAccountMap, nil
}

func serviceAccountPointer(serviceAccount core.ServiceAccount) *core.ServiceAccount {
	return &serviceAccount
}

func getServiceAccounts(k kubernetes.Interface, namespace, cont string) ([]core.ServiceAccount, error) {
	serviceAccounts, err := k.CoreV1().ServiceAccounts(namespace).List(context.Background(), meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if serviceAccounts.Continue != "" {
		nextServiceAccountsLayer, err := getServiceAccounts(k, namespace, serviceAccounts.Continue)
		if err != nil {
			return nil, err
		}

		return append(serviceAccounts.Items, nextServiceAccountsLayer...), nil
	}

	return serviceAccounts.Items, nil
}

func FilterServiceAccountsByLabels(labels map[string]string) serviceaccount.Filter {
	return func(serviceAccount *core.ServiceAccount) bool {
		for key, value := range labels {
			v, ok := serviceAccount.Labels[key]
			if !ok {
				return false
			}

			if v != value {
				return false
			}
		}

		return true
	}
}
