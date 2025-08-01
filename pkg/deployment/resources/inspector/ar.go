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

	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1beta1"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoroute"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(arangoRoutesInspectorLoaderObj)
}

var arangoRoutesInspectorLoaderObj = arangoRoutesInspectorLoader{}

type arangoRoutesInspectorLoader struct {
}

func (p arangoRoutesInspectorLoader) Component() definitions.Component {
	return definitions.ArangoRoute
}

func (p arangoRoutesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q arangoRoutesInspector

	q.v1beta1 = newInspectorVersion[*networkingApi.ArangoRouteList, *networkingApi.ArangoRoute](ctx,
		inspectorConstants.ArangoRouteGRv1Beta1(),
		inspectorConstants.ArangoRouteGKv1Beta1(),
		i.client.Arango().NetworkingV1beta1().ArangoRoutes(i.namespace),
		arangoroute.List())

	i.arangoRoutes = &q
	q.state = i
	q.last = time.Now()
}

func (p arangoRoutesInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p arangoRoutesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.arangoRoutes != nil {
		if !override {
			return
		}
	}

	to.arangoRoutes = from.arangoRoutes
	to.arangoRoutes.state = to
}

func (p arangoRoutesInspectorLoader) Name() string {
	return "arangoRoutes"
}

type arangoRoutesInspector struct {
	state *inspectorState

	last time.Time

	v1beta1 *inspectorVersion[*networkingApi.ArangoRoute]
}

func (p *arangoRoutesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *arangoRoutesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, arangoRoutesInspectorLoaderObj)
}

func (p *arangoRoutesInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1Beta1
}

func (p *arangoRoutesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ArangoRoute()
}

func (p *arangoRoutesInspector) validate() error {
	if p == nil {
		return errors.Errorf("ArangoRouteInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1beta1.validate()
}

func (p *arangoRoutesInspector) V1Beta1() (generic.Inspector[*networkingApi.ArangoRoute], error) {
	if p.v1beta1.err != nil {
		return nil, p.v1beta1.err
	}

	return p.v1beta1, nil
}
