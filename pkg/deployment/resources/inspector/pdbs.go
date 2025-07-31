//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget"
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

	if i.versionInfo.CompareTo("1.21") >= 1 {
		q.v1 = newInspectorVersion[*policy.PodDisruptionBudgetList, *policy.PodDisruptionBudget](ctx,
			inspectorConstants.PodDisruptionBudgetGRv1(),
			inspectorConstants.PodDisruptionBudgetGKv1(),
			i.client.Kubernetes().PolicyV1().PodDisruptionBudgets(i.namespace),
			poddisruptionbudget.List())
	} else {
		q.v1 = &inspectorVersion[*policy.PodDisruptionBudget]{
			err: newMinK8SVersion("1.20"),
		}
	}

	i.podDisruptionBudgets = &q
	q.state = i
	q.last = time.Now()
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

	v1 *inspectorVersion[*policy.PodDisruptionBudget]
}

func (p *podDisruptionBudgetsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *podDisruptionBudgetsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, podDisruptionBudgetsInspectorLoaderObj)
}

func (p *podDisruptionBudgetsInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *podDisruptionBudgetsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.PodDisruptionBudget()
}

func (p *podDisruptionBudgetsInspector) validate() error {
	if p == nil {
		return errors.Errorf("PodDisruptionBudgetInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	if err := p.v1.validate(); err != nil {
		if _, ok := IsK8SVersion(err); !ok {
			return err
		}
	}

	return nil
}

func (p *podDisruptionBudgetsInspector) V1() (generic.Inspector[*policy.PodDisruptionBudget], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
