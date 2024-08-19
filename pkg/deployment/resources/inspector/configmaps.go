//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	requireRegisterInspectorLoader(configMapsInspectorLoaderObj)
}

var configMapsInspectorLoaderObj = configMapsInspectorLoader{}

type configMapsInspectorLoader struct {
}

func (p configMapsInspectorLoader) Component() definitions.Component {
	return definitions.ConfigMap
}

func (p configMapsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q configMapsInspector
	p.loadV1(ctx, i, &q)
	i.configMaps = &q
	q.state = i
	q.last = time.Now()
}

func (p configMapsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *configMapsInspector) {
	var z configMapsInspectorV1

	z.configMapInspector = q

	z.configMaps, z.err = p.getV1ConfigMaps(ctx, i)

	q.v1 = &z
}

func (p configMapsInspectorLoader) getV1ConfigMaps(ctx context.Context, i *inspectorState) (map[string]*core.ConfigMap, error) {
	objs, err := p.getV1ConfigMapsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.ConfigMap, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p configMapsInspectorLoader) getV1ConfigMapsList(ctx context.Context, i *inspectorState) ([]*core.ConfigMap, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().ConfigMaps(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*core.ConfigMap, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1ConfigMapsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p configMapsInspectorLoader) getV1ConfigMapsListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.ConfigMap, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().ConfigMaps(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p configMapsInspectorLoader) Verify(i *inspectorState) error {
	if err := i.configMaps.v1.err; err != nil {
		return err
	}

	return nil
}

func (p configMapsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.configMaps != nil {
		if !override {
			return
		}
	}

	to.configMaps = from.configMaps
	to.configMaps.state = to
}

func (p configMapsInspectorLoader) Name() string {
	return "configMaps"
}

type configMapsInspector struct {
	state *inspectorState

	last time.Time

	v1 *configMapsInspectorV1
}

func (p *configMapsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *configMapsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, configMapsInspectorLoaderObj)
}

func (p *configMapsInspector) Version() version.Version {
	return version.V1
}

func (p *configMapsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ConfigMap()
}

func (p *configMapsInspector) validate() error {
	if p == nil {
		return errors.Errorf("ConfigMapInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}
