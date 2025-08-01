//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package k8sutil

import (
	"os"
	"path/filepath"

	core "k8s.io/api/core/v1"

	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	kresources "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/resources"
)

const (
	initLifecycleContainerName = "init-lifecycle"
	LifecycleVolumeMountDir    = "/lifecycle/tools"
	lifecycleVolumeName        = "lifecycle"
)

var (
	binaryPath string
)

func init() {
	if b, err := os.Executable(); err != nil {
		panic(err.Error())
	} else {
		binaryPath = b
	}
}

func SetBinaryPath(path string) {
	binaryPath = path
}

func BinaryPath() string {
	return binaryPath
}

func LifecycleBinary() string {
	return filepath.Join(LifecycleVolumeMountDir, filepath.Base(binaryPath))
}

// InitLifecycleContainer creates an init-container to copy the lifecycle binary to a shared volume.
func InitLifecycleContainer(image, binaryPath string, resources *core.ResourceRequirements, securityContext *core.SecurityContext) (core.Container, error) {
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
		c.Resources = kresources.ExtractPodAcceptedResourceRequirement(*resources)
	}
	return c, nil
}

// NewLifecycleFinalizers creates a lifecycle structure with preStop handler which wait for finalizers to be removed.
func NewLifecycleFinalizers() (*core.Lifecycle, error) {
	return NewLifecycle("finalizers")
}

// NewLifecycleFinalizersWithBinary creates a lifecycle structure with preStop handler which wait for finalizers to be removed using specific binary path.
func NewLifecycleFinalizersWithBinary(exePath string) (*core.Lifecycle, error) {
	return NewLifecycleWithBinary(exePath, "finalizers")
}

// NewLifecyclePort creates a lifecycle structure with preStop handler which wait for port to be closed.
func NewLifecyclePort() (*core.Lifecycle, error) {
	return NewLifecycle("port")
}

// NewLifecycle creates a lifecycle structure with preStop handler.
func NewLifecycle(t string) (*core.Lifecycle, error) {
	return NewLifecycleWithBinary(LifecycleBinary(), t)
}

// NewLifecycleWithBinary creates a lifecycle structure with preStop handler using specific binary path.
func NewLifecycleWithBinary(exePath string, t string) (*core.Lifecycle, error) {
	lifecycle := &core.Lifecycle{
		PreStop: &core.LifecycleHandler{
			Exec: &core.ExecAction{
				Command: append([]string{exePath}, "lifecycle", "preStop", t),
			},
		},
	}

	return lifecycle, nil
}

func AppendLifecycleEnv(in []core.EnvVar) []core.EnvVar {
	for _, e := range GetLifecycleEnv() {
		if !EnvExists(in, e.Name) {
			in = append(in, e)
		}
	}
	return in
}

func EnvExists(a []core.EnvVar, name string) bool {
	for _, q := range a {
		if q.Name == name {
			return true
		}
	}

	return false
}

func GetLifecycleEnv() []core.EnvVar {
	return []core.EnvVar{
		CreateEnvFieldPath(utilConstants.EnvOperatorPodName, "metadata.name"),
		CreateEnvFieldPath(utilConstants.EnvOperatorPodNamespace, "metadata.namespace"),
		CreateEnvFieldPath(utilConstants.EnvOperatorNodeName, "spec.nodeName"),
		CreateEnvFieldPath(utilConstants.EnvOperatorNodeNameArango, "spec.nodeName"),
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
