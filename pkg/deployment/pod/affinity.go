//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package pod

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AppendPodAntiAffinityDefault(p k8sutil.PodCreator, a *core.PodAntiAffinity) {
	labels := k8sutil.LabelsForDeployment(p.GetName(), p.GetRole())
	labelSelector := &meta.LabelSelector{
		MatchLabels: labels,
	}

	if !p.IsDeploymentMode() {
		a.RequiredDuringSchedulingIgnoredDuringExecution = append(a.RequiredDuringSchedulingIgnoredDuringExecution, core.PodAffinityTerm{
			LabelSelector: labelSelector,
			TopologyKey:   k8sutil.TopologyKeyHostname,
		})
	} else {
		a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution, core.WeightedPodAffinityTerm{
			Weight: 1,
			PodAffinityTerm: core.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   k8sutil.TopologyKeyHostname,
			},
		})
	}
}

func AppendNodeSelector(a *core.NodeAffinity) {
	if a.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		a.RequiredDuringSchedulingIgnoredDuringExecution = &core.NodeSelector{}
	}

	a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, core.NodeSelectorTerm{
		MatchExpressions: []core.NodeSelectorRequirement{
			{
				Key:      "beta.kubernetes.io/arch",
				Operator: "In",
				Values:   []string{"amd64"},
			},
		},
	})
}

func AppendAffinityWithRole(p k8sutil.PodCreator, a *core.PodAffinity, role string) {
	labelSelector := &meta.LabelSelector{
		MatchLabels: k8sutil.LabelsForDeployment(p.GetName(), role),
	}
	if !p.IsDeploymentMode() {
		a.RequiredDuringSchedulingIgnoredDuringExecution = append(a.RequiredDuringSchedulingIgnoredDuringExecution, core.PodAffinityTerm{
			LabelSelector: labelSelector,
			TopologyKey:   k8sutil.TopologyKeyHostname,
		})
	} else {
		a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution, core.WeightedPodAffinityTerm{
			Weight: 1,
			PodAffinityTerm: core.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   k8sutil.TopologyKeyHostname,
			},
		})
	}
}

func MergePodAntiAffinity(a ,b *core.PodAntiAffinity) {
	if a == nil || b == nil {
		return
	}

	for _, rule := range b.PreferredDuringSchedulingIgnoredDuringExecution {
		a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution, rule)
	}

	for _, rule := range b.RequiredDuringSchedulingIgnoredDuringExecution {
		a.RequiredDuringSchedulingIgnoredDuringExecution = append(a.RequiredDuringSchedulingIgnoredDuringExecution, rule)
	}
}

func ReturnPodAffinityOrNil(a core.PodAffinity) *core.PodAffinity{
	if len(a.RequiredDuringSchedulingIgnoredDuringExecution) > 0 || len(a.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		return &a
	}

	return nil
}

func ReturnPodAntiAffinityOrNil(a core.PodAntiAffinity) *core.PodAntiAffinity {
	if len(a.RequiredDuringSchedulingIgnoredDuringExecution) > 0 || len(a.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		return &a
	}

	return nil
}

func ReturnNodeAffinityOrNil( a core.NodeAffinity) *core.NodeAffinity {
	if len(a.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		return &a
	}

	if s := a.RequiredDuringSchedulingIgnoredDuringExecution; s != nil {
		if len(s.NodeSelectorTerms) > 0 {
			return &a
		}
	}

	return nil
}