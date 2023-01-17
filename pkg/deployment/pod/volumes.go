//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package pod

import core "k8s.io/api/core/v1"

func NewVolumes() Volumes {
	return &volumes{
		volumes:      []core.Volume{},
		volumeMounts: []core.VolumeMount{},
	}
}

type Volumes interface {
	Append(b Builder, i Input)
	AddVolume(volumes ...core.Volume)
	AddVolumeMount(mounts ...core.VolumeMount)
	Volumes() []core.Volume
	VolumeMounts() []core.VolumeMount
}

var _ Volumes = &volumes{}

type volumes struct {
	volumes      []core.Volume
	volumeMounts []core.VolumeMount
}

func (v *volumes) Append(b Builder, i Input) {
	vols, mounts := b.Volumes(i)
	v.AddVolume(vols...)
	v.AddVolumeMount(mounts...)
}

func (v *volumes) AddVolume(volumes ...core.Volume) {
	if len(volumes) == 0 {
		return
	}

	v.volumes = append(v.volumes, volumes...)
}

func (v *volumes) AddVolumeMount(mounts ...core.VolumeMount) {
	if len(mounts) == 0 {
		return
	}

	v.volumeMounts = append(v.volumeMounts, mounts...)
}

func (v *volumes) Volumes() []core.Volume {
	return v.volumes
}

func (v *volumes) VolumeMounts() []core.VolumeMount {
	return v.volumeMounts
}
