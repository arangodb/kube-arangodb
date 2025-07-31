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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(podsInspectorLoaderObj)
}

var podsInspectorLoaderObj = podsInspectorLoader{}

type podsInspectorLoader struct {
}

func (p podsInspectorLoader) Component() definitions.Component {
	return definitions.Pod
}

func (p podsInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q podsInspector

	q.v1 = newInspectorVersion[*core.PodList, *core.Pod](ctx,
		inspectorConstants.PodGRv1(),
		inspectorConstants.PodGKv1(),
		i.client.Kubernetes().CoreV1().Pods(i.namespace),
		pod.List())

	i.pods = &q
	q.state = i
	q.last = time.Now()
}

func (p podsInspectorLoader) Verify(i *inspectorState) error {
	if err := i.pods.v1.err; err != nil {
		return err
	}

	return nil
}

func (p podsInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.pods != nil {
		if !override {
			return
		}
	}

	to.pods = from.pods
	to.pods.state = to
}

func (p podsInspectorLoader) Name() string {
	return "pods"
}

type podsInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*core.Pod]
}

func (p *podsInspector) LastRefresh() time.Time {
	return p.last
}

func (p *podsInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, podsInspectorLoaderObj)
}

func (p *podsInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *podsInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Pod()
}

func (p *podsInspector) validate() error {
	if p == nil {
		return errors.Errorf("PodInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *podsInspector) V1() generic.Inspector[*core.Pod] {

	return p.v1
}
