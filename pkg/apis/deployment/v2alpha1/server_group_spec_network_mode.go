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

// ServerGroupNetworkMode is used to define Network mode of the Pod
type ServerGroupNetworkMode string

const (
	// ServerGroupNetworkModePod enable Pod level isolation of the network, default
	ServerGroupNetworkModePod ServerGroupNetworkMode = "pod"

	// ServerGroupNetworkModeHost enable Host level network access to the Pod
	ServerGroupNetworkModeHost ServerGroupNetworkMode = "host"

	DefaultServerGroupNetworkMode = ServerGroupNetworkModePod
)

func (n *ServerGroupNetworkMode) Validate() error {
	switch v := n.Get(); v {
	case ServerGroupNetworkModePod, ServerGroupNetworkModeHost:
		return nil
	default:
		return errors.WithStack(errors.Wrapf(ValidationError, "Unknown NetworkMode %s", v.String()))
	}
}

func (n *ServerGroupNetworkMode) Get() ServerGroupNetworkMode {
	if n == nil {
		return DefaultServerGroupNetworkMode
	}

	return *n
}

func (n *ServerGroupNetworkMode) String() string {
	return string(n.Get())
}

func (n *ServerGroupNetworkMode) New() *ServerGroupNetworkMode {
	v := n.Get()

	return &v
}
