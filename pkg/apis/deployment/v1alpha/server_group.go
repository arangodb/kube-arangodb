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

import time "time"

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
		return "single"
	case ServerGroupAgents:
		return "agent"
	case ServerGroupDBServers:
		return "dbserver"
	case ServerGroupCoordinators:
		return "coordinator"
	case ServerGroupSyncMasters:
		return "syncmaster"
	case ServerGroupSyncWorkers:
		return "syncworker"
	default:
		return "?"
	}
}

// AsRoleAbbreviated returns the abbreviation of the "role" value for the given group.
func (g ServerGroup) AsRoleAbbreviated() string {
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

// DefaultTerminationGracePeriod returns the default period between SIGTERM & SIGKILL for a server in the given group.
func (g ServerGroup) DefaultTerminationGracePeriod() time.Duration {
	switch g {
	case ServerGroupSingle:
		return time.Minute
	case ServerGroupAgents:
		return time.Minute
	case ServerGroupDBServers:
		return time.Hour
	default:
		return time.Second * 30
	}
}

// IsStateless returns true when the groups runs servers without a persistent volume.
func (g ServerGroup) IsStateless() bool {
	switch g {
	case ServerGroupCoordinators, ServerGroupSyncMasters, ServerGroupSyncWorkers:
		return true
	default:
		return false
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
