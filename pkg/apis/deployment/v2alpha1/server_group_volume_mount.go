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

package v2alpha1

import (
	"fmt"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedv1 "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
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
		shared.PrefixResourceError("name", sharedv1.AsKubernetesResourceName(&s.Name).Validate()),
	)
}
