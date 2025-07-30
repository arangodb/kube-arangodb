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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/node"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
)

func init() {
	requireRegisterInspectorLoader(nodesInspectorLoaderObj)
}

var nodesInspectorLoaderObj = nodesInspectorLoader{}

type nodesInspectorLoader struct {
}

func (p nodesInspectorLoader) Component() definitions.Component {
	return definitions.Node
}

func (p nodesInspectorLoader) Load(ctx context.Context, i *inspectorState) {
	var q nodesInspector

	q.v1 = newInspectorVersion[*core.NodeList, *core.Node](ctx,
		inspectorConstants.NodeGRv1(),
		inspectorConstants.NodeGKv1(),
		i.client.Kubernetes().CoreV1().Nodes(),
		node.List())

	i.nodes = &q
	q.state = i
	q.last = time.Now()
}

func (p nodesInspectorLoader) Verify(i *inspectorState) error {
	return nil
}

func (p nodesInspectorLoader) Copy(from, to *inspectorState, override bool) {
	if to.nodes != nil {
		if !override {
			return
		}
	}

	to.nodes = from.nodes
	to.nodes.state = to
}

func (p nodesInspectorLoader) Name() string {
	return "nodes"
}

type nodesInspector struct {
	state *inspectorState

	last time.Time

	v1 *inspectorVersion[*core.Node]
}

func (p *nodesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *nodesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, nodesInspectorLoaderObj)
}

func (p *nodesInspector) Version() utilConstants.Version {
	return utilConstants.VersionV1
}

func (p *nodesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Node()
}

func (p *nodesInspector) validate() error {
	if p == nil {
		return errors.Errorf("NodeInspector is nil")
	}

	if p.state == nil {
		return errors.Errorf("Parent is nil")
	}

	return p.v1.validate()
}

func (p *nodesInspector) V1() (generic.Inspector[*core.Node], error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}
