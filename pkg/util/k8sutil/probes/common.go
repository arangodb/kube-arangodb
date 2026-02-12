//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package probes

import (
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Common struct {
	// Number of seconds after the container has started before liveness probes are initiated (defaults to 30)
	InitialDelaySeconds *int32
	// Number of seconds after which the probe times out (defaults to 2).
	TimeoutSeconds *int32
	// How often (in seconds) to perform the probe (defaults to 10).
	PeriodSeconds *int32
	// Minimum consecutive successes for the probe to be considered successful after having failed (defaults to 1).
	SuccessThreshold *int32
	// Minimum consecutive failures for the probe to be considered failed after having succeeded (defaults to 3).
	FailureThreshold *int32
}

func (config *Common) SetSpec(spec *api.ServerGroupProbeSpec) {
	if spec == nil {
		return
	}

	if config.InitialDelaySeconds != nil {
		config.InitialDelaySeconds = spec.InitialDelaySeconds
	}

	if config.TimeoutSeconds != nil {
		config.TimeoutSeconds = spec.TimeoutSeconds
	}

	if config.PeriodSeconds != nil {
		config.PeriodSeconds = spec.PeriodSeconds
	}

	if config.SuccessThreshold != nil {
		config.SuccessThreshold = spec.SuccessThreshold
	}

	if config.FailureThreshold != nil {
		config.FailureThreshold = spec.FailureThreshold
	}
}

func (config *Common) config(handler core.ProbeHandler) *core.Probe {
	return &core.Probe{
		ProbeHandler:        handler,
		InitialDelaySeconds: util.OptionalType(config.InitialDelaySeconds, 900), // Wait 15min before first probe
		TimeoutSeconds:      util.OptionalType(config.TimeoutSeconds, 2),        // Timeout of each probe is 2s
		PeriodSeconds:       util.OptionalType(config.PeriodSeconds, 60),        // Interval between probes is 10s
		SuccessThreshold:    util.OptionalType(config.SuccessThreshold, 1),      // Single probe is enough to indicate success
		FailureThreshold:    util.OptionalType(config.FailureThreshold, 10),     // Need 10 failed probes to consider a failed state
	}
}
