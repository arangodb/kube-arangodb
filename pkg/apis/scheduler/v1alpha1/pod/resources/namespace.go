//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1/interfaces"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

var _ interfaces.Pod[Namespace] = &Namespace{}

type Namespace struct {
	// HostNetwork requests Host network for this pod. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// +doc/default: false
	HostNetwork *bool `json:"hostNetwork,omitempty" protobuf:"varint,11,opt,name=hostNetwork"`
	// HostPID define to use the host's pid namespace.
	// +doc/default: false
	HostPID *bool `json:"hostPID,omitempty" protobuf:"varint,12,opt,name=hostPID"`
	// HostIPC defines to use the host's ipc namespace.
	// +doc/default: false
	HostIPC *bool `json:"hostIPC,omitempty" protobuf:"varint,13,opt,name=hostIPC"`
	// ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
	// When this is set containers will be able to view and signal processes from other containers
	// in the same pod, and the first process in each container will not be assigned PID 1.
	// HostPID and ShareProcessNamespace cannot both be set.
	// +doc/default: false
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty" protobuf:"varint,27,opt,name=shareProcessNamespace"`
}

func (n *Namespace) Apply(template *core.PodTemplateSpec) error {
	if n == nil {
		return nil
	}

	template.Spec.HostNetwork = util.WithDefault(n.HostNetwork)
	template.Spec.HostPID = util.WithDefault(n.HostPID)
	template.Spec.HostIPC = util.WithDefault(n.HostIPC)
	if v := n.ShareProcessNamespace; v != nil {
		template.Spec.ShareProcessNamespace = util.NewType(*v)
	} else {
		template.Spec.ShareProcessNamespace = nil
	}

	return nil
}

func (n *Namespace) GetHostNetwork() *bool {
	if n == nil || n.HostNetwork == nil {
		return nil
	}

	return n.HostNetwork
}

func (n *Namespace) GetHostPID() *bool {
	if n == nil || n.HostPID == nil {
		return nil
	}

	return n.HostPID
}

func (n *Namespace) GetHostIPC() *bool {
	if n == nil || n.HostIPC == nil {
		return nil
	}

	return n.HostIPC
}

func (n *Namespace) GetShareProcessNamespace() *bool {
	if n == nil {
		return nil
	}

	return n.ShareProcessNamespace
}

func (n *Namespace) With(other *Namespace) *Namespace {
	if n == nil && other == nil {
		return nil
	}

	if n == nil {
		return other.DeepCopy()
	}

	if other == nil {
		return n.DeepCopy()
	}

	return &Namespace{
		HostNetwork:           util.First(other.HostNetwork, n.HostNetwork),
		HostPID:               util.First(other.HostPID, n.HostPID),
		HostIPC:               util.First(other.HostIPC, n.HostIPC),
		ShareProcessNamespace: util.First(other.ShareProcessNamespace, n.ShareProcessNamespace),
	}
}

func (n *Namespace) Validate() error {
	return nil
}
