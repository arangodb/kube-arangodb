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

package reconcile

import (
	"strconv"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/agency/state"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

func getResignLeadershipActionType() api.ActionType {
	if features.EnforcedResignLeadership().Enabled() {
		return api.ActionTypeEnforceResignLeadership
	}
	return api.ActionTypeResignLeadership
}

// isServerRebooted returns true when a given server ID was rebooted during resignation of leadership.
func isServerRebooted(log logging.Logger, action api.Action, agencyState state.State, serverID driver.ServerID) bool {
	rebootID, ok := agencyState.GetRebootID(serverID)
	if !ok {
		return false
	}

	v, ok := action.Params[actionResignLeadershipRebootID.String()]
	if !ok {
		return false
	}

	r, err := strconv.Atoi(v)
	if err != nil {
		log.Err(err).Warn("reboot ID '%s' supposed to be a number", v)
		return false
	}

	if rebootID <= r {
		// Server has not been restarted.
		return false
	}

	log.Warn("resign leadership aborted because rebootID has changed from %d to %d", r, rebootID)
	return true
}
