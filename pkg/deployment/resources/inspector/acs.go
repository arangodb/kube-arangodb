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

package inspector

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoclustersynchronization"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (i *inspector) GetArangoClusterSynchronizations() (arangoclustersynchronization.Inspector, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.acs == nil {
		return nil, false
	}

	return i.acs, i.acs.accessible
}

type arangoClusterSynchronizationLoader struct {
	accessible bool

	acs map[string]*api.ArangoClusterSynchronization
}

func (a *arangoClusterSynchronizationLoader) FilterArangoClusterSynchronizations(filters ...arangoclustersynchronization.Filter) []*api.ArangoClusterSynchronization {
	q := make([]*api.ArangoClusterSynchronization, 0, len(a.acs))

	for _, obj := range a.acs {
		if a.filterArangoClusterSynchronizations(obj, filters...) {
			q = append(q, obj)
		}
	}

	return q
}

func (a *arangoClusterSynchronizationLoader) filterArangoClusterSynchronizations(obj *api.ArangoClusterSynchronization, filters ...arangoclustersynchronization.Filter) bool {
	for _, f := range filters {
		if !f(obj) {
			return false
		}
	}

	return true
}

func (a *arangoClusterSynchronizationLoader) ArangoClusterSynchronizations() []*api.ArangoClusterSynchronization {
	var r []*api.ArangoClusterSynchronization
	for _, acs := range a.acs {
		r = append(r, acs)
	}

	return r
}

func (a *arangoClusterSynchronizationLoader) ArangoClusterSynchronization(name string) (*api.ArangoClusterSynchronization, bool) {
	acs, ok := a.acs[name]
	if !ok {
		return nil, false
	}

	return acs, true
}

func (a *arangoClusterSynchronizationLoader) IterateArangoClusterSynchronizations(action arangoclustersynchronization.Action, filters ...arangoclustersynchronization.Filter) error {
	for _, node := range a.ArangoClusterSynchronizations() {
		if err := a.iterateArangoClusterSynchronization(node, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (a *arangoClusterSynchronizationLoader) iterateArangoClusterSynchronization(acs *api.ArangoClusterSynchronization, action arangoclustersynchronization.Action, filters ...arangoclustersynchronization.Filter) error {
	for _, filter := range filters {
		if !filter(acs) {
			return nil
		}
	}

	return action(acs)
}

func (a *arangoClusterSynchronizationLoader) ArangoClusterSynchronizationReadInterface() arangoclustersynchronization.ReadInterface {
	return &arangoClusterSynchronizationReadInterface{i: a}
}

type arangoClusterSynchronizationReadInterface struct {
	i *arangoClusterSynchronizationLoader
}

func (a *arangoClusterSynchronizationReadInterface) Get(ctx context.Context, name string, opts meta.GetOptions) (*api.ArangoClusterSynchronization, error) {
	if s, ok := a.i.ArangoClusterSynchronization(name); !ok {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    deployment.ArangoDeploymentGroupName,
			Resource: "arangoclustersynchronizations",
		}, name)
	} else {
		return s, nil
	}
}

func arangoClusterSynchronizationPointer(acs api.ArangoClusterSynchronization) *api.ArangoClusterSynchronization {
	return &acs
}

func arangoClusterSynchronizationsToMap(ctx context.Context, inspector *inspector, k versioned.Interface, namespace string) func() error {
	return func() error {
		acss, err := getArangoClusterSynchronizations(ctx, k, namespace, "")
		if err != nil {
			if apiErrors.IsUnauthorized(err) || apiErrors.IsNotFound(err) {
				inspector.acs = &arangoClusterSynchronizationLoader{
					accessible: false,
				}
				return nil
			}
			return err
		}

		acssMap := map[string]*api.ArangoClusterSynchronization{}

		for _, acs := range acss {
			_, exists := acssMap[acs.GetName()]
			if exists {
				return errors.Newf("ArangoMember %s already exists in map, error received", acs.GetName())
			}

			acssMap[acs.GetName()] = arangoClusterSynchronizationPointer(acs)
		}

		inspector.acs = &arangoClusterSynchronizationLoader{
			accessible: true,
			acs:        acssMap,
		}

		return nil
	}
}

func getArangoClusterSynchronizations(ctx context.Context, k versioned.Interface, namespace, cont string) ([]api.ArangoClusterSynchronization, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	acss, err := k.DatabaseV1().ArangoClusterSynchronizations(namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if acss.Continue != "" {
		newACSLoader, err := getArangoClusterSynchronizations(ctx, k, namespace, acss.Continue)
		if err != nil {
			return nil, err
		}

		return append(acss.Items, newACSLoader...), nil
	}

	return acss.Items, nil
}
