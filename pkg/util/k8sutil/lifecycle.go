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
// Author Tomasz Mielech <tomasz@arangodb.com>
//

package k8sutil

import (
	"os"
	"path/filepath"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"

	core "k8s.io/api/core/v1"
)

const (
	initLifecycleContainerName = "init-lifecycle"
	LifecycleVolumeMountDir    = "/lifecycle/tools"
	lifecycleVolumeName        = "lifecycle"
)

// InitLifecycleContainer creates an init-container to copy the lifecycle binary to a shared volume.
func InitLifecycleContainer(image string, resources *core.ResourceRequirements, securityContext *core.SecurityContext) (core.Container, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return core.Container{}, maskAny(err)
	}
	c := core.Container{
		Name:    initLifecycleContainerName,
		Image:   image,
		Command: append([]string{binaryPath}, "lifecycle", "copy", "--target", LifecycleVolumeMountDir),
		VolumeMounts: []core.VolumeMount{
			LifecycleVolumeMount(),
		},
		ImagePullPolicy: core.PullIfNotPresent,
		SecurityContext: securityContext,
	}

	if resources != nil {
		c.Resources = ExtractPodResourceRequirement(*resources)
	}
	return c, nil
}

// NewLifecycle creates a lifecycle structure with preStop handler.
func NewLifecycle() (*core.Lifecycle, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return nil, maskAny(err)
	}
	exePath := filepath.Join(LifecycleVolumeMountDir, filepath.Base(binaryPath))
	lifecycle := &core.Lifecycle{
		PreStop: &core.Handler{
			Exec: &core.ExecAction{
				Command: append([]string{exePath}, "lifecycle", "preStop"),
			},
		},
	}

	return lifecycle, nil
}

func GetLifecycleEnv() []core.EnvVar {
	return []core.EnvVar{
		CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
		CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
		CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
		CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
	}
}

// LifecycleVolumeMount creates a volume mount structure for shared lifecycle emptyDir.
func LifecycleVolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:      lifecycleVolumeName,
		MountPath: LifecycleVolumeMountDir,
	}
}

// LifecycleVolume creates a volume mount structure for shared lifecycle emptyDir.
func LifecycleVolume() core.Volume {
	return core.Volume{
		Name: lifecycleVolumeName,
		VolumeSource: core.VolumeSource{
			EmptyDir: &core.EmptyDirVolumeSource{},
		},
	}
}
