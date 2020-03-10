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
// Author Ewout Prangsma
//

package k8sutil

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// HTTPProbeConfig contains settings for creating a liveness/readiness probe.
type HTTPProbeConfig struct {
	// Local path to GET
	LocalPath string // `e.g. /_api/version`
	// Secure connection?
	Secure bool
	// Value for an Authorization header (can be empty)
	Authorization string
	// Port to inspect (defaults to ArangoPort)
	Port int
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

// Create creates a probe from given config
func (config HTTPProbeConfig) Create() *v1.Probe {
	scheme := v1.URISchemeHTTP
	if config.Secure {
		scheme = v1.URISchemeHTTPS
	}
	var headers []v1.HTTPHeader
	if config.Authorization != "" {
		headers = append(headers, v1.HTTPHeader{
			Name:  "Authorization",
			Value: config.Authorization,
		})
	}
	def := func(value, defaultValue int32) int32 {
		if value != 0 {
			return value
		}
		return defaultValue
	}
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path:        config.LocalPath,
				Port:        intstr.FromInt(int(def(int32(config.Port), ArangoPort))),
				Scheme:      scheme,
				HTTPHeaders: headers,
			},
		},
		InitialDelaySeconds: def(config.InitialDelaySeconds, 15*60), // Wait 15min before first probe
		TimeoutSeconds:      def(config.TimeoutSeconds, 2),          // Timeout of each probe is 2s
		PeriodSeconds:       def(config.PeriodSeconds, 60),          // Interval between probes is 10s
		SuccessThreshold:    def(config.SuccessThreshold, 1),        // Single probe is enough to indicate success
		FailureThreshold:    def(config.FailureThreshold, 10),       // Need 10 failed probes to consider a failed state
	}
}
