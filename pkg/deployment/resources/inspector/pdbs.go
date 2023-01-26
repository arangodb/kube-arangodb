//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	policy "k8s.io/api/policy/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(podDisruptionBudgetsInspectorLoaderObj)
}

var podDisruptionBudgetsInspectorLoaderObj = podDisruptionBudgetsInspectorLoader{}

type podDisruptionBudgetsInspectorLoader struct {
}

func (p podDisruptionBudgetsInspectorLoader) Component() definitions.Component {
	return definitions.PodDisruptionBudget
}

func (p podDisruptionBudgetsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q podDisruptionBudgetsInspector

	if i.versionInfo.CompareTo("1.21") >= 1 {
		p.loadV1(ctx, i, &q)
	} else {
		q.v1 = &podDisruptionBudgetsInspectorV1{
			podDisruptionBudgetInspector: &q,
			err:                          newMinK8SVersion("1.20"),
		}
	}

	i.podDisruptionBudgets = &q
	q.state = i
	q.last = time.Now()
}

func (p podDisruptionBudgetsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *podDisruptionBudgetsInspector) {
	var z podDisruptionBudgetsInspectorV1

	z.podDisruptionBudgetInspector = q

	z.podDisruptionBudgets, z.err = p.getV1PodDisruptionBudgets(ctx, i)

	q.v1 = &z
}

func (p podDisruptionBudgetsInspectorLoader) getV1PodDisruptionBudgets(ctx context.Context, i *inspectorState) (map[string]*policy.PodDisruptionBudget, error) {
	objs, err := p.getV1PodDisruptionBudgetsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*policy.PodDisruptionBudget, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p podDisruptionBudgetsInspectorLoader) getV1PodDisruptionBudgetsList(ctx context.Context, i *inspectorState) ([]*policy.PodDisruptionBudget, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().PolicyV1().PodDisruptionBudgets(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*policy.PodDisruptionBudget, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1PodDisruptionBudgetsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p podDisruptionBudgetsInspectorLoader) getV1PodDisruptionBudgetsListRequest(ctx context.Context, i *inspectorState, cont string) ([]policy.PodDisruptionBudget, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().PolicyV1().PodDisruptionBudgets(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p podDisruptionBudgetsInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p podDisruptionBudgetsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.podDisruptionBudgets != nil {
		if !override {
			return
		}
	}

	to.podDisruptionBudgets = from.podDisruptionBudgets
	to.podDisruptionBudgets.state = to
}

func (p podDisruptionBudgetsInspectorLoader) Name() string {
	return "podDisruptionBudgets"
}

type podDisruptionBudgetsInspector struct {
	state *inspectorState

	last time.Time

	v1 *podDisruptionBudgetsInspectorV1
}

func (p *podDisruptionBudgetsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *podDisruptionBudgetsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, podDisruptionBudgetsInspectorLoaderObj)
}

func (p *podDisruptionBudgetsInspector) Version() version.Version {
	return version.V1
}

func (p *podDisruptionBudgetsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.PodDisruptionBudget()
}

func (p *podDisruptionBudgetsInspector) validate() error {
	if p == nil {
		return errors.Newf("PodDisruptionBudgetInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	if err := p.v1.validate(); err != nil {
		if _, ok := IsK8SVersion(err); !ok {
			return err
		}
	}

	return nil
}
