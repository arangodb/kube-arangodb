//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/node"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

func (i *inspector) GetNodes() (node.Inspector, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if i.nodes == nil {
		return nil, false
	}

	return i.nodes, i.nodes.authenticated
}

type nodeLoader struct {
	authenticated bool

	nodes map[string]*core.Node
}

func (n *nodeLoader) Node(name string) (*core.Node, bool) {
	node, ok := n.nodes[name]
	if !ok {
		return nil, false
	}

	return node, true
}

func (n *nodeLoader) Nodes() []*core.Node {
	var r []*core.Node
	for _, node := range n.nodes {
		r = append(r, node)
	}

	return r
}

func (n *nodeLoader) IterateNodes(action node.Action, filters ...node.Filter) error {
	for _, node := range n.Nodes() {
		if err := n.iteratePodDisruptionBudget(node, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (n *nodeLoader) iteratePodDisruptionBudget(node *core.Node, action node.Action, filters ...node.Filter) error {
	for _, filter := range filters {
		if !filter(node) {
			return nil
		}
	}

	return action(node)
}

func (n *nodeLoader) NodeReadInterface() node.ReadInterface {
	return &nodeReadInterface{i: n}
}

type nodeReadInterface struct {
	i *nodeLoader
}

func (s nodeReadInterface) Get(ctx context.Context, name string, opts meta.GetOptions) (*core.Node, error) {
	if s, ok := s.i.Node(name); !ok {
		return nil, apiErrors.NewNotFound(schema.GroupResource{
			Group:    policy.GroupName,
			Resource: "nodes",
		}, name)
	} else {
		return s, nil
	}
}

func nodePointer(pod core.Node) *core.Node {
	return &pod
}

func nodesToMap(ctx context.Context, inspector *inspector, k kubernetes.Interface) func() error {
	return func() error {
		nodes, err := getNodes(ctx, k, "")
		if err != nil {
			if apiErrors.IsUnauthorized(err) {
				inspector.nodes = &nodeLoader{
					authenticated: false,
				}
				return nil
			}
			return err
		}

		nodesMap := map[string]*core.Node{}

		for _, node := range nodes {
			_, exists := nodesMap[node.GetName()]
			if exists {
				return errors.Newf("ArangoMember %s already exists in map, error received", node.GetName())
			}

			nodesMap[node.GetName()] = nodePointer(node)
		}

		inspector.nodes = &nodeLoader{
			authenticated: true,
			nodes:         nodesMap,
		}

		return nil
	}
}

func getNodes(ctx context.Context, k kubernetes.Interface, cont string) ([]core.Node, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	defer cancel()
	nodes, err := k.CoreV1().Nodes().List(ctxChild, meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if nodes.Continue != "" {
		nextNodeLayer, err := getNodes(ctx, k, nodes.Continue)
		if err != nil {
			return nil, err
		}

		return append(nodes.Items, nextNodeLayer...), nil
	}

	return nodes.Items, nil
}
