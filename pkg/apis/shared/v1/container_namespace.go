//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

type ContainerNamespace struct {
	// HostNetwork requests Host network for this pod. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// +doc/default: false
	HostNetwork bool `json:"hostNetwork,omitempty" protobuf:"varint,11,opt,name=hostNetwork"`
	// HostPID define to use the host's pid namespace.
	// +doc/default: false
	HostPID bool `json:"hostPID,omitempty" protobuf:"varint,12,opt,name=hostPID"`
	// HostIPC defines to use the host's ipc namespace.
	// +doc/default: false
	HostIPC bool `json:"hostIPC,omitempty" protobuf:"varint,13,opt,name=hostIPC"`
	// ShareProcessNamespace defines to share a single process namespace between all of the containers in a pod.
	// When this is set containers will be able to view and signal processes from other containers
	// in the same pod, and the first process in each container will not be assigned PID 1.
	// HostPID and ShareProcessNamespace cannot both be set.
	// +doc/default: false
	ShareProcessNamespace *bool `json:"shareProcessNamespace,omitempty" protobuf:"varint,27,opt,name=shareProcessNamespace"`
}

func (c *ContainerNamespace) GetHostNetwork() bool {
	if c == nil {
		return false
	}

	return c.HostNetwork
}

func (c *ContainerNamespace) GetHostPID() bool {
	if c == nil {
		return false
	}

	return c.HostPID
}

func (c *ContainerNamespace) GetHostIPC() bool {
	if c == nil {
		return false
	}

	return c.HostIPC
}

func (c *ContainerNamespace) GetShareProcessNamespace() *bool {
	if c == nil {
		return nil
	}

	return c.ShareProcessNamespace
}

func (c *ContainerNamespace) Validate() error {
	return nil
}
