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

package deployment

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

var ciLogger = logging.Global().RegisterAndGetLogger("deployment-ci", logging.Info)

// clusterScalingIntegration is a helper to communicate with the clusters
// scaling UI.
type clusterScalingIntegration struct {
	log           logging.Logger
	depl          *Deployment
	pendingUpdate struct {
		mutex sync.Mutex
		spec  *api.DeploymentSpec
	}
	lastNumberOfServers struct {
		arangod.NumberOfServers
		mutex sync.Mutex
	}
	scaleEnabled struct {
		mutex   sync.Mutex
		enabled bool
	}
}

func (ci *clusterScalingIntegration) WrapLogger(in *zerolog.Event) *zerolog.Event {
	return in.Str("namespace", ci.depl.GetNamespace()).Str("name", ci.depl.Name())
}

const (
	maxClusterBootstrapTime = time.Minute * 2 // Time we allow a cluster bootstrap to take, before we can do cluster inspections.
)

// newClusterScalingIntegration creates a new clusterScalingIntegration.
func newClusterScalingIntegration(depl *Deployment) *clusterScalingIntegration {
	ci := &clusterScalingIntegration{
		depl: depl,
	}
	ci.log = ciLogger.WrapObj(ci)
	ci.scaleEnabled.enabled = true
	return ci
}

// SendUpdateToCluster records the given spec to be sended to the cluster.
func (ci *clusterScalingIntegration) SendUpdateToCluster(spec api.DeploymentSpec) {
	ci.pendingUpdate.mutex.Lock()
	defer ci.pendingUpdate.mutex.Unlock()
	ci.pendingUpdate.spec = spec.DeepCopy()
}

// checkScalingCluster checks if inspection
// returns true if inspection occurred
func (ci *clusterScalingIntegration) checkScalingCluster(ctx context.Context, expectSuccess bool) bool {
	ci.scaleEnabled.mutex.Lock()
	defer ci.scaleEnabled.mutex.Unlock()

	if !ci.depl.config.ScalingIntegrationEnabled {
		if err := ci.cleanClusterServers(ctx); err != nil {
			ci.log.Err(err).Debug("Clean failed")
			return false
		}
		return true
	}

	status := ci.depl.GetStatus()

	if !ci.scaleEnabled.enabled {
		// Check if it is possible to turn on scaling without any issue
		if status.Plan.IsEmpty() && ci.setNumberOfServers(ctx) == nil {
			// Scaling should be enabled because there is no Plan.
			// It can happen when the enabling action fails
			ci.scaleEnabled.enabled = true
		}
	}

	if ci.depl.GetPhase() != api.DeploymentPhaseRunning || !ci.scaleEnabled.enabled {
		// Deployment must be in running state and scaling must be enabled
		return false
	}

	// Update cluster with our state
	safeToAskCluster, err := ci.updateClusterServerCount(ctx, expectSuccess)
	if err != nil {
		if expectSuccess {
			ci.log.Err(err).Debug("Cluster update failed")
		}
	} else if safeToAskCluster {
		// Inspect once
		if err := ci.inspectCluster(ctx, expectSuccess); err != nil {
			if expectSuccess {
				ci.log.Err(err).Debug("Cluster inspection failed")
			}
		} else {
			return true
		}
	}
	return false
}

// ListenForClusterEvents keep listening for changes entered in the UI of the cluster.
func (ci *clusterScalingIntegration) ListenForClusterEvents(stopCh <-chan struct{}) {
	start := time.Now()
	goodInspections := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {
		expectSuccess := goodInspections > 0 || time.Since(start) > maxClusterBootstrapTime

		if ci.checkScalingCluster(ctx, expectSuccess) {
			goodInspections++
		}

		select {
		case <-timer.After(time.Second * 2):
			// Continue
		case <-stopCh:
			// We're done
			return
		}
	}
}

