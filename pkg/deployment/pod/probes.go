//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package pod

import api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

func newProbe(canBeEnabled, enabledByDefault bool) Probe {
	return Probe{
		EnabledByDefault: enabledByDefault,
		CanBeEnabled:     canBeEnabled,
	}
}

func ReadinessSpec(group api.ServerGroup) Probe {
	return probeMap[group].readiness
}

func LivenessSpec(group api.ServerGroup) Probe {
	return probeMap[group].liveness
}

type Probe struct {
	CanBeEnabled, EnabledByDefault bool
}

type probes struct {
	liveness, readiness Probe
}

// probeMap defines default values and if Probe can be enabled
var probeMap = map[api.ServerGroup]probes{
	api.ServerGroupSingle: {
		liveness: newProbe(true, true),
		readiness: newProbe(true, true),
	},
	api.ServerGroupAgents: {
		liveness: newProbe(true, true),
		readiness: newProbe(false, false),
	},
	api.ServerGroupDBServers: {
		liveness: newProbe(true, true),
		readiness: newProbe(false, false),
	},
	api.ServerGroupCoordinators: {
		liveness: newProbe(false, false),
		readiness: newProbe(true, true),
	},
	api.ServerGroupSyncMasters: {
		liveness: newProbe(true, true),
		readiness: newProbe(false, false),
	},
	api.ServerGroupSyncWorkers: {
		liveness: newProbe(true, true),
		readiness: newProbe(false, false),
	},
}