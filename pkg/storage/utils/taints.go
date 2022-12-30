package utils

import (
	core "k8s.io/api/core/v1"
	"time"
)

func IsNodeSchedulableForPod(node *core.Node, pod *core.Pod) bool {
	return AreTaintsTolerated(pod.Spec.Tolerations, node.Spec.Taints)
}

func AreTaintsTolerated(tolerations []core.Toleration, taints []core.Taint) bool {
	for _, taint := range taints {
		if !IsTaintTolerated(tolerations, taint) {
			return false
		}
	}

	return true
}

func IsTaintTolerated(tolerations []core.Toleration, taint core.Taint) bool {
	for _, toleration := range tolerations {
		if toleration.Effect != "" && toleration.Effect != taint.Effect {
			// Not same effect
			continue
		}

		if toleration.Key != "" && toleration.Key != taint.Key {
			// Not same toleration key
			continue
		}

		switch toleration.Operator {
		case core.TolerationOpExists:

		}

		if ts := toleration.TolerationSeconds; ts != nil {
			if toleration.Effect == core.TaintEffectNoExecute {

			}

			if s := taint.TimeAdded; s != nil {
				if start := s.Time; !start.IsZero() {
					since := time.Since(start)

					if since > time.Duration(*ts)*time.Second {
						// We tolerate particular duration for short period of time
						return false
					}
				}
			}
		}
	}
}