func (ci *clusterScalingIntegration) cleanClusterServers(ctx context.Context) error {
	log := ci.log

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := ci.depl.clientCache.GetDatabase(ctxChild)
	if err != nil {
		return errors.WithStack(err)
	}

	req, err := arangod.GetNumberOfServers(ctxChild, c.Connection())
	if err != nil {
		log.Err(err).Debug("Failed to get number of servers")
		return errors.WithStack(err)
	}

	if req.Coordinators != nil || req.DBServers != nil {
		log.Debug("Clean number of servers")
		if err := arangod.CleanNumberOfServers(ctx, c.Connection()); err != nil {
			log.Err(err).Debug("Failed to clean number of servers")
			return errors.WithStack(err)
		}
	}

	return nil
}

// Perform a single inspection of the cluster
func (ci *clusterScalingIntegration) inspectCluster(ctx context.Context, expectSuccess bool) error {
	log := ci.log

	ctxChild, cancel := globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	c, err := ci.depl.clientCache.GetDatabase(ctxChild)
	if err != nil {
		return errors.WithStack(err)
	}

	ctxChild, cancel = globals.GetGlobalTimeouts().ArangoD().WithTimeout(ctx)
	defer cancel()
	req, err := arangod.GetNumberOfServers(ctxChild, c.Connection())
	if err != nil {
		if expectSuccess {
			log.Err(err).Debug("Failed to get number of servers")
		}
		return errors.WithStack(err)
	}
	if req.Coordinators == nil && req.DBServers == nil {
		// Nothing to check
		return nil
	}
	coordinatorsChanged := false
	dbserversChanged := false
	ci.lastNumberOfServers.mutex.Lock()
	defer ci.lastNumberOfServers.mutex.Unlock()
	desired := ci.lastNumberOfServers.NumberOfServers
	if req.Coordinators != nil && desired.Coordinators != nil && req.GetCoordinators() != desired.GetCoordinators() {
		// #Coordinator has changed
		coordinatorsChanged = true
	}
	if req.DBServers != nil && desired.DBServers != nil && req.GetDBServers() != desired.GetDBServers() {
		// #DBServers has changed
		dbserversChanged = true
	}
	if !coordinatorsChanged && !dbserversChanged {
		// if there is nothing to change, check if we never have asked the cluster before
		// if so, fill in the values for the first time.
		// This happens, when the operator is redeployed and there has not been any
		// update events yet.
		if desired.Coordinators == nil || desired.DBServers == nil {
			if req.Coordinators != nil {
				ci.lastNumberOfServers.NumberOfServers.Coordinators = req.Coordinators
			}
			if req.DBServers != nil {
				ci.lastNumberOfServers.NumberOfServers.DBServers = req.DBServers
			}
		}

		// Nothing has changed
		return nil
	}
	// Let's update the spec
	p := make([]patch.Item, 0, 2)
	if coordinatorsChanged {
		if min, max, expected := ci.depl.GetSpec().Coordinators.GetMinCount(), ci.depl.GetSpec().Coordinators.GetMaxCount(), req.GetCoordinators(); min <= expected && expected <= max {
			p = append(p, patch.ItemReplace(patch.NewPath("spec", "coordinators", "count"), expected))
		}
	}
	if dbserversChanged {
		if min, max, expected := ci.depl.GetSpec().DBServers.GetMinCount(), ci.depl.GetSpec().DBServers.GetMaxCount(), req.GetDBServers(); min <= expected && expected <= max {
			p = append(p, patch.ItemReplace(patch.NewPath("spec", "dbservers", "count"), expected))
		}
	}
	return ci.depl.ApplyPatch(ctx, p...)
}

