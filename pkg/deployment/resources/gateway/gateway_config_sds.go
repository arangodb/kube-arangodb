//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package gateway

import (
	"fmt"

	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

// SDSFileName returns the filesystem SDS definition file name (also the ConfigMap key) for the
// given SDS secret name.
func SDSFileName(name string) string {
	return fmt.Sprintf("%s.yaml", name)
}

// tlsCertificates returns every certificate served by the gateway that is delivered to Envoy via
// SDS: the default (internal) serving certificate plus one per SNI mapping.
func (c Config) tlsCertificates() []*ConfigTLS {
	var r []*ConfigTLS

	if c.DefaultTLS != nil {
		r = append(r, c.DefaultTLS)
	}

	for id := range c.SNI {
		r = append(r, &c.SNI[id].ConfigTLS)
	}

	return r
}

// RenderSDS renders the filesystem SDS secret definition files (keyed by file name) backing the
// listener's tls_certificate_sds_secret_configs. Each file is a discovery response holding a single
// Secret whose certificate/key are referenced by path; Envoy watches the corresponding mounted
// secret directory and hot-reloads the certificate on rotation.
func (c Config) RenderSDS() (map[string]string, error) {
	files := map[string]string{}

	for _, tls := range c.tlsCertificates() {
		if tls.Name == "" {
			continue
		}

		resp, err := DynamicConfigResponse(tls.RenderSDSSecret())
		if err != nil {
			return nil, err
		}

		data, err := ugrpc.MarshalYAML(resp, ugrpc.WithUseProtoNames(true))
		if err != nil {
			return nil, err
		}

		files[SDSFileName(tls.Name)] = string(data)
	}

	return files, nil
}
