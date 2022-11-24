//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

// ServerGroupPIDMode define Pod PID share strategy
type ServerGroupPIDMode string

const (
	// ServerGroupPIDModeIsolated enable isolation of the Processes within Pod Container, default
	ServerGroupPIDModeIsolated ServerGroupPIDMode = "isolated"
	// ServerGroupPIDModePod enable isolation of the Processes on the Pod level. Processes started in this mode will have PID different from 1
	ServerGroupPIDModePod ServerGroupPIDMode = "pod"
	// ServerGroupPIDModeHost disable isolation of the Processes. Processes started in this mode are shared with the entire host
	ServerGroupPIDModeHost ServerGroupPIDMode = "host"

	DefaultServerGroupPIDMode = ServerGroupPIDModeIsolated
)

func (n *ServerGroupPIDMode) Validate() error {
	switch v := n.Get(); v {
	case ServerGroupPIDModeIsolated, ServerGroupPIDModePod, ServerGroupPIDModeHost:
		return nil
	default:
		return errors.WithStack(errors.Wrapf(ValidationError, "Unknown PIDMode %s", v.String()))
	}
}

func (n *ServerGroupPIDMode) Get() ServerGroupPIDMode {
	if n == nil {
		return DefaultServerGroupPIDMode
	}

	return *n
}

func (n *ServerGroupPIDMode) String() string {
	return string(n.Get())
}

func (n *ServerGroupPIDMode) New() *ServerGroupPIDMode {
	v := n.Get()

	return &v
}
