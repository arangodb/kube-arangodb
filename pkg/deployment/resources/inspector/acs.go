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
	requireRegisterInspectorLoader(arangoClusterSynchronizationsInspectorLoaderObj)
}

var arangoClusterSynchronizationsInspectorLoaderObj = arangoClusterSynchronizationsInspectorLoader{}

type arangoClusterSynchronizationsInspectorLoader struct {
}

func (p arangoClusterSynchronizationsInspectorLoader) Component() definitions.Component {
	return definitions.ArangoClusterSynchronization
}

func (p arangoClusterSynchronizationsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoClusterSynchronizationsInspector
	p.loadV1(ctx, i, &q)
	i.arangoClusterSynchronizations = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoClusterSynchronizationsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *arangoClusterSynchronizationsInspector) {
	var z arangoClusterSynchronizationsInspectorV1

	z.arangoClusterSynchronizationInspector = q

	z.arangoClusterSynchronizations, z.err = p.getV1ArangoClusterSynchronizations(ctx, i)

	q.v1 = &z
}

func (p arangoClusterSynchronizationsInspectorLoader) getV1ArangoClusterSynchronizations(ctx context.Context, i *inspectorState) (map[string]*api.ArangoClusterSynchronization, error) {
	objs, err := p.getV1ArangoClusterSynchronizationsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*api.ArangoClusterSynchronization, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p arangoClusterSynchronizationsInspectorLoader) getV1ArangoClusterSynchronizationsList(ctx context.Context, i *inspectorState) ([]*api.ArangoClusterSynchronization, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().DatabaseV1().ArangoClusterSynchronizations(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*api.ArangoClusterSynchronization, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1ArangoClusterSynchronizationsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p arangoClusterSynchronizationsInspectorLoader) getV1ArangoClusterSynchronizationsListRequest(ctx context.Context, i *inspectorState, cont string) ([]api.ArangoClusterSynchronization, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().DatabaseV1().ArangoClusterSynchronizations(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p arangoClusterSynchronizationsInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoClusterSynchronizationsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoClusterSynchronizations != nil {
		if !override {
			return
		}
	}

	to.arangoClusterSynchronizations = from.arangoClusterSynchronizations
	to.arangoClusterSynchronizations.state = to
}

func (p arangoClusterSynchronizationsInspectorLoader) Name() string {
	return "arangoClusterSynchronizations"
}

type arangoClusterSynchronizationsInspector struct {
	state *inspectorState

	last time.Time

	v1 *arangoClusterSynchronizationsInspectorV1
}

func (p *arangoClusterSynchronizationsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoClusterSynchronizationsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoClusterSynchronizationsInspectorLoaderObj)
}

func (p *arangoClusterSynchronizationsInspector) Version() version.Version {
	return version.V1
}

func (p *arangoClusterSynchronizationsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoClusterSynchronization()
}

func (p *arangoClusterSynchronizationsInspector) validate() error {
	if p == nil {
		return errors.Newf("ArangoClusterSynchronizationInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
