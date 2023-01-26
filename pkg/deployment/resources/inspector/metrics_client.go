//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
)

func withMetrics[S meta.Object](definition definitions.Component, in generic.ClientStatusGetter[S]) generic.ClientStatusGetter[S] {
	return func() generic.ModStatusClient[S] {
		return statusClientMetrics[S]{
			component: definition,
			in:        in(),
		}
	}
}

type statusClientMetrics[S meta.Object] struct {
	component definitions.Component
	in        generic.ModStatusClient[S]
}

func (s statusClientMetrics[S]) Create(ctx context.Context, obj S, opts meta.CreateOptions) (S, error) {
	r, err := s.in.Create(ctx, obj, opts)
	clientMetricsInstance.ObjectRequest(s.component, definitions.Create, obj, err)
	return r, err
}

func (s statusClientMetrics[S]) Update(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error) {
	r, err := s.in.Update(ctx, obj, opts)
	clientMetricsInstance.ObjectRequest(s.component, definitions.Update, obj, err)
	return r, err
}

func (s statusClientMetrics[S]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result S, err error) {
	r, err := s.in.Patch(ctx, name, pt, data, opts, subresources...)
	clientMetricsInstance.Request(s.component, definitions.Patch, name, err)
	return r, err
}

func (s statusClientMetrics[S]) Delete(ctx context.Context, name string, opts meta.DeleteOptions) error {
	verb := definitions.Delete
	if g := opts.GracePeriodSeconds; g != nil && *g == 0 {
		verb = definitions.ForceDelete
	}

	err := s.in.Delete(ctx, name, opts)
	clientMetricsInstance.Request(s.component, verb, name, err)
	return err
}

func (s statusClientMetrics[S]) UpdateStatus(ctx context.Context, obj S, opts meta.UpdateOptions) (S, error) {
	r, err := s.in.UpdateStatus(ctx, obj, opts)
	clientMetricsInstance.ObjectRequest(s.component, definitions.UpdateStatus, obj, err)
	return r, err
}
