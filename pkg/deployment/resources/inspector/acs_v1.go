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

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoclustersynchronization/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

func (p *arangoClusterSynchronizationsInspector) V1() (ins.Inspector, error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}

type arangoClusterSynchronizationsInspectorV1 struct {
	arangoClusterSynchronizationInspector *arangoClusterSynchronizationsInspector

	arangoClusterSynchronizations map[string]*api.ArangoClusterSynchronization
	err                           error
}

func (p *arangoClusterSynchronizationsInspectorV1) Filter(filters ...ins.Filter) []*api.ArangoClusterSynchronization {
	z := p.ListSimple()

	r := make([]*api.ArangoClusterSynchronization, 0, len(z))

	for _, o := range z {
		if !ins.FilterObject(o, filters...) {
			continue
		}

		r = append(r, o)
	}

	return r
}

func (p *arangoClusterSynchronizationsInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("ArangoClusterSynchronizationsV1Inspector is nil")
	}

	if p.arangoClusterSynchronizationInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.arangoClusterSynchronizations == nil && p.err == nil {
		return errors.Newf("ListSimple or err should be not nil")
	}

	if p.arangoClusterSynchronizations != nil && p.err != nil {
		return errors.Newf("ListSimple or err cannot be not nil together")
	}

	return nil
}

func (p *arangoClusterSynchronizationsInspectorV1) ListSimple() []*api.ArangoClusterSynchronization {
	var r []*api.ArangoClusterSynchronization
	for _, arangoClusterSynchronization := range p.arangoClusterSynchronizations {
		r = append(r, arangoClusterSynchronization)
	}

	return r
}

func (p *arangoClusterSynchronizationsInspectorV1) GetSimple(name string) (*api.ArangoClusterSynchronization, bool) {
	arangoClusterSynchronization, ok := p.arangoClusterSynchronizations[name]
	if !ok {
		return nil, false
	}

	return arangoClusterSynchronization, true
}

func (p *arangoClusterSynchronizationsInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, arangoClusterSynchronization := range p.arangoClusterSynchronizations {
		if err := p.iterateArangoClusterSynchronization(arangoClusterSynchronization, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *arangoClusterSynchronizationsInspectorV1) iterateArangoClusterSynchronization(arangoClusterSynchronization *api.ArangoClusterSynchronization, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(arangoClusterSynchronization) {
			return nil
		}
	}

	return action(arangoClusterSynchronization)
}

func (p *arangoClusterSynchronizationsInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *arangoClusterSynchronizationsInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*api.ArangoClusterSynchronization, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.ArangoClusterSynchronizationGR(), name)
	} else {
		return s, nil
	}
}
