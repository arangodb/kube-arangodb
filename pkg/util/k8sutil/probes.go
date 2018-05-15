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

package k8sutil

import (
	"k8s.io/api/core/v1"
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
	port := config.Port
	if port == 0 {
		port = ArangoPort
	}
	return &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path:        config.LocalPath,
				Port:        intstr.FromInt(port),
				Scheme:      scheme,
				HTTPHeaders: headers,
			},
		},
		InitialDelaySeconds: 30, // Wait 30s before first probe
		TimeoutSeconds:      2,  // Timeout of each probe is 2s
		PeriodSeconds:       10, // Interval between probes is 10s
		SuccessThreshold:    1,  // Single probe is enough to indicate success
		FailureThreshold:    3,  // Need 3 failed probes to consider a failed state
	}
}
