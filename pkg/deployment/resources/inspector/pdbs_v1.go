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

	policy "k8s.io/api/policy/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget/v1"
)

func (p *podDisruptionBudgetsInspector) V1() (ins.Inspector, error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}

type podDisruptionBudgetsInspectorV1 struct {
	podDisruptionBudgetInspector *podDisruptionBudgetsInspector

	podDisruptionBudgets map[string]*policy.PodDisruptionBudget
	err                  error
}

func (p *podDisruptionBudgetsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("PodDisruptionBudgetsV1Inspector is nil")
	}

	if p.podDisruptionBudgetInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.podDisruptionBudgets == nil && p.err == nil {
		return errors.Newf("PodDisruptionBudgets or err should be not nil")
	}

	if p.podDisruptionBudgets != nil && p.err != nil {
		return errors.Newf("PodDisruptionBudgets or err cannot be not nil together")
	}

	return nil
}

func (p *podDisruptionBudgetsInspectorV1) PodDisruptionBudgets() []*policy.PodDisruptionBudget {
	var r []*policy.PodDisruptionBudget
	for _, podDisruptionBudget := range p.podDisruptionBudgets {
		r = append(r, podDisruptionBudget)
	}

	return r
}

func (p *podDisruptionBudgetsInspectorV1) GetSimple(name string) (*policy.PodDisruptionBudget, bool) {
	podDisruptionBudget, ok := p.podDisruptionBudgets[name]
	if !ok {
		return nil, false
	}

	return podDisruptionBudget, true
}

func (p *podDisruptionBudgetsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, podDisruptionBudget := range p.podDisruptionBudgets {
		if err := p.iteratePodDisruptionBudget(podDisruptionBudget, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *podDisruptionBudgetsInspectorV1) iteratePodDisruptionBudget(podDisruptionBudget *policy.PodDisruptionBudget, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(podDisruptionBudget) {
			return nil
		}
	}

	return action(podDisruptionBudget)
}

func (p *podDisruptionBudgetsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *podDisruptionBudgetsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*policy.PodDisruptionBudget, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.PodDisruptionBudgetGR(), name)
	} else {
		return s, nil
	}
}
