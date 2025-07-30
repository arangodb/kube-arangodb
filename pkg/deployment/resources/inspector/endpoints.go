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

	core "k8s.io/api/core/v1"

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpoints"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(endpointsInspectorLoaderObj)
}

var endpointsInspectorLoaderObj = endpointsInspectorLoader{}

type endpointsInspectorLoader struct {
}

func (p endpointsInspectorLoader) Component() definitions.Component {
	return definitions.Endpoints
}

func (p endpointsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q endpointsInspector

	q.v1 = newInspectorVersion[*core.EndpointsList, *core.Endpoints](ctx,
		inspectorConstants.EndpointsGRv1(),
		inspectorConstants.EndpointsGKv1(),
		i.client.Kubernetes().CoreV1().Endpoints(i.namespace),
		endpoints.List())

	i.endpoints = &q
	q.state = i
	q.last = time.Now()
}

func (p endpointsInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p endpointsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.endpoints != nil {
		if !override {
			return
		}
	}

	to.endpoints = from.endpoints
	to.endpoints.state = to
}

func (p endpointsInspectorLoader) Name() string {
	return "endpoints"
}

type endpointsInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*core.Endpoints]
}

func (p *endpointsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *endpointsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, endpointsInspectorLoaderObj)
}

func (p *endpointsInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *endpointsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Endpoints()
}

func (p *endpointsInspector) validate() error {
	if p == nil {
		return errors.Errorf("EndpointsInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *endpointsInspector) V1() (generic.Inspector[*core.Endpoints], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
