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
func InitLifecycleContainer(image string) (v1.Container, error) {
	binaryPath, err := os.Executable()
	if err != nil {
		return v1.Container{}, maskAny(err)
	}
	c := v1.Container{
		Command:         append([]string{binaryPath}, "lifecycle", "copy", "--target", lifecycleVolumeMountDir),
		Name:            initLifecycleContainerName,
		Image:           image,
		ImagePullPolicy: v1.PullIfNotPresent,
		VolumeMounts: []v1.VolumeMount{
			LifecycleVolumeMounts(),
		},
		SecurityContext: SecurityContextWithoutCapabilities(),
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

// LifecycleVolumeMounts creates a volume mount structure for shared lifecycle emptyDir.
func LifecycleVolumeMounts() v1.VolumeMount {
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
