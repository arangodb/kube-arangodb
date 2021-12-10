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

	"github.com/arangodb/kube-arangodb/pkg/util/globals"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

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

func (i *inspector) ServiceAccountReadInterface() serviceaccount.ReadInterface {
	return &serviceAccountReadInterface{i: i}
}

type serviceAccountReadInterface struct {
	i *inspector
}

func (s serviceAccountReadInterface) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.ServiceAccount, error) {
	if s, ok := s.i.ServiceAccount(name); !ok {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    core.GroupName,
			Resource: "serviceaccounts",
		}, name)
	} else {
		return s, nil
	}
}

func serviceAccountsToMap(ctx context.Context, inspector *inspector, k kubernetes.Interface, namespace string) func() error {
	return func() error {
		serviceAccounts, err := getServiceAccounts(ctx, k, namespace, "")
		if err != nil {
			return err
		}

		serviceAccountMap := map[string]*core.ServiceAccount{}

		for _, serviceAccount := range serviceAccounts {
			_, exists := serviceAccountMap[serviceAccount.GetName()]
			if exists {
				return errors.Newf("ServiceAccount %s already exists in map, error received", serviceAccount.GetName())
			}

			serviceAccountMap[serviceAccount.GetName()] = serviceAccountPointer(serviceAccount)
		}

		inspector.serviceAccounts = serviceAccountMap

		return nil
	}
}

func serviceAccountPointer(serviceAccount core.ServiceAccount) *core.ServiceAccount {
	return &serviceAccount
}

func getServiceAccounts(ctx context.Context, k kubernetes.Interface, namespace, cont string) ([]core.ServiceAccount, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	serviceAccounts, err := k.CoreV1().ServiceAccounts(namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if serviceAccounts.Continue != "" {
		nextServiceAccountsLayer, err := getServiceAccounts(ctx, k, namespace, serviceAccounts.Continue)
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
