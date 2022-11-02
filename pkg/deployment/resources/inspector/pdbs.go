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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
	"time"

	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
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

	if i.versionInfo.CompareTo("1.21") >= 0 {
		p.loadV1(ctx, i, &q)

		q.v1beta1 = &podDisruptionBudgetsInspectorV1Beta1{
			podDisruptionBudgetInspector: &q,
			err: apiErrors.NewNotFound(schema.GroupResource{
				Group:    policyv1beta1.GroupName,
				Resource: "podDisruptionBudgets",
			}, ""),
		}
	} else {
		p.loadV1Beta1(ctx, i, &q)

		q.v1 = &podDisruptionBudgetsInspectorV1{
			podDisruptionBudgetInspector: &q,
			err: apiErrors.NewNotFound(schema.GroupResource{
				Group:    policyv1.GroupName,
				Resource: "podDisruptionBudgets",
			}, ""),
		}
	}
	i.podDisruptionBudgets = &q
	q.state = i
	q.last = time.Now()
}

func (p podDisruptionBudgetsInspectorLoader) loadV1Beta1(ctx context.Context, i *inspectorState, q *podDisruptionBudgetsInspector) {
	var z podDisruptionBudgetsInspectorV1Beta1

	z.podDisruptionBudgetInspector = q

	z.podDisruptionBudgets, z.err = p.getV1Beta1PodDisruptionBudgets(ctx, i)

	q.v1beta1 = &z
}

func (p podDisruptionBudgetsInspectorLoader) getV1Beta1PodDisruptionBudgets(ctx context.Context, i *inspectorState) (map[string]*policyv1beta1.PodDisruptionBudget, error) {
	objs, err := p.getV1Beta1PodDisruptionBudgetsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*policyv1beta1.PodDisruptionBudget, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p podDisruptionBudgetsInspectorLoader) getV1Beta1PodDisruptionBudgetsList(ctx context.Context, i *inspectorState) ([]*policyv1beta1.PodDisruptionBudget, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().PolicyV1beta1().PodDisruptionBudgets(i.namespace).List(ctxChild, meta.ListOptions{
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

	ptrs := make([]*policyv1beta1.PodDisruptionBudget, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1Beta1PodDisruptionBudgetsListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p podDisruptionBudgetsInspectorLoader) getV1Beta1PodDisruptionBudgetsListRequest(ctx context.Context, i *inspectorState, cont string) ([]policyv1beta1.PodDisruptionBudget, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().PolicyV1beta1().PodDisruptionBudgets(i.namespace).List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
}

func (p podDisruptionBudgetsInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *podDisruptionBudgetsInspector) {
	var z podDisruptionBudgetsInspectorV1

	z.podDisruptionBudgetInspector = q

	z.podDisruptionBudgets, z.err = p.getV1PodDisruptionBudgets(ctx, i)

	q.v1 = &z
}

func (p podDisruptionBudgetsInspectorLoader) getV1PodDisruptionBudgets(ctx context.Context, i *inspectorState) (map[string]*policyv1.PodDisruptionBudget, error) {
	objs, err := p.getV1PodDisruptionBudgetsList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*policyv1.PodDisruptionBudget, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p podDisruptionBudgetsInspectorLoader) getV1PodDisruptionBudgetsList(ctx context.Context, i *inspectorState) ([]*policyv1.PodDisruptionBudget, error) {
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

	ptrs := make([]*policyv1.PodDisruptionBudget, 0, s)

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

func (p podDisruptionBudgetsInspectorLoader) getV1PodDisruptionBudgetsListRequest(ctx context.Context, i *inspectorState, cont string) ([]policyv1.PodDisruptionBudget, string, error) {
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
	if errv1, errv1beta1 := i.podDisruptionBudgets.v1.err, i.podDisruptionBudgets.v1beta1.err; errv1 != nil && errv1beta1 != nil {
		return errors.Wrap(errv1, "Both requests failed")
	} else if errv1 == nil && errv1beta1 == nil {
		return errors.Newf("V1 and V1beta1 are not nil - only one should be picked")
	}

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

	v1      *podDisruptionBudgetsInspectorV1
	v1beta1 *podDisruptionBudgetsInspectorV1Beta1
}

func (p *podDisruptionBudgetsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *podDisruptionBudgetsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, podDisruptionBudgetsInspectorLoaderObj)
}

func (p *podDisruptionBudgetsInspector) Version() version.Version {
	if p.state.versionInfo.CompareTo("1.21") >= 0 {
		return version.V1
	}

	return version.V1Beta1
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
		return err
	}

	if err := p.v1beta1.validate(); err != nil {
		return err
	}

	return nil
}
