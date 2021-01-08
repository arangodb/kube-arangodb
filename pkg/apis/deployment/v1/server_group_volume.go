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
// Author Adam Janikowski
//

package v1

import (
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedv1 "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	core "k8s.io/api/core/v1"
)

var (
	restrictedVolumeNames = []string{
		k8sutil.ArangodVolumeName,
		k8sutil.TlsKeyfileVolumeName,
		k8sutil.RocksdbEncryptionVolumeName,
		k8sutil.ExporterJWTVolumeName,
		k8sutil.ClusterJWTSecretVolumeName,
		"lifecycle",
	}
)

// IsRestrictedVolumeName check of volume name is restricted, for example for originally mounted volumes
func IsRestrictedVolumeName(name string) bool {
	for _, restrictedVolumeName := range restrictedVolumeNames {
		if restrictedVolumeName == name {
			return true
		}
	}

	return false
}

// ServerGroupSpecVolumes definition of volume list which need to be mounted to Pod
type ServerGroupSpecVolumes []ServerGroupSpecVolume

// Validate if ServerGroupSpec volumes are valid and does not collide
func (s ServerGroupSpecVolumes) Validate() error {
	var validationErrors []error

	mappedVolumes := map[string]int{}

	for id, volume := range s {
		if i, ok := mappedVolumes[volume.Name]; ok {
			mappedVolumes[volume.Name] = i + 1
		} else {
			mappedVolumes[volume.Name] = 1
		}

		if err := volume.Validate(); err != nil {
			validationErrors = append(validationErrors, shared.PrefixResourceErrors(fmt.Sprintf("%d", id), err))
		}
	}

	for volumeName, count := range mappedVolumes {
		if IsRestrictedVolumeName(volumeName) {
			validationErrors = append(validationErrors, errors.Newf("volume with name %s is restricted", volumeName))
		}

		if count == 1 {
			continue
		}

		validationErrors = append(validationErrors, errors.Newf("volume with name %s defined more than once: %d", volumeName, count))
	}

	return shared.WithErrors(validationErrors...)
}

// Volumes create volumes
func (s ServerGroupSpecVolumes) Volumes() []core.Volume {
	volumes := make([]core.Volume, len(s))

	for id, volume := range s {
		volumes[id] = volume.Volume()
	}

	return volumes
}

// ServerGroupSpecVolume definition of volume which need to be mounted to Pod
type ServerGroupSpecVolume struct {
	// Name of volume
	Name string `json:"name"`

	// Secret which should be mounted into pod
	Secret *ServerGroupSpecVolumeSecret `json:"secret,omitempty"`

	// ConfigMap which should be mounted into pod
	ConfigMap *ServerGroupSpecVolumeConfigMap `json:"configMap,omitempty"`

	// EmptyDir
	EmptyDir *ServerGroupSpecVolumeEmptyDir `json:"emptyDir,omitempty"`
}

// Validate if ServerGroupSpec volume is valid
func (s *ServerGroupSpecVolume) Validate() error {
	if s == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceErrors("name", sharedv1.AsKubernetesResourceName(&s.Name).Validate()),
		shared.PrefixResourceErrors("secret", s.Secret.Validate()),
		shared.PrefixResourceErrors("configMap", s.ConfigMap.Validate()),
		shared.PrefixResourceErrors("emptyDir", s.EmptyDir.Validate()),
		s.validate(),
	)
}

// Volume create Pod Volume object
func (s ServerGroupSpecVolume) Volume() core.Volume {
	return core.Volume{
		Name: s.Name,
		VolumeSource: core.VolumeSource{
			ConfigMap: (*core.ConfigMapVolumeSource)(s.ConfigMap),
			Secret:    (*core.SecretVolumeSource)(s.Secret),
			EmptyDir:  (*core.EmptyDirVolumeSource)(s.EmptyDir),
		},
	}
}

func (s *ServerGroupSpecVolume) validate() error {
	count := s.notNilFields()

	if count == 0 {
		return errors.Newf("at least one option need to be defined: secret, configMap or emptyDir")
	}

	if count > 1 {
		return errors.Newf("only one option can be defined: secret, configMap or emptyDir")
	}

	return nil
}

func (s *ServerGroupSpecVolume) notNilFields() int {
	i := 0

	if s.ConfigMap != nil {
		i++
	}

	if s.Secret != nil {
		i++
	}

	if s.EmptyDir != nil {
		i++
	}

	return i
}

type ServerGroupSpecVolumeSecret core.SecretVolumeSource

func (s *ServerGroupSpecVolumeSecret) Validate() error {
	if s == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("secretName", sharedv1.AsKubernetesResourceName(&s.SecretName).Validate()),
	)
}

type ServerGroupSpecVolumeConfigMap core.ConfigMapVolumeSource

func (s *ServerGroupSpecVolumeConfigMap) Validate() error {
	if s == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("name", sharedv1.AsKubernetesResourceName(&s.Name).Validate()),
	)
}

type ServerGroupSpecVolumeEmptyDir core.EmptyDirVolumeSource

func (s *ServerGroupSpecVolumeEmptyDir) Validate() error {
	return nil
}
