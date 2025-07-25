//go:build !ignore_autogenerated
// +build !ignore_autogenerated

//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

// Code generated by deepcopy-gen. DO NOT EDIT.

package types

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	out.Platform = in.Platform
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServicePlatform) DeepCopyInto(out *ServicePlatform) {
	*out = *in
	out.Deployment = in.Deployment
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServicePlatform.
func (in *ServicePlatform) DeepCopy() *ServicePlatform {
	if in == nil {
		return nil
	}
	out := new(ServicePlatform)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServicePlatformDeployment) DeepCopyInto(out *ServicePlatformDeployment) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServicePlatformDeployment.
func (in *ServicePlatformDeployment) DeepCopy() *ServicePlatformDeployment {
	if in == nil {
		return nil
	}
	out := new(ServicePlatformDeployment)
	in.DeepCopyInto(out)
	return out
}
