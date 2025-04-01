//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
)

func List[L generic.ListContinue, S meta.Object](ctx context.Context, i generic.ListInterface[L], call generic.ExtractorList[L, S]) (map[string]S, error) {
	return list.APIMap[L, S](ctx, i, meta.ListOptions{}, call)
}

func newInspectorVersion[L generic.ListContinue, S meta.Object](ctx context.Context,
	gvr schema.GroupVersionResource,
	gvk schema.GroupVersionKind,
	i generic.ListInterface[L],
	call generic.ExtractorList[L, S]) *inspectorVersion[S] {
	var r inspectorVersion[S]

	r.gvr = gvr
	r.gvk = gvk

	r.items, r.err = List(ctx, i, call)

	return &r
}

type inspectorVersion[S meta.Object] struct {
	items map[string]S
	gvr   schema.GroupVersionResource
	gvk   schema.GroupVersionKind
	err   error
}

func (p *inspectorVersion[S]) GroupVersionKind() schema.GroupVersionKind {
	return p.gvk
}

func (p *inspectorVersion[S]) GroupVersionResource() schema.GroupVersionResource {
	return p.gvr
}

func (p *inspectorVersion[S]) Filter(filters ...generic.Filter[S]) []S {
	z := p.ListSimple()

	r := make([]S, 0, len(z))

	for _, o := range z {
		if !generic.FilterObject(o, filters...) {
			continue
		}

		r = append(r, o)
	}

	return r
}

func (p *inspectorVersion[S]) validate() error {
	if p == nil {
		return errors.Errorf("Inspector is nil")
	}

	if p.items == nil && p.err == nil {
		return errors.Errorf("Items or err should be not nil")
	}

	if p.items != nil && p.err != nil {
		return errors.Errorf("Items or err cannot be not nil together")
	}

	return nil
}

func (p *inspectorVersion[S]) ListSimple() []S {
	var r []S
	for _, item := range p.items {
		r = append(r, item)
	}

	return r
}

func (p *inspectorVersion[S]) GetSimple(name string) (S, bool) {
	item, ok := p.items[name]
	if !ok {
		return util.Default[S](), false
	}

	return item, true
}

func (p *inspectorVersion[S]) Iterate(action generic.Action[S], filters ...generic.Filter[S]) error {
	for _, item := range p.items {
		if err := p.iterateArangoProfile(item, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *inspectorVersion[S]) iterateArangoProfile(item S, action generic.Action[S], filters ...generic.Filter[S]) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(item) {
			return nil
		}
	}

	return action(item)
}

func (p *inspectorVersion[S]) Read() generic.ReadClient[S] {
	return p
}

func (p *inspectorVersion[S]) Get(ctx context.Context, name string, opts meta.GetOptions) (S, error) {
	if s, ok := p.GetSimple(name); !ok {
		return util.Default[S](), apiErrors.NewNotFound(p.gvr.GroupResource(), name)
	} else {
		return s, nil
	}
}
