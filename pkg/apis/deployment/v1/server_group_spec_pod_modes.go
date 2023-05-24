//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ServerGroupSpecPodMode struct {
	Network *ServerGroupNetworkMode `json:"network,omitempty"`
	PID     *ServerGroupPIDMode     `json:"pid,omitempty"`
}

func (s *ServerGroupSpecPodMode) GetNetwork() *ServerGroupNetworkMode {
	if s == nil {
		return nil
	}

	return s.Network
}

func (s *ServerGroupSpecPodMode) GetPID() *ServerGroupPIDMode {
	if s == nil {
		return nil
	}

	return s.PID
}

func (s *ServerGroupSpecPodMode) Validate() error {
	return errors.Wrapf(errors.Errors(s.GetNetwork().Validate(), s.GetPID().Validate()), "Validation of Pod modes failed")
}

func (s *ServerGroupSpecPodMode) Apply(p *core.PodSpec) {
	switch s.GetPID().Get() {
	case ServerGroupPIDModeIsolated:
	// Default, no change
	case ServerGroupPIDModePod:
		// Enable Pod shared namespaces
		p.ShareProcessNamespace = util.NewType[bool](true)
	case ServerGroupPIDModeHost:
		// Enable Host shared namespaces
		p.HostPID = true
	}

	switch s.GetNetwork().Get() {
	case ServerGroupNetworkModePod:
	// Default, no change
	case ServerGroupNetworkModeHost:
		// Enable Pod shared namespaces
		p.HostNetwork = true
	}
}
