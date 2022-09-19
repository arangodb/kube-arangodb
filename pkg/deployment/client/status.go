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

package client

import (
	"fmt"
	"strings"
)

const (
	// ServerProgressPhaseInWait describes success progress state of a server.
	ServerProgressPhaseInWait = "in wait"
	// ServerStatusEndpoint describes endpoint of a server status.
	ServerStatusEndpoint = "/_admin/status"
	// ServerApiVersionEndpoint describes endpoint of a server version.
	ServerApiVersionEndpoint = "/_api/version"
	// ServerAvailabilityEndpoint describes endpoint of a server availability.
	ServerAvailabilityEndpoint = "/_admin/server/availability"
)

// ServerProgress describes server progress.
type ServerProgress struct {
	// Phase is a name of the lifecycle phase the instance is currently in.
	Phase string `json:"phase,omitempty"`
	// Feature is internal name of the feature that is currently being prepared
	Feature string `json:"feature,omitempty"`
	// Current recovery sequence number value, if the instance is currently recovering.
	// If the instance is already past the recovery, this attribute contains the last handled recovery sequence number.
	RecoveryTick int `json:"recoveryTick,omitempty"`
}

// ServerInfo describes server information.
type ServerInfo struct {
	ServerProgress ServerProgress `json:"progress,omitempty"`
}

// ServerStatus describes server status.
type ServerStatus struct {
	ServerInfo ServerInfo `json:"serverInfo,omitempty"`
}

// GetProgress returns human-readable progress status of the server, and true if server is ready.
func (s ServerStatus) GetProgress() (string, bool) {
	p := s.ServerInfo.ServerProgress
	var result strings.Builder

	if len(p.Feature) > 0 {
		result.WriteString("feature: " + p.Feature + ", ")
	}

	result.WriteString("phase: " + p.Phase)

	if p.RecoveryTick > 0 {
		result.WriteString(", recoveryTick: " + fmt.Sprintf("%d", p.RecoveryTick))
	}

	return result.String(), p.Phase == ServerProgressPhaseInWait
}
