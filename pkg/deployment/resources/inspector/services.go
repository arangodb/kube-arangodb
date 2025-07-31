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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(servicesInspectorLoaderObj)
}

var servicesInspectorLoaderObj = servicesInspectorLoader{}

type servicesInspectorLoader struct {
}

func (p servicesInspectorLoader) Component() definitions.Component {
	return definitions.Service
}

func (p servicesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q servicesInspector

	q.v1 = newInspectorVersion[*core.ServiceList, *core.Service](ctx,
		inspectorConstants.ServiceGRv1(),
		inspectorConstants.ServiceGKv1(),
		i.client.Kubernetes().CoreV1().Services(i.namespace),
		service.List())

	i.services = &q
	q.state = i
	q.last = time.Now()
}

func (p servicesInspectorLoader) Verify(i *inspectorState) error {
	if err := i.services.v1.err; err != nil {
		return err
	}

	return nil
}

func (p servicesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.services != nil {
		if !override {
			return
		}
	}

	to.services = from.services
	to.services.state = to
}

func (p servicesInspectorLoader) Name() string {
	return "services"
}

type servicesInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*core.Service]
}

func (p *servicesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *servicesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, servicesInspectorLoaderObj)
}

func (p *servicesInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *servicesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Service()
}

func (p *servicesInspector) validate() error {
	if p == nil {
		return errors.Errorf("ServiceInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *servicesInspector) V1() generic.Inspector[*core.Service] {
	return p.v1
}
