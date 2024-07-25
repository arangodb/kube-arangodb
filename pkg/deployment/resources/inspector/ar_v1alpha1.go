//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoroute/v1alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *arangoRoutesInspector) V1Alpha1() (ins.Inspector, error) {
	if p.v1alpha1.err != nil {
		return nil, p.v1alpha1.err
	}

	return p.v1alpha1, nil
}

type arangoRoutesInspectorV1Alpha1 struct {
	arangoRouteInspector *arangoRoutesInspector

	arangoRoutes map[string]*networkingApi.ArangoRoute
	err          error
}

func (p *arangoRoutesInspectorV1Alpha1) Filter(filters ...ins.Filter) []*networkingApi.ArangoRoute {
	z := p.ListSimple()

	r := make([]*networkingApi.ArangoRoute, 0, len(z))

	for _, o := range z {
		if !ins.FilterObject(o, filters...) {
			continue
		}

		r = append(r, o)
	}

	return r
}

func (p *arangoRoutesInspectorV1Alpha1) validate() error {
	if p == nil {
		return errors.Errorf("ArangoRoutesV1AlphaInspector is nil")
	}

	if p.arangoRouteInspector == nil {
		return errors.Errorf("Parent is nil")
	}

	if p.arangoRoutes == nil && p.err == nil {
		return errors.Errorf("ArangoRoutes or err should be not nil")
	}

	if p.arangoRoutes != nil && p.err != nil {
		return errors.Errorf("ArangoRoutes or err cannot be not nil together")
	}

	return nil
}

func (p *arangoRoutesInspectorV1Alpha1) ListSimple() []*networkingApi.ArangoRoute {
	var r []*networkingApi.ArangoRoute
	for _, arangoRoute := range p.arangoRoutes {
		r = append(r, arangoRoute)
	}

	return r
}

func (p *arangoRoutesInspectorV1Alpha1) GetSimple(name string) (*networkingApi.ArangoRoute, bool) {
	arangoRoute, ok := p.arangoRoutes[name]
	if !ok {
		return nil, false
	}

	return arangoRoute, true
}

func (p *arangoRoutesInspectorV1Alpha1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, arangoRoute := range p.arangoRoutes {
		if err := p.iterateArangoRoute(arangoRoute, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *arangoRoutesInspectorV1Alpha1) iterateArangoRoute(arangoRoute *networkingApi.ArangoRoute, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(arangoRoute) {
			return nil
		}
	}

	return action(arangoRoute)
}

func (p *arangoRoutesInspectorV1Alpha1) Read() ins.ReadInterface {
	return p
}

func (p *arangoRoutesInspectorV1Alpha1) Get(ctx context.Context, name string, opts meta.GetOptions) (*networkingApi.ArangoRoute, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ArangoRouteGR(), name)
	} else {
		return s, nil
	}
}
