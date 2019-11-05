package v1

import (
	v1 "k8s.io/api/core/v1"
)

type LifecycleSpec struct {
	Resources v1.ResourceRequirements `json:"resources,omitempty"`
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *LifecycleSpec) SetDefaultsFrom(source LifecycleSpec) {
	setDefaultsFromResourceList(&s.Resources.Limits, source.Resources.Limits)
	setDefaultsFromResourceList(&s.Resources.Requests, source.Resources.Requests)
}
