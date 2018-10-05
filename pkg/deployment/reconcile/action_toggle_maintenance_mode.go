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

package reconcile

import (
	"context"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
)

// MaintenanceModeState is a strongly typed name for maintenance mode
type MaintenanceModeState string

const (
	// MaintenanceModeStateOn causes cluster supervision to pause
	MaintenanceModeStateOn MaintenanceModeState = "on"
	// MaintenanceModeStateOff causes cluster supervision to resume
	MaintenanceModeStateOff MaintenanceModeState = "off"

	superVisionStateMaintenance = "Maintenance"
	superVisionStateNormal      = "Normal"
)

var (
	superVisionMaintenanceKey = []string{"arango", "Supervision", "Maintenance"}
	superVisionStateKey       = []string{"arango", "Supervision", "State"}
)

// NewToggleMaintenanceModeAction toggles the maintenance mode
func NewToggleMaintenanceModeAction(log zerolog.Logger, action api.Action,
	actionCtx ActionContext, onOff MaintenanceModeState) Action {
	return &actionToggleMaintenanceMode{
		log:       log,
		action:    action,
		actionCtx: actionCtx,
		state:     onOff,
	}
}

// actionToggleMaintenanceMode implements an togglint the maintenance mode
type actionToggleMaintenanceMode struct {
	log       zerolog.Logger
	action    api.Action
	actionCtx ActionContext
	state     MaintenanceModeState
}

// Start performs the start of the action.
// Returns true if the action is completely finished, false in case
// the start time needs to be recorded and a ready condition needs to be checked.
func (a *actionToggleMaintenanceMode) Start(ctx context.Context) (bool, error) {
	if a.action.Group.IsArangosync() {
		return true, nil // nothing to do
	}
	switch a.actionCtx.GetMode() {

	case api.DeploymentModeCluster:
	case api.DeploymentModeActiveFailover:
		if a.action.Group == api.ServerGroupAgents {
			return true, nil // nothing to do
		}
		return a.toggleMaintenanceMode(ctx)
	case api.DeploymentModeSingle:
	default:
	}
	return true, nil // nothing to do
}

// CheckProgress checks the progress of the action.
// Returns true if the action is completely finished, false otherwise.
func (a *actionToggleMaintenanceMode) CheckProgress(ctx context.Context) (bool, bool, error) {
	return true, false, nil
}

func (a *actionToggleMaintenanceMode) toggleMaintenanceMode(ctx context.Context) (bool, error) {
	supported, err := a.isSuperVisionMaintenanceSupported(ctx)
	if err != nil {
		return false, maskAny(err)
	}
	if !supported {
		a.log.Info().Msg("Supervision maintenance is not supported on this version")
		return false, nil
	}

	if a.state == MaintenanceModeStateOn {
		err = a.disableSupervision(ctx)
	} else if a.state == MaintenanceModeStateOff {
		err = a.enableSupervision(ctx)
	}
	return true, err
}

// Timeout returns the amount of time after which this action will timeout.
func (a *actionToggleMaintenanceMode) Timeout() time.Duration {
	return toggleMaintenanceModeTimeout
}

// Return the MemberID used / created in this action
func (a *actionToggleMaintenanceMode) MemberID() string {
	return a.action.MemberID
}

// isSuperVisionMaintenanceSupported checks all agents for their version number.
// If it is to low to support supervision maintenance mode, false is returned.
func (a *actionToggleMaintenanceMode) isSuperVisionMaintenanceSupported(ctx context.Context) (bool, error) {
	// get all agents
	connections, err := a.actionCtx.GetAgencyClients(ctx)
	if err != nil {
		return false, maskAny(err)
	}

	// Check agent versions
	for _, conn := range connections {
		// authentication should be already configured
		c, err := driver.NewClient(driver.ClientConfig{
			Connection: conn,
		})
		if err != nil {
			return false, maskAny(err)
		}
		info, err := c.Version(ctx)
		if err != nil {
			return false, maskAny(err)
		}
		version := driver.Version(info.Version)
		if version.Major() < 3 {
			return false, nil
		}
		if version.Major() == 3 {
			sub, _ := version.SubInt()
			switch version.Minor() {
			case 0, 1:
				return false, nil
			case 2:
				if sub < 14 {
					return false, nil
				}
			case 3:
				if sub < 8 {
					return false, nil
				}
			}
		}
	}
	return true, nil
}

// disableSupervision blocks supervision of the agency and waits for the agency to acknowledge.
func (a *actionToggleMaintenanceMode) disableSupervision(ctx context.Context) error {
	api, err := a.actionCtx.GetAgency(ctx)
	if err != nil {
		return maskAny(err)
	}

	superVisionMaintenanceTTL := toggleMaintenanceModeTimeout
	// Set maintenance mode
	if err := api.WriteKey(ctx, superVisionMaintenanceKey, struct{}{}, superVisionMaintenanceTTL); err != nil {
		return maskAny(err)
	}
	// Wait for agency to acknowledge
	for {
		var value interface{}
		err := api.ReadKey(ctx, superVisionStateKey, &value)
		if err != nil {
			a.log.Warn().Err(err).Msg("Failed to read supervision state")
		} else if valueStr, ok := getMaintenanceMode(value); !ok {
			a.log.Warn().Msgf("Supervision state is not a string but: %v", value)
		} else if valueStr != superVisionStateMaintenance {
			a.log.Warn().Msgf("Supervision state is not yet '%s' but '%s'", superVisionStateMaintenance, valueStr)
		} else {
			return nil
		}
		select {
		case <-ctx.Done():
			return maskAny(ctx.Err())
		case <-time.After(time.Second):
			// Try again
		}
	}
}

func getMaintenanceMode(value interface{}) (string, bool) {
	if s, ok := value.(string); ok {
		return s, true
	}
	if m, ok := value.(map[string]interface{}); ok {
		if mode, ok := m["Mode"]; ok {
			return getMaintenanceMode(mode)
		} else if mode, ok := m["mode"]; ok {
			return getMaintenanceMode(mode)
		}
	}
	return "", false
}

// enableSupervision enabled supervision of the agency.
func (a *actionToggleMaintenanceMode) enableSupervision(ctx context.Context) error {

	api, err := a.actionCtx.GetAgency(ctx)
	if err != nil {
		return maskAny(err)
	}

	// Remove maintenance mode
	if err := api.RemoveKey(ctx, superVisionMaintenanceKey); err != nil {
		return maskAny(err)
	}
	return nil
}
