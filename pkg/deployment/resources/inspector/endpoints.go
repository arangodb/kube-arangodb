//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

	discovery "k8s.io/api/discovery/v1"

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpointslices"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(endpointSlicesInspectorLoaderObj)
}

var endpointSlicesInspectorLoaderObj = endpointSlicesInspectorLoader{}

type endpointSlicesInspectorLoader struct {
}

func (p endpointSlicesInspectorLoader) Component() definitions.Component {
	return definitions.EndpointSlices
}

func (p endpointSlicesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q endpointSlicesInspector

	q.v1 = newInspectorVersion[*discovery.EndpointSliceList, *discovery.EndpointSlice](ctx,
		inspectorConstants.EndpointSlicesGRv1(),
		inspectorConstants.EndpointSlicesGKv1(),
		i.client.Kubernetes().DiscoveryV1().EndpointSlices(i.namespace),
		endpointslices.List())

	i.endpointSlices = &q
	q.state = i
	q.last = time.Now()
}

func (p endpointSlicesInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p endpointSlicesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.endpointSlices != nil {
		if !override {
			return
		}
	}

	to.endpointSlices = from.endpointSlices
	to.endpointSlices.state = to
}

func (p endpointSlicesInspectorLoader) Name() string {
	return "endpointSlices"
}

type endpointSlicesInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*discovery.EndpointSlice]
}

func (p *endpointSlicesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *endpointSlicesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, endpointSlicesInspectorLoaderObj)
}

func (p *endpointSlicesInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *endpointSlicesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.EndpointSlices()
}

func (p *endpointSlicesInspector) validate() error {
	if p == nil {
		return errors.Errorf("endpointSlicesInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *endpointSlicesInspector) V1() (generic.Inspector[*discovery.EndpointSlice], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
