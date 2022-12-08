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

package probes

import (
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

// HTTPProbeConfig contains settings for creating a liveness/readiness probe.
type HTTPProbeConfig struct {
	// Local path to GET
	LocalPath string // `e.g. /_api/version`
	// Secure connection?
	Secure bool
	// Value for an Authorization header (can be empty)
	Authorization string
	// PortName define port name used to connect to the server for probes
	PortName string
	// Number of seconds after the container has started before liveness probes are initiated (defaults to 30)
	InitialDelaySeconds int32
	// Number of seconds after which the probe times out (defaults to 2).
	TimeoutSeconds int32
	// How often (in seconds) to perform the probe (defaults to 10).
	PeriodSeconds int32
	// Minimum consecutive successes for the probe to be considered successful after having failed (defaults to 1).
	SuccessThreshold int32
	// Minimum consecutive failures for the probe to be considered failed after having succeeded (defaults to 3).
	FailureThreshold int32
}

func (config *HTTPProbeConfig) SetSpec(spec *api.ServerGroupProbeSpec) {
	config.InitialDelaySeconds = spec.GetInitialDelaySeconds(config.InitialDelaySeconds)
	config.TimeoutSeconds = spec.GetTimeoutSeconds(config.TimeoutSeconds)
	config.PeriodSeconds = spec.GetPeriodSeconds(config.PeriodSeconds)
	config.SuccessThreshold = spec.GetSuccessThreshold(config.SuccessThreshold)
	config.FailureThreshold = spec.GetFailureThreshold(config.FailureThreshold)
}

// Create creates a probe from given config
func (config HTTPProbeConfig) Create() *core.Probe {
	scheme := core.URISchemeHTTP
	if config.Secure {
		scheme = core.URISchemeHTTPS
	}
	var headers []core.HTTPHeader
	if config.Authorization != "" {
		headers = append(headers, core.HTTPHeader{
			Name:  "Authorization",
			Value: config.Authorization,
		})
	}

	def := func(values ...string) string {
		for _, v := range values {
			if v != "" {
				return v
			}
		}

		return ""
	}

	return &core.Probe{
		Handler: core.Handler{
			HTTPGet: &core.HTTPGetAction{
				Path:        config.LocalPath,
				Port:        intstr.FromString(def(config.PortName, shared.ServerPortName)),
				Scheme:      scheme,
				HTTPHeaders: headers,
			},
		},
		InitialDelaySeconds: defaultInt32(config.InitialDelaySeconds, 900), // Wait 15min before first probe
		TimeoutSeconds:      defaultInt32(config.TimeoutSeconds, 2),        // Timeout of each probe is 2s
		PeriodSeconds:       defaultInt32(config.PeriodSeconds, 60),        // Interval between probes is 10s
		SuccessThreshold:    defaultInt32(config.SuccessThreshold, 1),      // Single probe is enough to indicate success
		FailureThreshold:    defaultInt32(config.FailureThreshold, 10),     // Need 10 failed probes to consider a failed state
	}
}

type CMDProbeConfig struct {
	// Command to be executed
	Command []string
	// Number of seconds after the container has started before liveness probes are initiated (defaults to 30)
	InitialDelaySeconds int32
	// Number of seconds after which the probe times out (defaults to 2).
	TimeoutSeconds int32
	// How often (in seconds) to perform the probe (defaults to 10).
	PeriodSeconds int32
	// Minimum consecutive successes for the probe to be considered successful after having failed (defaults to 1).
	SuccessThreshold int32
	// Minimum consecutive failures for the probe to be considered failed after having succeeded (defaults to 3).
	FailureThreshold int32
}

func (config *CMDProbeConfig) SetSpec(spec *api.ServerGroupProbeSpec) {
	config.InitialDelaySeconds = spec.GetInitialDelaySeconds(config.InitialDelaySeconds)
	config.TimeoutSeconds = spec.GetTimeoutSeconds(config.TimeoutSeconds)
	config.PeriodSeconds = spec.GetPeriodSeconds(config.PeriodSeconds)
	config.SuccessThreshold = spec.GetSuccessThreshold(config.SuccessThreshold)
	config.FailureThreshold = spec.GetFailureThreshold(config.FailureThreshold)
}

// Create creates a probe from given config
func (config CMDProbeConfig) Create() *core.Probe {
	return &core.Probe{
		Handler: core.Handler{
			Exec: &core.ExecAction{
				Command: config.Command,
			},
		},
		InitialDelaySeconds: defaultInt32(config.InitialDelaySeconds, 900), // Wait 15min before first probe
		TimeoutSeconds:      defaultInt32(config.TimeoutSeconds, 2),        // Timeout of each probe is 2s
		PeriodSeconds:       defaultInt32(config.PeriodSeconds, 60),        // Interval between probes is 10s
		SuccessThreshold:    defaultInt32(config.SuccessThreshold, 1),      // Single probe is enough to indicate success
		FailureThreshold:    defaultInt32(config.FailureThreshold, 10),     // Need 10 failed probes to consider a failed state
	}
}

func defaultInt32(value, defaultValue int32) int32 {
	if value != 0 {
		return value
	}
	return defaultValue
}
