//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
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

	Common `json:",inline"`
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
	return config.config(core.ProbeHandler{
		HTTPGet: &core.HTTPGetAction{
			Path:        config.LocalPath,
			Port:        intstr.FromString(def(config.PortName, shared.ServerPortName)),
			Scheme:      scheme,
			HTTPHeaders: headers,
		},
	})
}
