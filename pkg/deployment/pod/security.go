//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

func Security() Builder {
	return security{}
}

type security struct{}

func (s security) Args(i Input) k8sutil.OptionPairs {
	opts := k8sutil.CreateOptionPairs()

	if features.EphemeralVolumes().Enabled() {
		opts.Add("--temp.path", "/ephemeral/app")
		opts.Add("--javascript.app-path", "/ephemeral/tmp")
	}

	return opts
}

func (s security) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	var v []core.Volume
	var vm []core.VolumeMount

	if features.EphemeralVolumes().Enabled() {
		// Add Volumes
		{
			v = append(v, core.Volume{
				Name: shared.FoxxAppEphemeralVolumeName,
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{
						SizeLimit: i.GroupSpec.EphemeralVolumes.GetAppsSize(),
					},
				},
			})
		}

		{
			v = append(v, core.Volume{
				Name: shared.TMPEphemeralVolumeName,
				VolumeSource: core.VolumeSource{
					EmptyDir: &core.EmptyDirVolumeSource{
						SizeLimit: i.GroupSpec.EphemeralVolumes.GetTempSize(),
					},
				},
			})
		}

		// Mount volumes
		vm = append(vm, core.VolumeMount{
			Name:      shared.FoxxAppEphemeralVolumeName,
			MountPath: "/ephemeral/app",
		})
		vm = append(vm, core.VolumeMount{
			Name:      shared.TMPEphemeralVolumeName,
			MountPath: "/ephemeral/tmp",
		})
	}

	return v, vm
}

func (s security) Envs(i Input) []core.EnvVar {
	return nil
}

func (s security) Verify(i Input, cachedStatus interfaces.Inspector) error {
	return nil
}
