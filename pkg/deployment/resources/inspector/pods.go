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
	requireRegisterInspectorLoader(podsInspectorLoaderObj)
}

var podsInspectorLoaderObj = podsInspectorLoader{}

type podsInspectorLoader struct {
}

func (p podsInspectorLoader) Component() definitions.Component {
	return definitions.Pod
}

func (p podsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q podsInspector
	p.loadV1(ctx, i, &q)
	i.pods = &q
	q.state = i
	q.last = time.Now()
}

func (p podsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *podsInspector) {
	var z podsInspectorV1

	z.podInspector = q

	z.pods, z.err = p.getV1Pods(ctx, i)

	q.v1 = &z
}

func (p podsInspectorLoader) getV1Pods(ctx context.Context, i *inspectorState) (map[string]*core.Pod, error) {
	objs, err := p.getV1PodsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.Pod, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p podsInspectorLoader) getV1PodsList(ctx context.Context, i *inspectorState) ([]*core.Pod, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Pods(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*core.Pod, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1PodsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p podsInspectorLoader) getV1PodsListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.Pod, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Pods(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p podsInspectorLoader) Verify(i *inspectorState) error {
	if err := i.pods.v1.err; err != nil {
		return err
	}

	return nil
}

func (p podsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.pods != nil {
		if !override {
			return
		}
	}

	to.pods = from.pods
	to.pods.state = to
}

func (p podsInspectorLoader) Name() string {
	return "pods"
}

type podsInspector struct {
	state *inspectorState

	last time.Time

	v1 *podsInspectorV1
}

func (p *podsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *podsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, podsInspectorLoaderObj)
}

func (p *podsInspector) Version() version.Version {
	return version.V1
}

func (p *podsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Pod()
}

func (p *podsInspector) validate() error {
	if p == nil {
		return errors.Newf("PodInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
