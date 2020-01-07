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

package v1

import "time"

type ServerGroup int

const (
	ServerGroupUnknown      ServerGroup = 0
	ServerGroupSingle       ServerGroup = 1
	ServerGroupAgents       ServerGroup = 2
	ServerGroupDBServers    ServerGroup = 3
	ServerGroupCoordinators ServerGroup = 4
	ServerGroupSyncMasters  ServerGroup = 5
	ServerGroupSyncWorkers  ServerGroup = 6

	ServerGroupSingleString       = "single"
	ServerGroupAgentsString       = "agent"
	ServerGroupDBServersString    = "dbserver"
	ServerGroupCoordinatorsString = "coordinator"
	ServerGroupSyncMastersString  = "syncmaster"
	ServerGroupSyncWorkersString  = "syncworker"

	ServerGroupSingleAbbreviatedString       = "sngl"
	ServerGroupAgentsAbbreviatedString       = "agnt"
	ServerGroupDBServersAbbreviatedString    = "prmr"
	ServerGroupCoordinatorsAbbreviatedString = "crdn"
	ServerGroupSyncMastersAbbreviatedString  = "syma"
	ServerGroupSyncWorkersAbbreviatedString  = "sywo"
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
		return ServerGroupSingleString
	case ServerGroupAgents:
		return ServerGroupAgentsString
	case ServerGroupDBServers:
		return ServerGroupDBServersString
	case ServerGroupCoordinators:
		return ServerGroupCoordinatorsString
	case ServerGroupSyncMasters:
		return ServerGroupSyncMastersString
	case ServerGroupSyncWorkers:
		return ServerGroupSyncWorkersString
	default:
		return "?"
	}
}

// AsRoleAbbreviated returns the abbreviation of the "role" value for the given group.
func (g ServerGroup) AsRoleAbbreviated() string {
	switch g {
	case ServerGroupSingle:
		return ServerGroupSingleAbbreviatedString
	case ServerGroupAgents:
		return ServerGroupAgentsAbbreviatedString
	case ServerGroupDBServers:
		return ServerGroupDBServersAbbreviatedString
	case ServerGroupCoordinators:
		return ServerGroupCoordinatorsAbbreviatedString
	case ServerGroupSyncMasters:
		return ServerGroupSyncMastersAbbreviatedString
	case ServerGroupSyncWorkers:
		return ServerGroupSyncWorkersAbbreviatedString
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

// IsExportMetrics return true when the group can be used with the arangodbexporter
func (g ServerGroup) IsExportMetrics() bool {
	switch g {
	case ServerGroupCoordinators, ServerGroupDBServers, ServerGroupSingle:
		return true
	default:
		return false
	}
}

// ServerGroupFromAbbreviatedRole returns ServerGroup from abbreviated role
func ServerGroupFromAbbreviatedRole(label string) ServerGroup {
	switch label {
	case ServerGroupSingleAbbreviatedString:
		return ServerGroupSingle
	case ServerGroupAgentsAbbreviatedString:
		return ServerGroupAgents
	case ServerGroupDBServersAbbreviatedString:
		return ServerGroupDBServers
	case ServerGroupCoordinatorsAbbreviatedString:
		return ServerGroupCoordinators
	case ServerGroupSyncMastersAbbreviatedString:
		return ServerGroupSyncMasters
	case ServerGroupSyncWorkersAbbreviatedString:
		return ServerGroupSyncWorkers
	default:
		return ServerGroupUnknown
	}
}

// ServerGroupFromAbbreviatedRole returns ServerGroup from role
func ServerGroupFromRole(label string) ServerGroup {
	switch label {
	case ServerGroupSingleString:
		return ServerGroupSingle
	case ServerGroupAgentsString:
		return ServerGroupAgents
	case ServerGroupDBServersString:
		return ServerGroupDBServers
	case ServerGroupCoordinatorsString:
		return ServerGroupCoordinators
	case ServerGroupSyncMastersString:
		return ServerGroupSyncMasters
	case ServerGroupSyncWorkersString:
		return ServerGroupSyncWorkers
	default:
		return ServerGroupUnknown
	}
}
