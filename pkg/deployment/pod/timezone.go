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
	"fmt"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/interfaces"
)

const (
	TimezoneNameKey string = "name"
	TimezoneDataKey string = "data"
	TimezoneTZKey   string = "timezone"
)

func TimezoneSecret(name string) string {
	return fmt.Sprintf("%s-timezone", name)
}

func Timezone() Builder {
	return timezone{}
}

type timezone struct {
}

func (t timezone) Args(i Input) k8sutil.OptionPairs {
	return nil
}

func (t timezone) Volumes(i Input) ([]core.Volume, []core.VolumeMount) {
	if !features.Timezone().Enabled() {
		return nil, nil
	}

	return []core.Volume{
			{
				Name: shared.ArangoDTimezoneVolumeName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: TimezoneSecret(i.ApiObject.GetName()),
					},
				},
			},
		}, []core.VolumeMount{
			{
				Name:      shared.ArangoDTimezoneVolumeName,
				ReadOnly:  true,
				MountPath: "/etc/localtime",
				SubPath:   TimezoneDataKey,
			},
			{
				Name:      shared.ArangoDTimezoneVolumeName,
				ReadOnly:  true,
				MountPath: "/etc/timezone",
				SubPath:   TimezoneTZKey,
			},
		}
}

func (t timezone) Envs(i Input) []core.EnvVar {
	return nil
}

func (t timezone) Verify(i Input, cachedStatus interfaces.Inspector) error {
	return nil
}
