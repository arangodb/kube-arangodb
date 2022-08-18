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

package v1

import (
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedv1 "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

var (
	restrictedVolumeNames = []string{
		shared.ArangodVolumeName,
		shared.TlsKeyfileVolumeName,
		shared.RocksdbEncryptionVolumeName,
		shared.ExporterJWTVolumeName,
		shared.ClusterJWTSecretVolumeName,
		shared.LifecycleVolumeName,
		shared.FoxxAppEphemeralVolumeName,
		shared.TMPEphemeralVolumeName,
		shared.ArangoDTimezoneVolumeName,
	}
)

const (
	ServerGroupSpecVolumeRenderParamDeploymentName      = "DEPLOYMENT_NAME"
	ServerGroupSpecVolumeRenderParamDeploymentNamespace = "DEPLOYMENT_NAMESPACE"
	ServerGroupSpecVolumeRenderParamMemberID            = "MEMBER_ID"
	ServerGroupSpecVolumeRenderParamMemberRoleAbbr      = "ROLE_ABBR"
	ServerGroupSpecVolumeRenderParamMemberRole          = "ROLE"
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

// RenderVolumes render volumes
func (s ServerGroupSpecVolumes) RenderVolumes(depl meta.Object, group ServerGroup, member MemberStatus) []core.Volume {
	volumes := make([]core.Volume, len(s))

	for id, volume := range s {
		volumes[id] = volume.RenderVolume(depl, group, member)
	}

	return volumes
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

	// HostPath
	HostPath *ServerGroupSpecVolumeHostPath `json:"hostPath,omitempty"`

	// PersistentVolumeClaim
	PersistentVolumeClaim *ServerGroupSpecVolumePersistentVolumeClaim `json:"persistentVolumeClaim,omitempty"`
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
		shared.PrefixResourceErrors("hostPath", s.HostPath.Validate()),
		shared.PrefixResourceErrors("persistentVolumeClaim", s.PersistentVolumeClaim.Validate()),
		s.validate(),
	)
}

// RenderVolume create Pod Volume object with dynamic names
func (s ServerGroupSpecVolume) RenderVolume(depl meta.Object, group ServerGroup, member MemberStatus) core.Volume {
	return core.Volume{
		Name: s.Name,
		VolumeSource: core.VolumeSource{
			ConfigMap:             s.ConfigMap.render(depl, group, member),
			Secret:                s.Secret.render(depl, group, member),
			EmptyDir:              s.EmptyDir.render(),
			HostPath:              s.HostPath.render(depl, group, member),
			PersistentVolumeClaim: s.PersistentVolumeClaim.render(depl, group, member),
		},
	}
}

// Volume create Pod Volume object
func (s ServerGroupSpecVolume) Volume() core.Volume {
	return core.Volume{
		Name: s.Name,
		VolumeSource: core.VolumeSource{
			ConfigMap:             (*core.ConfigMapVolumeSource)(s.ConfigMap),
			Secret:                (*core.SecretVolumeSource)(s.Secret),
			EmptyDir:              (*core.EmptyDirVolumeSource)(s.EmptyDir),
			HostPath:              (*core.HostPathVolumeSource)(s.HostPath),
			PersistentVolumeClaim: (*core.PersistentVolumeClaimVolumeSource)(s.PersistentVolumeClaim),
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

	if s.HostPath != nil {
		i++
	}

	if s.PersistentVolumeClaim != nil {
		i++
	}

	return i
}

func renderVolumeResourceName(in string, depl meta.Object, group ServerGroup, member MemberStatus) string {
	return shared.RenderResourceName(in, map[string]string{
		ServerGroupSpecVolumeRenderParamDeploymentName:      depl.GetName(),
		ServerGroupSpecVolumeRenderParamDeploymentNamespace: depl.GetNamespace(),
		ServerGroupSpecVolumeRenderParamMemberID:            shared.StripArangodPrefix(member.ID),
		ServerGroupSpecVolumeRenderParamMemberRole:          group.AsRole(),
		ServerGroupSpecVolumeRenderParamMemberRoleAbbr:      group.AsRoleAbbreviated(),
	})
}

type ServerGroupSpecVolumeSecret core.SecretVolumeSource

func (s *ServerGroupSpecVolumeSecret) Validate() error {
	q := s.render(&ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "render",
			Namespace: "render",
		},
	}, ServerGroupSingle, MemberStatus{
		ID: "render",
	})

	if q == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("secretName", sharedv1.AsKubernetesResourceName(&q.SecretName).Validate()),
	)
}

func (s *ServerGroupSpecVolumeSecret) render(depl meta.Object, group ServerGroup, member MemberStatus) *core.SecretVolumeSource {
	if s == nil {
		return nil
	}

	var obj = core.SecretVolumeSource(*s)

	obj.SecretName = renderVolumeResourceName(obj.SecretName, depl, group, member)

	return &obj
}

type ServerGroupSpecVolumeConfigMap core.ConfigMapVolumeSource

func (s *ServerGroupSpecVolumeConfigMap) Validate() error {
	q := s.render(&ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "render",
			Namespace: "render",
		},
	}, ServerGroupSingle, MemberStatus{
		ID: "render",
	})

	if q == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("name", sharedv1.AsKubernetesResourceName(&q.Name).Validate()),
	)
}

func (s *ServerGroupSpecVolumeConfigMap) render(depl meta.Object, group ServerGroup, member MemberStatus) *core.ConfigMapVolumeSource {
	if s == nil {
		return nil
	}

	var obj = core.ConfigMapVolumeSource(*s)

	obj.Name = renderVolumeResourceName(obj.Name, depl, group, member)

	return &obj
}

type ServerGroupSpecVolumeEmptyDir core.EmptyDirVolumeSource

func (s *ServerGroupSpecVolumeEmptyDir) Validate() error {
	return nil
}

func (s *ServerGroupSpecVolumeEmptyDir) render() *core.EmptyDirVolumeSource {
	if s == nil {
		return nil
	}

	return (*core.EmptyDirVolumeSource)(s)
}

type ServerGroupSpecVolumeHostPath core.HostPathVolumeSource

func (s *ServerGroupSpecVolumeHostPath) Validate() error {
	return nil
}

func (s *ServerGroupSpecVolumeHostPath) render(depl meta.Object, group ServerGroup, member MemberStatus) *core.HostPathVolumeSource {
	if s == nil {
		return nil
	}

	var obj = core.HostPathVolumeSource(*s)

	obj.Path = renderVolumeResourceName(obj.Path, depl, group, member)

	return &obj
}

type ServerGroupSpecVolumePersistentVolumeClaim core.PersistentVolumeClaimVolumeSource

func (s *ServerGroupSpecVolumePersistentVolumeClaim) Validate() error {
	q := s.render(&ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "render",
			Namespace: "render",
		},
	}, ServerGroupSingle, MemberStatus{
		ID: "render",
	})

	if q == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("claimName", sharedv1.AsKubernetesResourceName(&q.ClaimName).Validate()),
	)
}

func (s *ServerGroupSpecVolumePersistentVolumeClaim) render(depl meta.Object, group ServerGroup, member MemberStatus) *core.PersistentVolumeClaimVolumeSource {
	if s == nil {
		return nil
	}

	var obj = core.PersistentVolumeClaimVolumeSource(*s)

	obj.ClaimName = renderVolumeResourceName(obj.ClaimName, depl, group, member)

	return &obj
}
