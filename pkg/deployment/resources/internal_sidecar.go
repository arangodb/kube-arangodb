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

package resources

import (
	"path/filepath"

	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/probes"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

func createInternalSidecarArgs(spec api.DeploymentSpec, groupSpec api.ServerGroupSpec) []string {
	options := k8sutil.CreateOptionPairs(64)

	if spec.Authentication.IsAuthenticated() {
		options.Add("--sidecar.auth", shared.ExporterJWTVolumeMountDir)
	}

	if port := groupSpec.InternalPort; port == nil {
		scheme := "http"
		if spec.IsSecure() {
			scheme = "https"
		}
		options.Addf("--arangodb.endpoint", "%s://localhost:%d", scheme, groupSpec.GetPort())
	} else {
		options.Addf("--arangodb.endpoint", "http://localhost:%d", *port)
	}

	if spec.TLS.IsSecure() {
		options.Add("--sidecar.keyfile", filepath.Join(shared.TLSKeyfileVolumeMountDir, utilConstants.SecretTLSKeyfile))
	}

	return options.AsArgs()
}

// ArangodbInternalSidecarContainer creates sidecar container based on internal sidecar
func ArangodbInternalSidecarContainer(image string, args []string,
	res core.ResourceRequirements, spec api.DeploymentSpec, groupSpec api.ServerGroupSpec) (core.Container, error) {
	exePath := k8sutil.LifecycleBinary()

	c := core.Container{
		Name:    shared.IntegrationContainerName,
		Image:   image,
		Command: append([]string{exePath, "sidecar"}, args...),
		Ports: []core.ContainerPort{
			{
				Name:          shared.InternalSidecarContainerPortGRPCName,
				ContainerPort: int32(shared.InternalSidecarContainerPortGRPC),
				Protocol:      core.ProtocolTCP,
			},
			{
				Name:          shared.InternalSidecarContainerPortHTTPName,
				ContainerPort: int32(shared.InternalSidecarContainerPortHTTP),
				Protocol:      core.ProtocolTCP,
			},
			{
				Name:          shared.InternalSidecarContainerPortHealthName,
				ContainerPort: int32(shared.InternalSidecarContainerPortHealth),
				Protocol:      core.ProtocolTCP,
			},
		},
		Resources:       kresources.ExtractPodAcceptedResourceRequirement(res),
		SecurityContext: k8sutil.CreateSecurityContext(groupSpec.SecurityContext),
		ImagePullPolicy: core.PullIfNotPresent,
		VolumeMounts:    []core.VolumeMount{k8sutil.LifecycleVolumeMount()},
	}

	probe := probes.GRPCProbeConfig{Port: shared.InternalSidecarContainerPortHealth, Common: probes.Common{
		InitialDelaySeconds: util.NewType[int32](0),
		PeriodSeconds:       util.NewType[int32](5),
		TimeoutSeconds:      util.NewType[int32](2),
	}}

	c.LivenessProbe = probe.Create()
	c.ReadinessProbe = probe.Create()

	return c, nil
}
