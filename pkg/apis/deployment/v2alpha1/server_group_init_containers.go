//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	core "k8s.io/api/core/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	ServerGroupReservedInitContainerNameLifecycle    = "init-lifecycle"
	ServerGroupReservedInitContainerNameUUID         = "uuid"
	ServerGroupReservedInitContainerNameWait         = "wait"
	ServerGroupReservedInitContainerNameStartup      = "arango-init-startup"
	ServerGroupReservedInitContainerNameUpgrade      = "upgrade"
	ServerGroupReservedInitContainerNameVersionCheck = "version-check"
)

func IsReservedServerGroupInitContainerName(name string) bool {
	switch name {
	case ServerGroupReservedInitContainerNameLifecycle, ServerGroupReservedInitContainerNameUUID, ServerGroupReservedInitContainerNameUpgrade, ServerGroupReservedInitContainerNameVersionCheck, ServerGroupReservedInitContainerNameStartup:
		return true
	default:
		return false
	}
}

func ValidateServerGroupInitContainerName(name string) error {
	if IsReservedServerGroupInitContainerName(name) {
		return errors.Errorf("InitContainer name %s is restricted", name)
	}

	return sharedApi.AsKubernetesResourceName(&name).Validate()
}

type ServerGroupInitContainerMode string

func (s *ServerGroupInitContainerMode) Get() ServerGroupInitContainerMode {
	if s == nil {
		return ServerGroupInitContainerDefaultMode // default
	}

	return *s
}

func (s ServerGroupInitContainerMode) New() *ServerGroupInitContainerMode {
	return &s
}

func (s *ServerGroupInitContainerMode) Validate() error {
	switch v := s.Get(); v {
	case ServerGroupInitContainerIgnoreMode, ServerGroupInitContainerUpdateMode:
		return nil
	default:
		return errors.Errorf("Unknown serverGroupInitContainerMode %s", v)
	}
}

const (
	// ServerGroupInitContainerDefaultMode default mode
	ServerGroupInitContainerDefaultMode = ServerGroupInitContainerIgnoreMode
	// ServerGroupInitContainerIgnoreMode ignores init container changes in pod recreation flow
	ServerGroupInitContainerIgnoreMode ServerGroupInitContainerMode = "ignore"
	// ServerGroupInitContainerUpdateMode enforce update of pod if init container has been changed
	ServerGroupInitContainerUpdateMode ServerGroupInitContainerMode = "update"
)

type ServerGroupInitContainers struct {
	// Containers contains list of containers
	// +doc/type: []core.Container
	// +doc/link: Documentation of core.Container|https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.29/#container-v1-core
	Containers []core.Container `json:"containers,omitempty"`

	// Mode keep container replace mode
	// +doc/enum: update|Enforce update of pod if init container has been changed
	// +doc/enum: ignore|Ignores init container changes in pod recreation flow
	Mode *ServerGroupInitContainerMode `json:"mode,omitempty"`
}

func (s *ServerGroupInitContainers) GetMode() *ServerGroupInitContainerMode {
	if s == nil {
		return nil
	}

	return s.Mode
}

func (s *ServerGroupInitContainers) GetContainers() []core.Container {
	if s == nil {
		return nil
	}

	return s.Containers
}

func (s *ServerGroupInitContainers) Validate() error {
	if s == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("mode", s.Mode.Validate()),
		shared.PrefixResourceError("containers", s.validateInitContainers()),
	)
}

func (s *ServerGroupInitContainers) validateInitContainers() error {
	for _, c := range s.Containers {
		if err := ValidateServerGroupInitContainerName(c.Name); err != nil {
			return err
		}
	}

	return nil
}
