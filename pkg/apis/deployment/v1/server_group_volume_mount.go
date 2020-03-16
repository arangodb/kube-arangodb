package v1

import (
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	core "k8s.io/api/core/v1"
)

type ServerGroupSpecVolumeMounts []ServerGroupSpecVolumeMount

func (s ServerGroupSpecVolumeMounts) VolumeMounts() []core.VolumeMount {
	mounts := make([]core.VolumeMount, len(s))

	for id, mount := range s {
		mounts[id] = mount.VolumeMount()
	}

	return mounts
}

func (s ServerGroupSpecVolumeMounts) Validate() error {
	if s == nil {
		return nil
	}

	validateErrors := make([]error, len(s))

	for id, mount := range s {
		validateErrors[id] = shared.PrefixResourceErrors(fmt.Sprintf("%d", id), mount.Validate())
	}

	return shared.WithErrors(validateErrors...)
}

type ServerGroupSpecVolumeMount core.VolumeMount

func (s ServerGroupSpecVolumeMount) VolumeMount() core.VolumeMount {
	return core.VolumeMount(s)
}

func (s *ServerGroupSpecVolumeMount) Validate() error {
	if s == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("name", shared.AsKubernetesResourceName(&s.Name).Validate()),
	)
}
