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
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package resources

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/probes"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"

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

func createExporterArgs(spec api.DeploymentSpec) []string {
	tokenpath := filepath.Join(k8sutil.ExporterJWTVolumeMountDir, constants.SecretKeyToken)
	options := make([]k8sutil.OptionPair, 0, 64)
	scheme := "http"
	if spec.IsSecure() {
		scheme = "https"
	}
	options = append(options,
		k8sutil.OptionPair{Key: "--arangodb.jwt-file", Value: tokenpath},
		k8sutil.OptionPair{Key: "--arangodb.endpoint", Value: scheme + "://localhost:" + strconv.Itoa(k8sutil.ArangoPort)},
	)
	keyPath := filepath.Join(k8sutil.TLSKeyfileVolumeMountDir, constants.SecretTLSKeyfile)
	if spec.IsSecure() {
		options = append(options,
			k8sutil.OptionPair{Key: "--ssl.keyfile", Value: keyPath},
		)
	}

	if port := spec.Metrics.GetPort(); port != k8sutil.ArangoExporterPort {
		options = append(options,
			k8sutil.OptionPair{Key: "--server.address", Value: fmt.Sprintf(":%d", port)},
		)
	}

	args := make([]string, 0, 2+len(options))
	sort.Slice(options, func(i, j int) bool {
		return options[i].CompareTo(options[j]) < 0
	})
	for _, o := range options {
		args = append(args, o.Key+"="+o.Value)
	}

	return args
}

func createExporterLivenessProbe(isSecure bool) *probes.HTTPProbeConfig {
	probeCfg := &probes.HTTPProbeConfig{
		LocalPath: "/",
		Port:      k8sutil.ArangoExporterPort,
		Secure:    isSecure,
	}

	return probeCfg
}
