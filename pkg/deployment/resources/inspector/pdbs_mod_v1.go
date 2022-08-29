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

	policyv1 "k8s.io/api/policy/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	policytypedv1 "k8s.io/client-go/kubernetes/typed/policy/v1"

	poddisruptionbudgetv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget/v1"
)

type podDisruptionBudgetsModV1 struct {
	i *inspectorState
}

func (p podDisruptionBudgetsMod) V1() poddisruptionbudgetv1.ModInterface {
	return podDisruptionBudgetsModV1(p)
}

func (p podDisruptionBudgetsModV1) client() policytypedv1.PodDisruptionBudgetInterface {
	return p.i.Client().Kubernetes().PolicyV1().PodDisruptionBudgets(p.i.Namespace())
}

func (p podDisruptionBudgetsModV1) Create(ctx context.Context, podDisruptionBudget *policyv1.PodDisruptionBudget,
	opts meta.CreateOptions) (*policyv1.PodDisruptionBudget, error) {
	if podDisruptionBudget, err := p.client().Create(ctx, podDisruptionBudget, opts); err != nil {
		return podDisruptionBudget, err
	} else {
		p.i.GetThrottles().PodDisruptionBudget().Invalidate()
		return podDisruptionBudget, err
	}
}

func (p podDisruptionBudgetsModV1) Update(ctx context.Context, podDisruptionBudget *policyv1.PodDisruptionBudget,
	opts meta.UpdateOptions) (*policyv1.PodDisruptionBudget, error) {
	if podDisruptionBudget, err := p.client().Update(ctx, podDisruptionBudget, opts); err != nil {
		return podDisruptionBudget, err
	} else {
		p.i.GetThrottles().PodDisruptionBudget().Invalidate()
		return podDisruptionBudget, err
	}
}

func (p podDisruptionBudgetsModV1) Patch(ctx context.Context, name string, pt types.PatchType, data []byte,
	opts meta.PatchOptions, subresources ...string) (result *policyv1.PodDisruptionBudget, err error) {
	if podDisruptionBudget, err := p.client().Patch(ctx, name, pt, data, opts, subresources...); err != nil {
		return podDisruptionBudget, err
	} else {
		p.i.GetThrottles().PodDisruptionBudget().Invalidate()
		return podDisruptionBudget, err
	}
}

func (p podDisruptionBudgetsModV1) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	if err := p.client().Delete(ctx, name, opts); err != nil {
		return err
	} else {
		p.i.GetThrottles().PodDisruptionBudget().Invalidate()
		return err
	}
}
