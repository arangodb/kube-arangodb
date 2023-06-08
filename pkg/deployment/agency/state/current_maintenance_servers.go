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

package state

type CurrentMaintenanceServers map[Server]CurrentMaintenanceServer

func (s CurrentMaintenanceServers) InMaintenance(server Server) bool {
	if v, ok := s[server]; ok {
		return v.InMaintenance()
	}

	return false
}

type CurrentMaintenanceServerMode string

const (
	CurrentMaintenanceServerModeMaintenance CurrentMaintenanceServerMode = "maintenance"
)

type CurrentMaintenanceServer struct {
	Mode  *CurrentMaintenanceServerMode `json:"Mode,omitempty"`
	Until Timestamp                     `json:"Until,omitempty"`
}

func (s *CurrentMaintenanceServer) InMaintenance() bool {
	if s != nil {
		if m := s.Mode; m != nil {
			if *m == CurrentMaintenanceServerModeMaintenance {
				return true
			}
		}
	}

	return false
}
