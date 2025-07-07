//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package sidecar

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	mountNameShutdownDebugPackage = "debug-package-mount"
)

type ExtensionShutdownV1Debug struct {
	Core *Core
}

func (i ExtensionShutdownV1Debug) Name() []string {
	return []string{"SHUTDOWN", "V1", "DEBUG"}
}

func (i ExtensionShutdownV1Debug) Validate() error {
	return nil
}

func (i ExtensionShutdownV1Debug) Envs() ([]core.EnvVar, error) {
	var envs = []core.EnvVar{
		{
			Name:  "INTEGRATION_SHUTDOWN_V1_DEBUG_ENABLED",
			Value: "true",
		},
		{
			Name:  "INTEGRATION_SHUTDOWN_V1_DEBUG_PATH",
			Value: "/debug",
		},
	}

	return i.Core.Envs(i, envs...), nil
}

func (i ExtensionShutdownV1Debug) GlobalEnvs() ([]core.EnvVar, error) {
	return nil, nil
}

func (i ExtensionShutdownV1Debug) Volumes() ([]core.Volume, []core.VolumeMount, error) {
	var volumeMounts []core.VolumeMount
	var volumes []core.Volume

	volumes = append(volumes, k8sutil.CreateVolumeEmptyDir(mountNameShutdownDebugPackage))
	volumeMounts = append(volumeMounts, core.VolumeMount{
		Name:      mountNameShutdownDebugPackage,
		MountPath: "/debug",
		ReadOnly:  true,
	})

	return volumes, volumeMounts, nil
}
