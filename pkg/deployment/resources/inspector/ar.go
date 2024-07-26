//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(arangoRoutesInspectorLoaderObj)
}

var arangoRoutesInspectorLoaderObj = arangoRoutesInspectorLoader{}

type arangoRoutesInspectorLoader struct {
}

func (p arangoRoutesInspectorLoader) Component() definitions.Component {
	return definitions.ArangoRoute
}

func (p arangoRoutesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoRoutesInspector
	p.loadV1Alpha1(ctx, i, &q)
	i.arangoRoutes = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoRoutesInspectorLoader) loadV1Alpha1(ctx context.Context, i *inspectorState, q *arangoRoutesInspector) {
	var z arangoRoutesInspectorV1Alpha1

	z.arangoRouteInspector = q

	z.arangoRoutes, z.err = p.getV1ArangoRoutes(ctx, i)

	q.v1alpha1 = &z
}

func (p arangoRoutesInspectorLoader) getV1ArangoRoutes(ctx context.Context, i *inspectorState) (map[string]*networkingApi.ArangoRoute, error) {
	objs, err := p.getV1ArangoRoutesList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*networkingApi.ArangoRoute, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p arangoRoutesInspectorLoader) getV1ArangoRoutesList(ctx context.Context, i *inspectorState) ([]*networkingApi.ArangoRoute, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().NetworkingV1alpha1().ArangoRoutes(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*networkingApi.ArangoRoute, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1ArangoRoutesListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p arangoRoutesInspectorLoader) getV1ArangoRoutesListRequest(ctx context.Context, i *inspectorState, cont string) ([]networkingApi.ArangoRoute, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Arango().NetworkingV1alpha1().ArangoRoutes(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p arangoRoutesInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoRoutesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoRoutes != nil {
		if !override {
			return
		}
	}

	to.arangoRoutes = from.arangoRoutes
	to.arangoRoutes.state = to
}

func (p arangoRoutesInspectorLoader) Name() string {
	return "arangoRoutes"
}

type arangoRoutesInspector struct {
	state *inspectorState

	last time.Time

	v1alpha1 *arangoRoutesInspectorV1Alpha1
}

func (p *arangoRoutesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoRoutesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoRoutesInspectorLoaderObj)
}

func (p *arangoRoutesInspector) Version() version.Version {
	return version.V1
}

func (p *arangoRoutesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoRoute()
}

func (p *arangoRoutesInspector) validate() error {
	if p == nil {
		return errors.Errorf("ArangoRouteInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1alpha1.validate()
}
