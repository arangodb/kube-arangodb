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
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoclustersynchronization"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(arangoClusterSynchronizationsInspectorLoaderObj)
}

var arangoClusterSynchronizationsInspectorLoaderObj = arangoClusterSynchronizationsInspectorLoader{}

type arangoClusterSynchronizationsInspectorLoader struct {
}

func (p arangoClusterSynchronizationsInspectorLoader) Component() definitions.Component {
	return definitions.ArangoClusterSynchronization
}

func (p arangoClusterSynchronizationsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoClusterSynchronizationsInspector

	q.v1 = newInspectorVersion[*api.ArangoClusterSynchronizationList, *api.ArangoClusterSynchronization](ctx,
		constants.ArangoClusterSynchronizationGRv1(),
		constants.ArangoClusterSynchronizationGKv1(),
		i.client.Arango().DatabaseV1().ArangoClusterSynchronizations(i.namespace),
		arangoclustersynchronization.List())

	i.arangoClusterSynchronizations = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoClusterSynchronizationsInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoClusterSynchronizationsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoClusterSynchronizations != nil {
		if !override {
			return
		}
	}

	to.arangoClusterSynchronizations = from.arangoClusterSynchronizations
	to.arangoClusterSynchronizations.state = to
}

func (p arangoClusterSynchronizationsInspectorLoader) Name() string {
	return "arangoClusterSynchronizations"
}

type arangoClusterSynchronizationsInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*api.ArangoClusterSynchronization]
}

func (p *arangoClusterSynchronizationsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoClusterSynchronizationsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoClusterSynchronizationsInspectorLoaderObj)
}

func (p *arangoClusterSynchronizationsInspector) Version() version.Version {
	return version.V1
}

func (p *arangoClusterSynchronizationsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoClusterSynchronization()
}

func (p *arangoClusterSynchronizationsInspector) validate() error {
	if p == nil {
		return errors.Errorf("ArangoClusterSynchronizationInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *arangoClusterSynchronizationsInspector) V1() (generic.Inspector[*api.ArangoClusterSynchronization], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
