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

package resources

import core "k8s.io/api/core/v1"

func Merge(a, b *core.Affinity) *core.Affinity {
	if a == nil && b == nil {
		return nil
	}

	if a == nil {
		return b.DeepCopy()
	}

	if b == nil {
		return a.DeepCopy()
	}

	return Optional(&core.Affinity{
		PodAntiAffinity: OptionalPodAntiAffinity(MergePodAntiAffinity(a.PodAntiAffinity, b.PodAntiAffinity)),
		PodAffinity:     OptionalPodAffinity(MergePodAffinity(a.PodAffinity, b.PodAffinity)),
		NodeAffinity:    OptionalNodeAffinity(MergeNodeAffinity(a.NodeAffinity, b.NodeAffinity)),
	})
}

func Optional(a *core.Affinity) *core.Affinity {
	if a.PodAntiAffinity == nil && a.NodeAffinity == nil && a.PodAffinity == nil {
		return nil
	}

	return a
}

func MergePodAffinity(a, b *core.PodAffinity) *core.PodAffinity {
	if a == nil && b == nil {
		return nil
	}

	if a == nil {
		return b.DeepCopy()
	}

	if b == nil {
		return a.DeepCopy()
	}

	n := a.DeepCopy()

	n.PreferredDuringSchedulingIgnoredDuringExecution = append(n.PreferredDuringSchedulingIgnoredDuringExecution,
		b.PreferredDuringSchedulingIgnoredDuringExecution...)

	n.RequiredDuringSchedulingIgnoredDuringExecution = append(n.RequiredDuringSchedulingIgnoredDuringExecution,
		b.RequiredDuringSchedulingIgnoredDuringExecution...)

	return n
}

func OptionalPodAffinity(a *core.PodAffinity) *core.PodAffinity {
	if a == nil {
		return nil
	}

	if len(a.RequiredDuringSchedulingIgnoredDuringExecution) > 0 || len(a.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		return a
	}

	return nil
}

func MergePodAntiAffinity(a, b *core.PodAntiAffinity) *core.PodAntiAffinity {
	if a == nil && b == nil {
		return nil
	}

	if a == nil {
		return b.DeepCopy()
	}

	if b == nil {
		return a.DeepCopy()
	}

	n := a.DeepCopy()

	n.PreferredDuringSchedulingIgnoredDuringExecution = append(n.PreferredDuringSchedulingIgnoredDuringExecution,
		b.PreferredDuringSchedulingIgnoredDuringExecution...)

	n.RequiredDuringSchedulingIgnoredDuringExecution = append(n.RequiredDuringSchedulingIgnoredDuringExecution,
		b.RequiredDuringSchedulingIgnoredDuringExecution...)

	return n
}

func OptionalPodAntiAffinity(a *core.PodAntiAffinity) *core.PodAntiAffinity {
	if a == nil {
		return nil
	}

	if len(a.RequiredDuringSchedulingIgnoredDuringExecution) > 0 || len(a.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		return a
	}

	return nil
}

func MergeNodeAffinity(a, b *core.NodeAffinity) *core.NodeAffinity {
	if a == nil && b == nil {
		return nil
	}

	if a == nil {
		return b.DeepCopy()
	}

	if b == nil {
		return a.DeepCopy()
	}

	n := a.DeepCopy()

	n.PreferredDuringSchedulingIgnoredDuringExecution = append(n.PreferredDuringSchedulingIgnoredDuringExecution,
		b.PreferredDuringSchedulingIgnoredDuringExecution...)

	n.RequiredDuringSchedulingIgnoredDuringExecution = MergeNodeSelector(n.RequiredDuringSchedulingIgnoredDuringExecution, b.RequiredDuringSchedulingIgnoredDuringExecution)

	return n
}

func MergeNodeSelector(a, b *core.NodeSelector) *core.NodeSelector {
	if a == nil && b == nil {
		return nil
	}

	if a == nil {
		return b.DeepCopy()
	}

	if b == nil {
		return a.DeepCopy()
	}

	if len(a.NodeSelectorTerms) == 0 && len(b.NodeSelectorTerms) == 0 {
		return nil
	}

	if len(a.NodeSelectorTerms) == 0 {
		return b.DeepCopy()
	}

	if len(b.NodeSelectorTerms) == 0 {
		return a.DeepCopy()
	}

	var terms []core.NodeSelectorTerm

	for _, ida := range a.NodeSelectorTerms {
		for _, idb := range b.NodeSelectorTerms {
			var sa, sb core.NodeSelectorTerm

			ida.DeepCopyInto(&sa)
			idb.DeepCopyInto(&sb)

			sa.MatchExpressions = append(sa.MatchExpressions, sb.MatchExpressions...)
			sa.MatchFields = append(sa.MatchFields, sb.MatchFields...)

			if len(sa.MatchExpressions) == 0 && len(sa.MatchFields) == 0 {
				continue
			}

			terms = append(terms, sa)
		}
	}

	if len(terms) == 0 {
		return nil
	}

	return &core.NodeSelector{NodeSelectorTerms: terms}
}

func OptionalNodeAffinity(a *core.NodeAffinity) *core.NodeAffinity {
	if a == nil {
		return nil
	}

	if len(a.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		return a
	}

	if s := a.RequiredDuringSchedulingIgnoredDuringExecution; s != nil {
		if len(s.NodeSelectorTerms) > 0 {
			return a
		}
	}

	return nil
}
