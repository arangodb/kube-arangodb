//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
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

type ServerGroupSpecVolumeMount struct {
	// This must match the Name of a Volume.
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	// Mounted read-only if true, read-write otherwise (false or unspecified).
	// Defaults to false.
	ReadOnly bool `json:"readOnly,omitempty" protobuf:"varint,2,opt,name=readOnly"`

	// RecursiveReadOnly specifies whether read-only mounts should be handled
	// recursively.
	//
	// If ReadOnly is false, this field has no meaning and must be unspecified.
	//
	// If ReadOnly is true, and this field is set to Disabled, the mount is not made
	// recursively read-only.  If this field is set to IfPossible, the mount is made
	// recursively read-only, if it is supported by the container runtime.  If this
	// field is set to Enabled, the mount is made recursively read-only if it is
	// supported by the container runtime, otherwise the pod will not be started and
	// an error will be generated to indicate the reason.
	//
	// If this field is set to IfPossible or Enabled, MountPropagation must be set to
	// None (or be unspecified, which defaults to None).
	//
	// If this field is not specified, it is treated as an equivalent of Disabled.
	RecursiveReadOnly *core.RecursiveReadOnlyMode `json:"recursiveReadOnly,omitempty" protobuf:"bytes,7,opt,name=recursiveReadOnly,casttype=RecursiveReadOnlyMode"`

	// Path within the container at which the volume should be mounted.  Must
	// not contain ':'.
	MountPath string `json:"mountPath" protobuf:"bytes,3,opt,name=mountPath"`

	// Path within the volume from which the container's volume should be mounted.
	// Defaults to "" (volume's root).
	SubPath string `json:"subPath,omitempty" protobuf:"bytes,4,opt,name=subPath"`

	// mountPropagation determines how mounts are propagated from the host
	// to container and the other way around.
	// When not set, MountPropagationNone is used.
	// This field is beta in 1.10.
	// When RecursiveReadOnly is set to IfPossible or to Enabled, MountPropagation must be None or unspecified
	// (which defaults to None).
	MountPropagation *core.MountPropagationMode `json:"mountPropagation,omitempty" protobuf:"bytes,5,opt,name=mountPropagation,casttype=MountPropagationMode"`

	// Expanded path within the volume from which the container's volume should be mounted.
	// Behaves similarly to SubPath but environment variable references $(VAR_NAME) are expanded using the container's environment.
	// Defaults to "" (volume's root).
	// SubPathExpr and SubPath are mutually exclusive.
	SubPathExpr string `json:"subPathExpr,omitempty" protobuf:"bytes,6,opt,name=subPathExpr"`
}

func (s ServerGroupSpecVolumeMount) VolumeMount() core.VolumeMount {
	return core.VolumeMount{
		Name:              s.Name,
		ReadOnly:          s.ReadOnly,
		RecursiveReadOnly: s.RecursiveReadOnly,
		MountPath:         s.MountPath,
		SubPath:           s.SubPath,
		MountPropagation:  s.MountPropagation,
		SubPathExpr:       s.SubPathExpr,
	}
}

func (s *ServerGroupSpecVolumeMount) Validate() error {
	if s == nil {
		return nil
	}

	return shared.WithErrors(
		shared.PrefixResourceError("name", sharedApi.AsKubernetesResourceName(&s.Name).Validate()),
	)
}
