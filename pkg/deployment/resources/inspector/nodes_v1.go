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

	core "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	ins "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/node/v1"
)

func (p *nodesInspector) V1() (ins.Inspector, error) {
	if p.v1.err != nil {
		return nil, p.v1.err
	}

	return p.v1, nil
}

type nodesInspectorV1 struct {
	nodeInspector *nodesInspector

	nodes map[string]*core.Node
	err   error
}

func (p *nodesInspectorV1) validate() error {
	if p == nil {
		return errors.Newf("NodesV1Inspector is nil")
	}

	if p.nodeInspector == nil {
		return errors.Newf("Parent is nil")
	}

	if p.nodes == nil && p.err == nil {
		return errors.Newf("Nodes or err should be not nil")
	}

	if p.nodes != nil && p.err != nil {
		return errors.Newf("Nodes or err cannot be not nil together")
	}

	return nil
}

func (p *nodesInspectorV1) ListSimple() []*core.Node {
	var r []*core.Node
	for _, node := range p.nodes {
		r = append(r, node)
	}

	return r
}

func (p *nodesInspectorV1) GetSimple(name string) (*core.Node, bool) {
	node, ok := p.nodes[name]
	if !ok {
		return nil, false
	}

	return node, true
}

func (p *nodesInspectorV1) Iterate(action ins.Action, filters ...ins.Filter) error {
	for _, node := range p.nodes {
		if err := p.iterateNode(node, action, filters...); err != nil {
			return err
		}
	}

	return nil
}

func (p *nodesInspectorV1) iterateNode(node *core.Node, action ins.Action, filters ...ins.Filter) error {
	for _, f := range filters {
		if f == nil {
			continue
		}

		if !f(node) {
			return nil
		}
	}

	return action(node)
}

func (p *nodesInspectorV1) Read() ins.ReadInterface {
	return p
}

func (p *nodesInspectorV1) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Node, error) {
	if s, ok := p.GetSimple(name); !ok {
		return nil, apiErrors.NewNotFound(constants.NodeGR(), name)
	} else {
		return s, nil
	}
}