// updateClusterServerCount updates the intended number of servers of the cluster.
// Returns true when it is safe to ask the cluster for updates.
func (ci *clusterScalingIntegration) updateClusterServerCount(ctx context.Context, expectSuccess bool) (bool, error) {
	// Any update needed?
	ci.pendingUpdate.mutex.Lock()
	spec := ci.pendingUpdate.spec
	ci.pendingUpdate.mutex.Unlock()

	if spec == nil {
		// Nothing pending
		return true, nil
	}

	log := ci.log
	var coordinatorCountPtr *int
	var dbserverCountPtr *int

	coordinatorCount, dbserverCount := ci.getNumbersOfServers()

	if spec.Coordinators.GetMaxCount() == spec.Coordinators.GetMinCount() {
		coordinatorCountPtr = nil
	} else {
		coordinatorCountPtr = &coordinatorCount
	}

	if spec.DBServers.GetMaxCount() == spec.DBServers.GetMinCount() {
		dbserverCountPtr = nil
	} else {
		dbserverCountPtr = &dbserverCount
	}

	lastNumberOfServers := ci.GetLastNumberOfServers()

	// This is to prevent unneseccary updates that may override some values written by the WebUI (in the case of a update loop)
	if coordinatorCount != lastNumberOfServers.GetCoordinators() || dbserverCount != lastNumberOfServers.GetDBServers() {
		log.Debug("Setting number of servers %d/%d", coordinatorCount, dbserverCount)
		if err := ci.depl.SetNumberOfServers(ctx, coordinatorCountPtr, dbserverCountPtr); err != nil {
			if expectSuccess {
				log.Err(err).Debug("Failed to set number of servers")
			}
			return false, errors.WithStack(err)
		}
	}

	// Success, now update internal state
	safeToAskCluster := false
	ci.pendingUpdate.mutex.Lock()
	if spec == ci.pendingUpdate.spec {
		ci.pendingUpdate.spec = nil
		safeToAskCluster = true
	}
	ci.pendingUpdate.mutex.Unlock()

	ci.lastNumberOfServers.mutex.Lock()
	defer ci.lastNumberOfServers.mutex.Unlock()

	ci.lastNumberOfServers.Coordinators = &coordinatorCount
	ci.lastNumberOfServers.DBServers = &dbserverCount
	return safeToAskCluster, nil
}

// GetLastNumberOfServers returns the last number of servers
func (ci *clusterScalingIntegration) GetLastNumberOfServers() arangod.NumberOfServers {
	ci.lastNumberOfServers.mutex.Lock()
	defer ci.lastNumberOfServers.mutex.Unlock()

	return ci.lastNumberOfServers.NumberOfServers
}

// DisableScalingCluster disables scaling DBservers and coordinators
func (ci *clusterScalingIntegration) DisableScalingCluster(ctx context.Context) error {
	ci.scaleEnabled.mutex.Lock()
	defer ci.scaleEnabled.mutex.Unlock()

	// Turn off scaling DBservers and coordinators in arangoDB for the UI
	if err := ci.depl.SetNumberOfServers(ctx, nil, nil); err != nil {
		return errors.WithStack(err)
	}

	ci.scaleEnabled.enabled = false
	return nil
}

// EnableScalingCluster enables scaling DBservers and coordinators
func (ci *clusterScalingIntegration) EnableScalingCluster(ctx context.Context) error {
	ci.scaleEnabled.mutex.Lock()
	defer ci.scaleEnabled.mutex.Unlock()

	if ci.scaleEnabled.enabled {
		return nil
	}

	if err := ci.setNumberOfServers(ctx); err != nil {
		return errors.WithStack(err)
	}
	ci.scaleEnabled.enabled = true
	return nil
}

func (ci *clusterScalingIntegration) setNumberOfServers(ctx context.Context) error {
	numOfCoordinators, numOfDBServers := ci.getNumbersOfServers()
	return ci.depl.SetNumberOfServers(ctx, &numOfCoordinators, &numOfDBServers)
}

func (ci *clusterScalingIntegration) getNumbersOfServers() (int, int) {
	status := ci.depl.GetStatus()
	return len(status.Members.Coordinators), len(status.Members.DBServers)
}
