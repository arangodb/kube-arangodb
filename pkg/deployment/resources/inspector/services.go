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
	requireRegisterInspectorLoader(servicesInspectorLoaderObj)
}

var servicesInspectorLoaderObj = servicesInspectorLoader{}

type servicesInspectorLoader struct {
}

func (p servicesInspectorLoader) Component() definitions.Component {
	return definitions.Service
}

func (p servicesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q servicesInspector
	p.loadV1(ctx, i, &q)
	i.services = &q
	q.state = i
	q.last = time.Now()
}

func (p servicesInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *servicesInspector) {
	var z servicesInspectorV1

	z.serviceInspector = q

	z.services, z.err = p.getV1Services(ctx, i)

	q.v1 = &z
}

func (p servicesInspectorLoader) getV1Services(ctx context.Context, i *inspectorState) (map[string]*core.Service, error) {
	objs, err := p.getV1ServicesList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.Service, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p servicesInspectorLoader) getV1ServicesList(ctx context.Context, i *inspectorState) ([]*core.Service, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Services(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*core.Service, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1ServicesListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p servicesInspectorLoader) getV1ServicesListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.Service, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Services(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p servicesInspectorLoader) Verify(i *inspectorState) error {
	if err := i.services.v1.err; err != nil {
		return err
	}

	return nil
}

func (p servicesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.services != nil {
		if !override {
			return
		}
	}

	to.services = from.services
	to.services.state = to
}

func (p servicesInspectorLoader) Name() string {
	return "services"
}

type servicesInspector struct {
	state *inspectorState

	last time.Time

	v1 *servicesInspectorV1
}

func (p *servicesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *servicesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, servicesInspectorLoaderObj)
}

func (p *servicesInspector) Version() version.Version {
	return version.V1
}

func (p *servicesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Service()
}

func (p *servicesInspector) validate() error {
	if p == nil {
		return errors.Newf("ServiceInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
