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

package pod

import (
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

func AppendPodAntiAffinityDefault(p interfaces.PodCreator, a *core.PodAntiAffinity) {
	labels := k8sutil.LabelsForDeployment(p.GetName(), p.GetRole())
	labelSelector := &meta.LabelSelector{
		MatchLabels: labels,
	}

	if !p.IsDeploymentMode() {
		a.RequiredDuringSchedulingIgnoredDuringExecution = append(a.RequiredDuringSchedulingIgnoredDuringExecution, core.PodAffinityTerm{
			LabelSelector: labelSelector,
			TopologyKey:   shared.TopologyKeyHostname,
		})
	} else {
		a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution, core.WeightedPodAffinityTerm{
			Weight: 1,
			PodAffinityTerm: core.PodAffinityTerm{
				LabelSelector: labelSelector,
				TopologyKey:   shared.TopologyKeyHostname,
			},
		})
	}
}

func AppendArchSelector(a *core.NodeAffinity, nodeSelectorForArch core.NodeSelectorTerm) {
	if a.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		a.RequiredDuringSchedulingIgnoredDuringExecution = &core.NodeSelector{}
	}

	a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = append(a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, nodeSelectorForArch)
}

func AppendAffinityWithRole(p interfaces.PodCreator, a *core.PodAffinity, role string) {
	labelSelector := &meta.LabelSelector{
		MatchLabels: k8sutil.LabelsForDeployment(p.GetName(), role),
	}
	a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution, core.WeightedPodAffinityTerm{
		Weight: 1,
		PodAffinityTerm: core.PodAffinityTerm{
			LabelSelector: labelSelector,
			TopologyKey:   shared.TopologyKeyHostname,
		},
	})
}

func ReturnPodAffinityOrNil(a core.PodAffinity) *core.PodAffinity {
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

func ReturnNodeAffinityOrNil(a core.NodeAffinity) *core.NodeAffinity {
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
