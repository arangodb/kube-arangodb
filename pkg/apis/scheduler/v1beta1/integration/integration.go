//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package integration

import (
	schedulerContainerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Sidecar struct {
	// ListenPort defines on which port the sidecar container will be listening for connections
	// +doc/default: 9201
	ListenPort *uint16 `json:"listenPort,omitempty"`

	// ControllerListenPort defines on which port the sidecar container will be listening for controller requests
	// +doc/default: 9202
	ControllerListenPort *uint16 `json:"controllerListenPort,omitempty"`

	// HTTPListenPort defines on which port the sidecar container will be listening for connections on http
	// +doc/default: 9203
	HTTPListenPort *uint16 `json:"httpListenPort,omitempty"`

	// Container Keeps the information about Container configuration
	*schedulerContainerApi.Container `json:",inline"`
}

func (s *Sidecar) GetContainer() *schedulerContainerApi.Container {
	if s == nil || s.Container == nil {
		return nil
	}

	return s.Container
}

func (s *Sidecar) Validate() error {
	if s == nil {
		s = &Sidecar{}
	}

	var err []error

	if s.GetListenPort() < 1 {
		err = append(err, shared.PrefixResourceErrors("listenPort", errors.Errorf("must be positive")))
	}

	if s.GetControllerListenPort() < 1 {
		err = append(err, shared.PrefixResourceErrors("controllerListenPort", errors.Errorf("must be positive")))
	}

	if s.GetHTTPListenPort() < 1 {
		err = append(err, shared.PrefixResourceErrors("httpListenPort", errors.Errorf("must be positive")))
	}

	err = append(err, s.GetContainer().Validate())

	return shared.WithErrors(err...)
}

func (s *Sidecar) GetListenPort() uint16 {
	if s == nil || s.ListenPort == nil {
		return 9201
	}
	return *s.ListenPort
}

func (s *Sidecar) GetControllerListenPort() uint16 {
	if s == nil || s.ControllerListenPort == nil {
		return 9202
	}
	return *s.ControllerListenPort
}

func (s *Sidecar) GetHTTPListenPort() uint16 {
	if s == nil || s.HTTPListenPort == nil {
		return 9203
	}
	return *s.HTTPListenPort
}
