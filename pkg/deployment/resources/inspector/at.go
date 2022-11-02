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
	"time"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(arangoTasksInspectorLoaderObj)
}

var arangoTasksInspectorLoaderObj = arangoTasksInspectorLoader{}

type arangoTasksInspectorLoader struct {
}

func (p arangoTasksInspectorLoader) Component() definitions.Component {
	return definitions.ArangoTask
}

func (p arangoTasksInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoTasksInspector
	p.loadV1(ctx, i, &q)
	i.arangoTasks = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoTasksInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *arangoTasksInspector) {
	var z arangoTasksInspectorV1

	z.arangoTaskInspector = q

	z.arangoTasks, z.err = p.getV1ArangoTasks(ctx, i)

	q.v1 = &z
}

func (p arangoTasksInspectorLoader) getV1ArangoTasks(ctx context.Context, i *inspectorState) (map[string]*api.ArangoTask, error) {
	objs, err := p.getV1ArangoTasksList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*api.ArangoTask, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p arangoTasksInspectorLoader) getV1ArangoTasksList(ctx context.Context, i *inspectorState) ([]*api.ArangoTask, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().DatabaseV1().ArangoTasks(i.namespace).List(ctxChild, meta.ListOptions{
		Limit: globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
	})

	if err != nil {
		return nil, err
	}

	items := obj.Items
	cont := obj.Continue
	var s = int64(len(items))

	if z := obj.RemainingItemCount; z != nil {
		s += *z
	}

	ptrs := make([]*api.ArangoTask, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1ArangoTasksListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p arangoTasksInspectorLoader) getV1ArangoTasksListRequest(ctx context.Context, i *inspectorState, cont string) ([]api.ArangoTask, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().DatabaseV1().ArangoTasks(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p arangoTasksInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoTasksInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoTasks != nil {
		if !override {
			return
		}
	}

	to.arangoTasks = from.arangoTasks
	to.arangoTasks.state = to
}

func (p arangoTasksInspectorLoader) Name() string {
	return "arangoTasks"
}

type arangoTasksInspector struct {
	state *inspectorState

	last time.Time

	v1 *arangoTasksInspectorV1
}

func (p *arangoTasksInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoTasksInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoTasksInspectorLoaderObj)
}

func (p *arangoTasksInspector) Version() version.Version {
	return version.V1
}

func (p *arangoTasksInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoTask()
}

func (p *arangoTasksInspector) validate() error {
	if p == nil {
		return errors.Newf("ArangoTaskInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
