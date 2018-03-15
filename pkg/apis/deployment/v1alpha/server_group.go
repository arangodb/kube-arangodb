//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package v1alpha

type ServerGroup int

const (
	ServerGroupSingle       ServerGroup = 1
	ServerGroupAgents       ServerGroup = 2
	ServerGroupDBServers    ServerGroup = 3
	ServerGroupCoordinators ServerGroup = 4
	ServerGroupSyncMasters  ServerGroup = 5
	ServerGroupSyncWorkers  ServerGroup = 6
)

var (
	// AllServerGroups contains a constant list of all known server groups
	AllServerGroups = []ServerGroup{
		ServerGroupSingle,
		ServerGroupAgents,
		ServerGroupDBServers,
		ServerGroupCoordinators,
		ServerGroupSyncMasters,
		ServerGroupSyncWorkers,
	}
)

// AsRole returns the "role" value for the given group.
func (g ServerGroup) AsRole() string {
	switch g {
	case ServerGroupSingle:
		return "sngl"
	case ServerGroupAgents:
		return "agnt"
	case ServerGroupDBServers:
		return "prmr"
	case ServerGroupCoordinators:
		return "crdn"
	case ServerGroupSyncMasters:
		return "syma"
	case ServerGroupSyncWorkers:
		return "sywo"
	default:
		return "?"
	}
}

// IsArangod returns true when the groups runs servers of type `arangod`.
func (g ServerGroup) IsArangod() bool {
	switch g {
	case ServerGroupSingle, ServerGroupAgents, ServerGroupDBServers, ServerGroupCoordinators:
		return true
	default:
		return false
	}
}

// IsArangosync returns true when the groups runs servers of type `arangosync`.
func (g ServerGroup) IsArangosync() bool {
	switch g {
	case ServerGroupSyncMasters, ServerGroupSyncWorkers:
		return true
	default:
		return false
	}
}
