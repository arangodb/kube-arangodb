//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"
	"sort"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/storage/utils"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Nodes []*core.Node

func (p Nodes) Filter(f func(node *core.Node) bool) Nodes {
	var r = make(Nodes, 0, len(p))

	for _, c := range p {
		if f(c) {
			r = append(r, c)
		}
	}

	return r
}

func (p Nodes) Copy() Nodes {
	c := make(Nodes, len(p))
	copy(c, p)
	return c
}

func (p Nodes) Sort(f func(a, b *core.Node) bool) Nodes {
	z := p.Copy()
	sort.Slice(z, func(i, j int) bool {
		return f(z[i], z[j])
	})
	return z
}

func (p Nodes) SortBySchedulablePodsTaints(pods Pods) Nodes {
	return p.Sort(func(a, b *core.Node) bool {
		return utils.IsNodeSchedulableForPods(a, pods...) > utils.IsNodeSchedulableForPods(b, pods...)
	})
}

func (p Nodes) SortBySchedulablePodTaints(pod *core.Pod) Nodes {
	return p.Sort(func(a, b *core.Node) bool {
		return utils.IsNodeSchedulableForPod(a, pod) > utils.IsNodeSchedulableForPod(b, pod)
	})
}

func (p Nodes) FilterPodsTaints(pods Pods) Nodes {
	return p.Filter(func(node *core.Node) bool {
		return utils.IsNodeSchedulableForPods(node, pods...).Schedulable()
	})
}

func (p Nodes) FilterTaints(pod *core.Pod) Nodes {
	return p.Filter(func(node *core.Node) bool {
		return utils.IsNodeSchedulableForPod(node, pod).Schedulable()
	})
}

func (p Nodes) FilterSchedulable() Nodes {
	return p.Filter(func(node *core.Node) bool {
		return !node.Spec.Unschedulable
	})
}

func (p Nodes) PickAny() *core.Node {
	if len(p) == 0 {
		return nil
	}

	util.Rand().Shuffle(len(p), func(i, j int) {
		p[i], p[j] = p[j], p[i]
	})

	return p[0]
}

func ListNodes(ctx context.Context, in typedCore.NodeInterface) (Nodes, error) {
	var nodes Nodes

	cont := ""

	for {
		nextNodes, c, err := listNodes(ctx, in, cont)
		if err != nil {
			return nil, err
		}

		nodes = append(nodes, nextNodes...)

		if c == "" {
			return nodes, nil
		}

		cont = c
	}
}

func listNodes(ctx context.Context, in typedCore.NodeInterface, next string) (Nodes, string, error) {
	opts := meta.ListOptions{}

	opts.Continue = next

	nodes, err := in.List(ctx, opts)
	if err != nil {
		return nil, "", err
	}

	nodesPointers := make(Nodes, len(nodes.Items))

	for id := range nodes.Items {
		nodesPointers[id] = nodes.Items[id].DeepCopy()
	}

	return nodesPointers, nodes.Continue, nil
}
