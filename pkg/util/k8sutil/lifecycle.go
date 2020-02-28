//
// DISCLAIMER
//
// Copyright 2019 ArangoDB GmbH, Cologne, Germany
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

	v1 "k8s.io/api/core/v1"
)

const (
	initLifecycleContainerName = "init-lifecycle"
	lifecycleVolumeMountDir    = "/lifecycle/tools"
	lifecycleVolumeName        = "lifecycle"
)

// InitLifecycleContainer creates an init-container to copy the lifecycle binary to a shared volume.
func InitLifecycleContainer(image string, resources *v1.ResourceRequirements, securityContext *v1.SecurityContext) (v1.Container, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return v1.Container{}, maskAny(err)
	}
	c := v1.Container{
		Name:    initLifecycleContainerName,
		Image:   image,
		Command: append([]string{binaryPath}, "lifecycle", "copy", "--target", lifecycleVolumeMountDir),
		VolumeMounts: []v1.VolumeMount{
			LifecycleVolumeMount(),
		},
		ImagePullPolicy: v1.PullIfNotPresent,
		SecurityContext: securityContext,
	}

	if resources != nil {
		c.Resources = ExtractPodResourceRequirement(*resources)
	}
	return c, nil
}

// NewLifecycle creates a lifecycle structure with preStop handler.
func NewLifecycle() (*v1.Lifecycle, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return nil, maskAny(err)
	}
	exePath := filepath.Join(lifecycleVolumeMountDir, filepath.Base(binaryPath))
	lifecycle := &v1.Lifecycle{
		PreStop: &v1.Handler{
			Exec: &v1.ExecAction{
				Command: append([]string{exePath}, "lifecycle", "preStop"),
			},
		},
	}

	return lifecycle, nil
}

func GetLifecycleEnv() []v1.EnvVar {
	return []v1.EnvVar{
		CreateEnvFieldPath(constants.EnvOperatorPodName, "metadata.name"),
		CreateEnvFieldPath(constants.EnvOperatorPodNamespace, "metadata.namespace"),
		CreateEnvFieldPath(constants.EnvOperatorNodeName, "spec.nodeName"),
		CreateEnvFieldPath(constants.EnvOperatorNodeNameArango, "spec.nodeName"),
	}
}

// LifecycleVolumeMount creates a volume mount structure for shared lifecycle emptyDir.
func LifecycleVolumeMount() v1.VolumeMount {
	return v1.VolumeMount{
		Name:      lifecycleVolumeName,
		MountPath: lifecycleVolumeMountDir,
	}
}

// LifecycleVolume creates a volume mount structure for shared lifecycle emptyDir.
func LifecycleVolume() v1.Volume {
	return v1.Volume{
		Name: lifecycleVolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}
