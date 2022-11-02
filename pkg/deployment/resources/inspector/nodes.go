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
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/definitions"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/version"
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
	p.loadV1(ctx, i, &q)
	i.nodes = &q
	q.state = i
	q.last = time.Now()
}

func (p nodesInspectorLoader) loadV1(ctx context.Context, i *inspectorState, q *nodesInspector) {
	var z nodesInspectorV1

	z.nodeInspector = q

	z.nodes, z.err = p.getV1Nodes(ctx, i)

	q.v1 = &z
}

func (p nodesInspectorLoader) getV1Nodes(ctx context.Context, i *inspectorState) (map[string]*core.Node, error) {
	objs, err := p.getV1NodesList(ctx, i)
	if err != nil {
		return nil, err
	}

	r := make(map[string]*core.Node, len(objs))

	for id := range objs {
		r[objs[id].GetName()] = objs[id]
	}

	return r, nil
}

func (p nodesInspectorLoader) getV1NodesList(ctx context.Context, i *inspectorState) ([]*core.Node, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Nodes().List(ctxChild, meta.ListOptions{
		Limit: globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
	})

	if err != nil {
		return nil, err
	}

	items := obj.Items
	cont := obj.Continue
	var s = int64(len(items))

	if z := obj.RemainingItemCount; z != nil {
		s += *z
	}

	ptrs := make([]*core.Node, 0, s)

	for {
		for id := range items {
			ptrs = append(ptrs, &items[id])
		}

		if cont == "" {
			break
		}

		items, cont, err = p.getV1NodesListRequest(ctx, i, cont)

		if err != nil {
			return nil, err
		}
	}

	return ptrs, nil
}

func (p nodesInspectorLoader) getV1NodesListRequest(ctx context.Context, i *inspectorState, cont string) ([]core.Node, string, error) {
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	obj, err := i.client.Kubernetes().CoreV1().Nodes().List(ctxChild, meta.ListOptions{
		Limit:    globals.GetGlobals().Kubernetes().RequestBatchSize().Get(),
		Continue: cont,
	})

	if err != nil {
		return nil, "", err
	}

	return obj.Items, obj.Continue, err
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

	v1 *nodesInspectorV1
}

func (p *nodesInspector) LastRefresh() time.Time {
	return p.last
}

func (p *nodesInspector) Refresh(ctx context.Context) error {
	p.Throttle(p.state.throttles).Invalidate()
	return p.state.refresh(ctx, nodesInspectorLoaderObj)
}

func (p *nodesInspector) Version() version.Version {
	return version.V1
}

func (p *nodesInspector) Throttle(c throttle.Components) throttle.Throttle {
	return c.Node()
}

func (p *nodesInspector) validate() error {
	if p == nil {
		return errors.Newf("NodeInspector is nil")
	}

	if p.state == nil {
		return errors.Newf("Parent is nil")
	}

	return p.v1.validate()
}
