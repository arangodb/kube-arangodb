package scheduling

import core "k8s.io/api/core/v1"

func CopyAffinity(dst, src *core.Affinity) {
	if src.PodAffinity != nil && dst.PodAffinity == nil {
		dst.PodAffinity = &core.PodAffinity{}
	}
	MergePodAffinity(dst.PodAffinity, src.PodAffinity)

	if src.PodAntiAffinity != nil && dst.PodAntiAffinity == nil {
		dst.PodAntiAffinity = &core.PodAntiAffinity{}
	}
	MergePodAntiAffinity(dst.PodAntiAffinity, src.PodAntiAffinity)

	if src.NodeAffinity != nil && dst.NodeAffinity == nil {
		dst.NodeAffinity = &core.NodeAffinity{}
	}
	MergeNodeAffinity(dst.NodeAffinity, src.NodeAffinity)
}

func MergePodAntiAffinity(a, b *core.PodAntiAffinity) {
	if a == nil || b == nil {
		return
	}

	a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution,
		b.PreferredDuringSchedulingIgnoredDuringExecution...)

	a.RequiredDuringSchedulingIgnoredDuringExecution = append(a.RequiredDuringSchedulingIgnoredDuringExecution,
		b.RequiredDuringSchedulingIgnoredDuringExecution...)
}

func MergePodAffinity(a, b *core.PodAffinity) {
	if a == nil || b == nil {
		return
	}

	a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution,
		b.PreferredDuringSchedulingIgnoredDuringExecution...)

	a.RequiredDuringSchedulingIgnoredDuringExecution = append(a.RequiredDuringSchedulingIgnoredDuringExecution,
		b.RequiredDuringSchedulingIgnoredDuringExecution...)
}

func MergeNodeAffinity(a, b *core.NodeAffinity) {
	if a == nil || b == nil {
		return
	}

	a.PreferredDuringSchedulingIgnoredDuringExecution = append(a.PreferredDuringSchedulingIgnoredDuringExecution,
		b.PreferredDuringSchedulingIgnoredDuringExecution...)

	var newSelectorTerms []core.NodeSelectorTerm

	if b.RequiredDuringSchedulingIgnoredDuringExecution == nil || len(b.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		newSelectorTerms = a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	} else if a.RequiredDuringSchedulingIgnoredDuringExecution == nil || len(a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		newSelectorTerms = b.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
	} else {
		for _, aTerms := range a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			for _, bTerms := range b.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				term := aTerms.DeepCopy()
				if len(bTerms.MatchExpressions) != 0 {
					term.MatchExpressions = append(term.MatchExpressions, bTerms.MatchExpressions...)
				}
				if len(bTerms.MatchFields) != 0 {
					term.MatchFields = append(term.MatchFields, bTerms.MatchFields...)
				}
				newSelectorTerms = append(newSelectorTerms, *term)
			}
		}
	}

	a.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = newSelectorTerms
}
