//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/configmap"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
)

func init() {
	requireRegisterInspectorLoader(configMapsInspectorLoaderObj)
}

var configMapsInspectorLoaderObj = configMapsInspectorLoader{}

type configMapsInspectorLoader struct {
}

func (p configMapsInspectorLoader) Component() definitions.Component {
	return definitions.ConfigMap
}

func (p configMapsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q configMapsInspector

	q.v1 = newInspectorVersion[*core.ConfigMapList, *core.ConfigMap](ctx,
		constants.ConfigMapGRv1(),
		constants.ConfigMapGKv1(),
		i.client.Kubernetes().CoreV1().ConfigMaps(i.namespace),
		configmap.List())

	i.configMaps = &q
	q.state = i
	q.last = time.Now()
}

func (p configMapsInspectorLoader) Verify(i *inspectorState) error {
	if err := i.configMaps.v1.err; err != nil {
		return err
	}

	return nil
}

func (p configMapsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.configMaps != nil {
		if !override {
			return
		}
	}

	to.configMaps = from.configMaps
	to.configMaps.state = to
}

func (p configMapsInspectorLoader) Name() string {
	return "configMaps"
}

type configMapsInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*core.ConfigMap]
}

func (p *configMapsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *configMapsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, configMapsInspectorLoaderObj)
}

func (p *configMapsInspector) Version() version.Version {
	return version.V1
}

func (p *configMapsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.ConfigMap()
}

func (p *configMapsInspector) validate() error {
	if p == nil {
		return errors.Errorf("ConfigMapInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *configMapsInspector) V1() generic.Inspector[*core.ConfigMap] {
	return p.v1
}
