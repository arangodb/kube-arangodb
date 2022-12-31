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

package resources

import (
	"context"
	"math/rand"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Pods []*core.Pod

func (p Pods) Filter(f func(pod *core.Pod) bool) Pods {
	var r = make(Pods, 0, len(p))

	for _, c := range p {
		if f(c) {
			r = append(r, c)
		}
	}

	return r
}

func (p Pods) FilterByScheduled() Pods {
	return p.Filter(func(pod *core.Pod) bool {
		return pod.Status.NominatedNodeName != "" || pod.Spec.NodeName != ""
	})
}

func (p Pods) FilterByPVCName(pvc string) Pods {
	return p.Filter(func(pod *core.Pod) bool {
		for _, v := range pod.Spec.Volumes {
			if p := v.PersistentVolumeClaim; p != nil {
				if p.ClaimName == pvc {
					return true
				}
			}
		}

		return false
	})
}

func (p Pods) PickAny() *core.Pod {
	if len(p) == 0 {
		return nil
	}

	rand.Shuffle(len(p), func(i, j int) {
		p[i], p[j] = p[j], p[i]
	})

	return p[0]
}

func ListPods(ctx context.Context, in typedCore.PodInterface) (Pods, error) {
	var pods Pods

	cont := ""

	for {
		nextPods, c, err := listPods(ctx, in, cont)
		if err != nil {
			return nil, err
		}

		pods = append(pods, nextPods...)

		if c == "" {
			return pods, nil
		}

		cont = c
	}
}

func listPods(ctx context.Context, in typedCore.PodInterface, next string) (Pods, string, error) {
	opts := meta.ListOptions{}

	opts.Continue = next

	pods, err := in.List(ctx, opts)
	if err != nil {
		return nil, "", err
	}

	podsPointers := make(Pods, len(pods.Items))

	for id := range pods.Items {
		podsPointers[id] = pods.Items[id].DeepCopy()
	}

	return podsPointers, pods.Continue, nil
}
