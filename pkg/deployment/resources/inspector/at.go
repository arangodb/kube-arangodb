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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangotask"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (i *inspector) GetArangoTasks() (arangotask.Inspector, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.at == nil {
		return nil, false
	}

	return i.at, i.at.accessible
}

type arangoTaskLoader struct {
	accessible bool

	at map[string]*api.ArangoTask
}

func (a *arangoTaskLoader) FilterArangoTasks(filters ...arangotask.Filter) []*api.ArangoTask {
	q := make([]*api.ArangoTask, 0, len(a.at))

	for _, obj := range a.at {
		if a.filterArangoTasks(obj, filters...) {
			q = append(q, obj)
		}
	}

	return q
}

func (a *arangoTaskLoader) filterArangoTasks(obj *api.ArangoTask, filters ...arangotask.Filter) bool {
	for _, f := range filters {
		if !f(obj) {
			return false
		}
	}

	return true
}

func (a *arangoTaskLoader) ArangoTasks() []*api.ArangoTask {
	var r []*api.ArangoTask
	for _, at := range a.at {
		r = append(r, at)
	}

	return r
}

func (a *arangoTaskLoader) ArangoTask(name string) (*api.ArangoTask, bool) {
	at, ok := a.at[name]
	if !ok {
		return nil, false
	}

	return at, true
}

func (a *arangoTaskLoader) IterateArangoTasks(action arangotask.Action, filters ...arangotask.Filter) error {
	for _, node := range a.ArangoTasks() {
		if err := a.iterateArangoTask(node, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (a *arangoTaskLoader) iterateArangoTask(at *api.ArangoTask, action arangotask.Action, filters ...arangotask.Filter) error {
	for _, filter := range filters {
		if !filter(at) {
			return nil
		}
	}

	return action(at)
}

func (a *arangoTaskLoader) ArangoTaskReadInterface() arangotask.ReadInterface {
	return &arangoTaskReadInterface{i: a}
}

type arangoTaskReadInterface struct {
	i *arangoTaskLoader
}

func (a *arangoTaskReadInterface) Get(ctx context.Context, name string, opts meta.GetOptions) (*api.ArangoTask, error) {
	if s, ok := a.i.ArangoTask(name); !ok {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    deployment.ArangoDeploymentGroupName,
			Resource: "arangotasks",
		}, name)
	} else {
		return s, nil
	}
}

func arangoTaskPointer(at api.ArangoTask) *api.ArangoTask {
	return &at
}

func arangoTasksToMap(ctx context.Context, inspector *inspector, k versioned.Interface, namespace string) func() error {
	return func() error {
		ats, err := getArangoTasks(ctx, k, namespace, "")
		if err != nil {
			if apiErrors.IsUnauthorized(err) || apiErrors.IsNotFound(err) {
				inspector.at = &arangoTaskLoader{
					accessible: false,
				}
				return nil
			}
			return err
		}

		atsMap := map[string]*api.ArangoTask{}

		for _, at := range ats {
			_, exists := atsMap[at.GetName()]
			if exists {
				return errors.Newf("ArangoMember %s already exists in map, error received", at.GetName())
			}

			atsMap[at.GetName()] = arangoTaskPointer(at)
		}

		inspector.at = &arangoTaskLoader{
			accessible: true,
			at:         atsMap,
		}

		return nil
	}
}

func getArangoTasks(ctx context.Context, k versioned.Interface, namespace, cont string) ([]api.ArangoTask, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	ats, err := k.DatabaseV1().ArangoTasks(namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if ats.Continue != "" {
		newATLoader, err := getArangoTasks(ctx, k, namespace, ats.Continue)
		if err != nil {
			return nil, err
		}

		return append(ats.Items, newATLoader...), nil
	}

	return ats.Items, nil
}
