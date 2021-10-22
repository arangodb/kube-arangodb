//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package resources

import (
	"path/filepath"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/probes"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	v1 "k8s.io/api/core/v1"
)

// ArangodbExporterContainer creates metrics container
func ArangodbExporterContainer(image string, args []string, livenessProbe *probes.HTTPProbeConfig,
	resources v1.ResourceRequirements, securityContext *v1.SecurityContext,
	spec api.DeploymentSpec) v1.Container {

	c := v1.Container{
		Name:    k8sutil.ExporterContainerName,
		Image:   image,
		Command: append([]string{"/app/arangodb-exporter"}, args...),
		Ports: []v1.ContainerPort{
			{
				Name:          "exporter",
				ContainerPort: int32(spec.Metrics.GetPort()),
				Protocol:      v1.ProtocolTCP,
			},
		},
		Resources:       k8sutil.ExtractPodResourceRequirement(resources),
		ImagePullPolicy: v1.PullIfNotPresent,
		SecurityContext: securityContext,
	}

	if livenessProbe != nil {
		c.LivenessProbe = livenessProbe.Create()
	}

	return c
}

func createInternalExporterArgs(spec api.DeploymentSpec, groupSpec api.ServerGroupSpec, version driver.Version) []string {
	tokenpath := filepath.Join(k8sutil.ExporterJWTVolumeMountDir, constants.SecretKeyToken)
	options := k8sutil.CreateOptionPairs(64)

	options.Add("--arangodb.jwt-file", tokenpath)

	path := k8sutil.ArangoExporterInternalEndpoint
	if version.CompareTo("3.8.0") >= 0 {
		path = k8sutil.ArangoExporterInternalEndpointV2
	}

	if port := groupSpec.InternalPort; port == nil {
		scheme := "http"
		if spec.IsSecure() {
			scheme = "https"
		}
		options.Addf("--arangodb.endpoint", "%s://localhost:%d%s", scheme, k8sutil.ArangoPort, path)
	} else {
		options.Addf("--arangodb.endpoint", "http://localhost:%d%s", *port, path)
	}

	keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
	if spec.IsSecure() && spec.Metrics.IsTLS() {
		options.Add("--ssl.keyfile", keyPath)
	}

	if port := spec.Metrics.GetPort(); port != k8sutil.ArangoExporterPort {
		options.Addf("--server.address", ":%d", port)
	}

	return options.Sort().AsArgs()
}

func createExporterLivenessProbe(isSecure bool) *probes.HTTPProbeConfig {
	probeCfg := &probes.HTTPProbeConfig{
		LocalPath: "/",
		Port:      k8sutil.ArangoExporterPort,
		Secure:    isSecure,
	}

	return probeCfg
}
