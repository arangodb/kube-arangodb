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

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/interfaces"
	schedulerPolicyApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/policy"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

var _ interfaces.Container[Core] = &Core{}

type Core struct {
	*schedulerPolicyApi.Policy `json:",inline"`

	// Entrypoint array. Not executed within a shell.
	// The container image's ENTRYPOINT is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
	// produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
	// of whether the variable exists or not. Cannot be updated.
	// +doc/link: Kubernetes Docs|https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Command []string `json:"command,omitempty"`

	// Arguments to the entrypoint.
	// The container image's CMD is used if this is not provided.
	// Variable references $(VAR_NAME) are expanded using the container's environment. If a variable
	// cannot be resolved, the reference in the input string will be unchanged. Double $$ are reduced
	// to a single $, which allows for escaping the $(VAR_NAME) syntax: i.e. "$$(VAR_NAME)" will
	// produce the string literal "$(VAR_NAME)". Escaped references will never be expanded, regardless
	// of whether the variable exists or not. Cannot be updated.
	// +doc/link: Kubernetes Docs|https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#running-a-command-in-a-shell
	Args []string `json:"args,omitempty"`

	// Container's working directory.
	// If not specified, the container runtime's default will be used, which
	// might be configured in the container image.
	WorkingDir string `json:"workingDir,omitempty"`
}

func (c *Core) Apply(_ *core.PodTemplateSpec, container *core.Container) error {
	if c == nil {
		return nil
	}

	d := c.DeepCopy()

	container.Args = d.Args
	container.Command = d.Command
	container.WorkingDir = d.WorkingDir

	return nil
}

func (c *Core) With(other *Core) *Core {
	if c == nil && other == nil {
		return nil
	}

	if other == nil {
		return c.DeepCopy()
	}

	if c == nil {
		return other.DeepCopy()
	}

	o := other.DeepCopy()

	if o.GetMethod(schedulerPolicyApi.Override) == schedulerPolicyApi.Append {
		o.Args = append(c.Args, o.Args...)
	}

	return o
}

func (c *Core) Validate() error {
	if c == nil {
		return nil
	}

	return shared.WithErrors(
		shared.ValidateOptionalInterface(c.Policy),
	)
}
