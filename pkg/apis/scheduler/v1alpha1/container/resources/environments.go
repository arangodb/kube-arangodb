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

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/envs"
)

type Environments struct {
	// Env keeps the information about environment variables provided to the container
	// +doc/type: core.EnvVar
	// +doc/link: Kubernetes Docs|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envvar-v1-core
	Env []core.EnvVar `json:"env,omitempty"`

	// EnvFrom keeps the information about environment variable sources provided to the container
	// +doc/type: core.EnvFromSource
	// +doc/link: Kubernetes Docs|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envfromsource-v1-core
	EnvFrom []core.EnvFromSource `json:"envFrom,omitempty"`
}

func (e *Environments) Apply(container *core.Container) error {
	if e == nil {
		return nil
	}

	container.Env = envs.MergeEnvs(container.Env, e.Env...)
	container.EnvFrom = envs.MergeEnvFrom(container.EnvFrom, e.EnvFrom...)

	return nil
}

func (e *Environments) With(other *Environments) *Environments {
	if e == nil && other == nil {
		return nil
	}

	if e == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return e.DeepCopy()
	}

	return &Environments{
		Env:     envs.MergeEnvs(e.Env, other.Env...),
		EnvFrom: envs.MergeEnvFrom(e.EnvFrom, other.EnvFrom...),
	}
}

func (e *Environments) Validate() error {
	return nil
}
