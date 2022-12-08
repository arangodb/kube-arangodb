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

package resources

import (
	"path/filepath"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/probes"
)

func createInternalExporterArgs(spec api.DeploymentSpec, groupSpec api.ServerGroupSpec, version driver.Version) []string {
	tokenpath := filepath.Join(shared.ExporterJWTVolumeMountDir, constants.SecretKeyToken)
	options := k8sutil.CreateOptionPairs(64)

	if spec.Authentication.IsAuthenticated() {
		options.Add("--arangodb.jwt-file", tokenpath)
	}

	path := getArangoExporterInternalEndpoint(version)

	if port := groupSpec.InternalPort; port == nil {
		scheme := "http"
		if spec.IsSecure() {
			scheme = "https"
		}
		options.Addf("--arangodb.endpoint", "%s://localhost:%d%s", scheme, groupSpec.GetPort(), path)
	} else {
		options.Addf("--arangodb.endpoint", "http://localhost:%d%s", *port, path)
	}

	keyPath := filepath.Join(shared.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
	if spec.IsSecure() && spec.Metrics.IsTLS() {
		options.Add("--ssl.keyfile", keyPath)
	}

	var port uint16 = shared.ArangoExporterPort

	if p := spec.Metrics.Port; p != nil {
		port = *p
	}

	if p := groupSpec.ExporterPort; p != nil {
		port = *p
	}

	if port != shared.ArangoExporterPort {
		options.Addf("--server.address", ":%d", port)
	}

	return options.Sort().AsArgs()
}

func getArangoExporterInternalEndpoint(version driver.Version) string {
	path := shared.ArangoExporterInternalEndpoint
	if version.CompareTo("3.8.0") >= 0 {
		path = shared.ArangoExporterInternalEndpointV2
	}
	return path
}

func createExporterLivenessProbe(isSecure bool) *probes.HTTPProbeConfig {
	probeCfg := &probes.HTTPProbeConfig{
		LocalPath: "/",
		PortName:  shared.ExporterPortName,
		Secure:    isSecure,
	}

	return probeCfg
}
