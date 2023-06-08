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

import (
	"encoding/json"
)

var (
	_ json.Unmarshaler = &ArangoSyncLazy{}
)

// ArangoSyncLazy allows failure during load of the Sync state
type ArangoSyncLazy struct {
	Error error
	*ArangoSync
}

func (a *ArangoSyncLazy) UnmarshalJSON(bytes []byte) error {
	var s ArangoSync

	if err := json.Unmarshal(bytes, &s); err != nil {
		a.ArangoSync = nil
		a.Error = err
	} else {
		a.ArangoSync = &s
		a.Error = nil
	}

	return nil
}

type ArangoSync struct {
	State ArangoSyncState `json:"synchronizationState"`
}

func (a *ArangoSync) IsSyncInProgress() bool {
	if a == nil {
		return false
	}

	// Check Incoming
	if s := a.State.Incoming.State; s != nil && *s != "inactive" && *s != "" {
		return true
	}

	if a.State.Outgoing.Targets.Exists() {
		return true
	}

	return false
}

type ArangoSyncState struct {
	Incoming ArangoSyncIncomingState `json:"incoming"`
	Outgoing ArangoSyncOutgoingState `json:"outgoing"`
}

type ArangoSyncIncomingState struct {
	State *string `json:"state,omitempty"`
}

type ArangoSyncOutgoingState struct {
	Targets Exists `json:"targets,omitempty"`
}
