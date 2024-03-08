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

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

var _ interfaces.Container[Probes] = &Probes{}

type Probes struct {
	// LivenessProbe keeps configuration of periodic probe of container liveness.
	// Container will be restarted if the probe fails.
	// +doc/type: core.Probe
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	LivenessProbe *core.Probe `json:"livenessProbe,omitempty"`
	// ReadinessProbe keeps configuration of periodic probe of container service readiness.
	// Container will be removed from service endpoints if the probe fails.
	// +doc/type: core.Probe
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	ReadinessProbe *core.Probe `json:"readinessProbe,omitempty"`
	// StartupProbe indicates that the Pod has successfully initialized.
	// If specified, no other probes are executed until this completes successfully.
	// If this probe fails, the Pod will be restarted, just as if the livenessProbe failed.
	// This can be used to provide different probe parameters at the beginning of a Pod's lifecycle,
	// when it might take a long time to load data or warm a cache, than during steady-state operation.
	// +doc/type: core.Probe
	// +doc/link: Kubernetes docs|https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	StartupProbe *core.Probe `json:"startupProbe,omitempty"`
}

func (n *Probes) Apply(_ *core.PodTemplateSpec, template *core.Container) error {
	if n == nil {
		return nil
	}

	template.LivenessProbe = n.LivenessProbe.DeepCopy()
	template.StartupProbe = n.StartupProbe.DeepCopy()
	template.ReadinessProbe = n.ReadinessProbe.DeepCopy()

	return nil
}

func (n *Probes) With(newResources *Probes) *Probes {
	if n == nil && newResources == nil {
		return nil
	}

	if n == nil {
		return newResources.DeepCopy()
	}

	if newResources == nil {
		return n.DeepCopy()
	}

	return &Probes{
		LivenessProbe:  kresources.MergeProbes(n.LivenessProbe, newResources.LivenessProbe),
		ReadinessProbe: kresources.MergeProbes(n.ReadinessProbe, newResources.ReadinessProbe),
		StartupProbe:   kresources.MergeProbes(n.StartupProbe, newResources.StartupProbe),
	}
}

func (n *Probes) Validate() error {
	return nil
}
