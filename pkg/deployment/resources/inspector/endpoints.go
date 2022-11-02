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

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(endpointsInspectorLoaderObj)
}

var endpointsInspectorLoaderObj = endpointsInspectorLoader{}

type endpointsInspectorLoader struct {
}

func (p endpointsInspectorLoader) Component() definitions.Component {
	return definitions.Endpoints
}

func (p endpointsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q endpointsInspector
	p.loadV1(ctx, i, &q)
	i.endpoints = &q
	q.state = i
	q.last = time.Now()
}

func (p endpointsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *endpointsInspector) {
	var z endpointsInspectorV1

	z.endpointsInspector = q

	z.endpoints, z.err = p.getV1Endpoints(ctx, i)

	q.v1 = &z
}

func (p endpointsInspectorLoader) getV1Endpoints(ctx context.Context, i *inspectorState) (map[string]*core.Endpoints, error) {
	objs, err := p.getV1EndpointsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.Endpoints, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p endpointsInspectorLoader) getV1EndpointsList(ctx context.Context, i *inspectorState) ([]*core.Endpoints, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Endpoints(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*core.Endpoints, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1EndpointsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p endpointsInspectorLoader) getV1EndpointsListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.Endpoints, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Endpoints(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p endpointsInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p endpointsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.endpoints != nil {
		if !override {
			return
		}
	}

	to.endpoints = from.endpoints
	to.endpoints.state = to
}

func (p endpointsInspectorLoader) Name() string {
	return "endpoints"
}

type endpointsInspector struct {
	state *inspectorState

	last time.Time

	v1 *endpointsInspectorV1
}

func (p *endpointsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *endpointsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, endpointsInspectorLoaderObj)
}

func (p *endpointsInspector) Version() version.Version {
	return version.V1
}

func (p *endpointsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Endpoints()
}

func (p *endpointsInspector) validate() error {
	if p == nil {
		return errors.Newf("EndpointsInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
