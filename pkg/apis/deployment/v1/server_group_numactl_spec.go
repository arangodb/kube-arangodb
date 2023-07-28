//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

import "github.com/arangodb/kube-arangodb/pkg/util"

const ServerGroupSpecNumactlPathDefault = "/usr/bin/numactl"

type ServerGroupSpecNumactl struct {
	// Enabled define if numactl should be enabled
	// +doc/default: false
	Enabled *bool `json:"enabled,omitempty"`

	// Path define numactl path within the container
	// +doc/default: /usr/bin/numactl
	Path *string `json:"path,omitempty"`

	// Args define list of the numactl process
	// +doc/default: []
	Args []string `json:"args,omitempty"`
}

// IsEnabled returns flag if Numactl should be enabled
func (s *ServerGroupSpecNumactl) IsEnabled() bool {
	if s == nil {
		return false
	}

	return util.TypeOrDefault(s.Enabled, false)
}

// GetPath returns path of the numactl binary
func (s *ServerGroupSpecNumactl) GetPath() string {
	if s == nil {
		return ServerGroupSpecNumactlPathDefault
	}

	return util.TypeOrDefault(s.Path, ServerGroupSpecNumactlPathDefault)
}

// GetArgs returns args of the numactl command
func (s *ServerGroupSpecNumactl) GetArgs() []string {
	if s == nil {
		return nil
	}

	return s.Args
}
