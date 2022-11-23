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

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpoints/v1"
)

func (p *endpointsInspector) V1() (ins.Inspector, error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}

type endpointsInspectorV1 struct {
	endpointsInspector *endpointsInspector

	endpoints map[string]*core.Endpoints
	err       error
}

func (p *endpointsInspectorV1) Filter(filters ...ins.Filter) []*core.Endpoints {
	z := p.ListSimple()

	r := make([]*core.Endpoints, 0, len(z))

	for _, o := range z {
		if !ins.FilterObject(o, filters...) {
			continue
		}

		r = append(r, o)
	}

	return r
}

func (p *endpointsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("EndpointsV1Inspector is nil")
	}

	if p.endpointsInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.endpoints == nil && p.err == nil {
		return errors.Newf("Endpoints or err should be not nil")
	}

	if p.endpoints != nil && p.err != nil {
		return errors.Newf("Endpoints or err cannot be not nil together")
	}

	return nil
}

func (p *endpointsInspectorV1) ListSimple() []*core.Endpoints {
	var r []*core.Endpoints
	for _, endpoints := range p.endpoints {
		r = append(r, endpoints)
	}

	return r
}

func (p *endpointsInspectorV1) GetSimple(name string) (*core.Endpoints, bool) {
	endpoints, ok := p.endpoints[name]
	if !ok {
		return nil, false
	}

	return endpoints, true
}

func (p *endpointsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, endpoints := range p.endpoints {
		if err := p.iterateEndpoints(endpoints, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *endpointsInspectorV1) iterateEndpoints(endpoints *core.Endpoints, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(endpoints) {
			return nil
		}
	}

	return action(endpoints)
}

func (p *endpointsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *endpointsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Endpoints, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.EndpointsGR(), name)
	} else {
		return s, nil
	}
}
