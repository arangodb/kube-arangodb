//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1beta1

import (
	deploymentv1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	schedulerv1beta1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	container "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/container"
	pod "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1/pod"
	v1 "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	corev1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtension) DeepCopyInto(out *ArangoMLExtension) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtension.
func (in *ArangoMLExtension) DeepCopy() *ArangoMLExtension {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtension)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArangoMLExtension) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionList) DeepCopyInto(out *ArangoMLExtensionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ArangoMLExtension, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionList.
func (in *ArangoMLExtensionList) DeepCopy() *ArangoMLExtensionList {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArangoMLExtensionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionSpec) DeepCopyInto(out *ArangoMLExtensionSpec) {
	*out = *in
	if in.MetadataService != nil {
		in, out := &in.MetadataService, &out.MetadataService
		*out = new(ArangoMLExtensionSpecMetadataService)
		(*in).DeepCopyInto(*out)
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = new(v1.Object)
		(*in).DeepCopyInto(*out)
	}
	if in.Init != nil {
		in, out := &in.Init, &out.Init
		*out = new(ArangoMLExtensionTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(ArangoMLExtensionSpecDeployment)
		(*in).DeepCopyInto(*out)
	}
	if in.JobsTemplates != nil {
		in, out := &in.JobsTemplates, &out.JobsTemplates
		*out = new(ArangoMLJobsTemplates)
		(*in).DeepCopyInto(*out)
	}
	if in.IntegrationSidecar != nil {
		in, out := &in.IntegrationSidecar, &out.IntegrationSidecar
		*out = new(schedulerv1beta1.IntegrationSidecar)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionSpec.
func (in *ArangoMLExtensionSpec) DeepCopy() *ArangoMLExtensionSpec {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionSpecDeployment) DeepCopyInto(out *ArangoMLExtensionSpecDeployment) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ArangoMLExtensionSpecDeploymentService)
		(*in).DeepCopyInto(*out)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(v1.TLS)
		(*in).DeepCopyInto(*out)
	}
	if in.Pod != nil {
		in, out := &in.Pod, &out.Pod
		*out = new(pod.Pod)
		(*in).DeepCopyInto(*out)
	}
	if in.Container != nil {
		in, out := &in.Container, &out.Container
		*out = new(container.Container)
		(*in).DeepCopyInto(*out)
	}
	if in.GPU != nil {
		in, out := &in.GPU, &out.GPU
		*out = new(bool)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int32)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionSpecDeployment.
func (in *ArangoMLExtensionSpecDeployment) DeepCopy() *ArangoMLExtensionSpecDeployment {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionSpecDeployment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionSpecDeploymentService) DeepCopyInto(out *ArangoMLExtensionSpecDeploymentService) {
	*out = *in
	if in.Type != nil {
		in, out := &in.Type, &out.Type
		*out = new(corev1.ServiceType)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionSpecDeploymentService.
func (in *ArangoMLExtensionSpecDeploymentService) DeepCopy() *ArangoMLExtensionSpecDeploymentService {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionSpecDeploymentService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionSpecMetadataService) DeepCopyInto(out *ArangoMLExtensionSpecMetadataService) {
	*out = *in
	if in.Local != nil {
		in, out := &in.Local, &out.Local
		*out = new(ArangoMLExtensionSpecMetadataServiceLocal)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionSpecMetadataService.
func (in *ArangoMLExtensionSpecMetadataService) DeepCopy() *ArangoMLExtensionSpecMetadataService {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionSpecMetadataService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionSpecMetadataServiceLocal) DeepCopyInto(out *ArangoMLExtensionSpecMetadataServiceLocal) {
	*out = *in
	if in.ArangoPipeDatabase != nil {
		in, out := &in.ArangoPipeDatabase, &out.ArangoPipeDatabase
		*out = new(string)
		**out = **in
	}
	if in.ArangoMLFeatureStoreDatabase != nil {
		in, out := &in.ArangoMLFeatureStoreDatabase, &out.ArangoMLFeatureStoreDatabase
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionSpecMetadataServiceLocal.
func (in *ArangoMLExtensionSpecMetadataServiceLocal) DeepCopy() *ArangoMLExtensionSpecMetadataServiceLocal {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionSpecMetadataServiceLocal)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionStatus) DeepCopyInto(out *ArangoMLExtensionStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(deploymentv1.ConditionList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.MetadataService != nil {
		in, out := &in.MetadataService, &out.MetadataService
		*out = new(ArangoMLExtensionStatusMetadataService)
		**out = **in
	}
	if in.ServiceAccount != nil {
		in, out := &in.ServiceAccount, &out.ServiceAccount
		*out = new(v1.ServiceAccount)
		(*in).DeepCopyInto(*out)
	}
	if in.ArangoDB != nil {
		in, out := &in.ArangoDB, &out.ArangoDB
		*out = new(ArangoMLExtensionStatusArangoDBRef)
		(*in).DeepCopyInto(*out)
	}
	if in.Reconciliation != nil {
		in, out := &in.Reconciliation, &out.Reconciliation
		*out = new(ArangoMLExtensionStatusReconciliation)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionStatus.
func (in *ArangoMLExtensionStatus) DeepCopy() *ArangoMLExtensionStatus {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionStatusArangoDBRef) DeepCopyInto(out *ArangoMLExtensionStatusArangoDBRef) {
	*out = *in
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(v1.Object)
		(*in).DeepCopyInto(*out)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(v1.Object)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionStatusArangoDBRef.
func (in *ArangoMLExtensionStatusArangoDBRef) DeepCopy() *ArangoMLExtensionStatusArangoDBRef {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionStatusArangoDBRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionStatusMetadataService) DeepCopyInto(out *ArangoMLExtensionStatusMetadataService) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionStatusMetadataService.
func (in *ArangoMLExtensionStatusMetadataService) DeepCopy() *ArangoMLExtensionStatusMetadataService {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionStatusMetadataService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionStatusReconciliation) DeepCopyInto(out *ArangoMLExtensionStatusReconciliation) {
	*out = *in
	if in.StatefulSet != nil {
		in, out := &in.StatefulSet, &out.StatefulSet
		*out = new(v1.Object)
		(*in).DeepCopyInto(*out)
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(v1.Object)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionStatusReconciliation.
func (in *ArangoMLExtensionStatusReconciliation) DeepCopy() *ArangoMLExtensionStatusReconciliation {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionStatusReconciliation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLExtensionTemplate) DeepCopyInto(out *ArangoMLExtensionTemplate) {
	*out = *in
	if in.Pod != nil {
		in, out := &in.Pod, &out.Pod
		*out = new(pod.Pod)
		(*in).DeepCopyInto(*out)
	}
	if in.Container != nil {
		in, out := &in.Container, &out.Container
		*out = new(container.Container)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLExtensionTemplate.
func (in *ArangoMLExtensionTemplate) DeepCopy() *ArangoMLExtensionTemplate {
	if in == nil {
		return nil
	}
	out := new(ArangoMLExtensionTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLJobTemplates) DeepCopyInto(out *ArangoMLJobTemplates) {
	*out = *in
	if in.CPU != nil {
		in, out := &in.CPU, &out.CPU
		*out = new(ArangoMLExtensionTemplate)
		(*in).DeepCopyInto(*out)
	}
	if in.GPU != nil {
		in, out := &in.GPU, &out.GPU
		*out = new(ArangoMLExtensionTemplate)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLJobTemplates.
func (in *ArangoMLJobTemplates) DeepCopy() *ArangoMLJobTemplates {
	if in == nil {
		return nil
	}
	out := new(ArangoMLJobTemplates)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLJobsTemplates) DeepCopyInto(out *ArangoMLJobsTemplates) {
	*out = *in
	if in.Prediction != nil {
		in, out := &in.Prediction, &out.Prediction
		*out = new(ArangoMLJobTemplates)
		(*in).DeepCopyInto(*out)
	}
	if in.Training != nil {
		in, out := &in.Training, &out.Training
		*out = new(ArangoMLJobTemplates)
		(*in).DeepCopyInto(*out)
	}
	if in.Featurization != nil {
		in, out := &in.Featurization, &out.Featurization
		*out = new(ArangoMLJobTemplates)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLJobsTemplates.
func (in *ArangoMLJobsTemplates) DeepCopy() *ArangoMLJobsTemplates {
	if in == nil {
		return nil
	}
	out := new(ArangoMLJobsTemplates)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorage) DeepCopyInto(out *ArangoMLStorage) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorage.
func (in *ArangoMLStorage) DeepCopy() *ArangoMLStorage {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArangoMLStorage) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorageList) DeepCopyInto(out *ArangoMLStorageList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ArangoMLStorage, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorageList.
func (in *ArangoMLStorageList) DeepCopy() *ArangoMLStorageList {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorageList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArangoMLStorageList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorageSpec) DeepCopyInto(out *ArangoMLStorageSpec) {
	*out = *in
	if in.BucketName != nil {
		in, out := &in.BucketName, &out.BucketName
		*out = new(string)
		**out = **in
	}
	if in.BucketPath != nil {
		in, out := &in.BucketPath, &out.BucketPath
		*out = new(string)
		**out = **in
	}
	if in.Mode != nil {
		in, out := &in.Mode, &out.Mode
		*out = new(ArangoMLStorageSpecMode)
		(*in).DeepCopyInto(*out)
	}
	if in.Backend != nil {
		in, out := &in.Backend, &out.Backend
		*out = new(ArangoMLStorageSpecBackend)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorageSpec.
func (in *ArangoMLStorageSpec) DeepCopy() *ArangoMLStorageSpec {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorageSpecBackend) DeepCopyInto(out *ArangoMLStorageSpecBackend) {
	*out = *in
	if in.S3 != nil {
		in, out := &in.S3, &out.S3
		*out = new(ArangoMLStorageSpecBackendS3)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorageSpecBackend.
func (in *ArangoMLStorageSpecBackend) DeepCopy() *ArangoMLStorageSpecBackend {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorageSpecBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorageSpecBackendS3) DeepCopyInto(out *ArangoMLStorageSpecBackendS3) {
	*out = *in
	if in.Endpoint != nil {
		in, out := &in.Endpoint, &out.Endpoint
		*out = new(string)
		**out = **in
	}
	if in.CredentialsSecret != nil {
		in, out := &in.CredentialsSecret, &out.CredentialsSecret
		*out = new(v1.Object)
		(*in).DeepCopyInto(*out)
	}
	if in.AllowInsecure != nil {
		in, out := &in.AllowInsecure, &out.AllowInsecure
		*out = new(bool)
		**out = **in
	}
	if in.CASecret != nil {
		in, out := &in.CASecret, &out.CASecret
		*out = new(v1.Object)
		(*in).DeepCopyInto(*out)
	}
	if in.Region != nil {
		in, out := &in.Region, &out.Region
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorageSpecBackendS3.
func (in *ArangoMLStorageSpecBackendS3) DeepCopy() *ArangoMLStorageSpecBackendS3 {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorageSpecBackendS3)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorageSpecMode) DeepCopyInto(out *ArangoMLStorageSpecMode) {
	*out = *in
	if in.Sidecar != nil {
		in, out := &in.Sidecar, &out.Sidecar
		*out = new(ArangoMLStorageSpecModeSidecar)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorageSpecMode.
func (in *ArangoMLStorageSpecMode) DeepCopy() *ArangoMLStorageSpecMode {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorageSpecMode)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorageSpecModeSidecar) DeepCopyInto(out *ArangoMLStorageSpecModeSidecar) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorageSpecModeSidecar.
func (in *ArangoMLStorageSpecModeSidecar) DeepCopy() *ArangoMLStorageSpecModeSidecar {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorageSpecModeSidecar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArangoMLStorageStatus) DeepCopyInto(out *ArangoMLStorageStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make(deploymentv1.ConditionList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArangoMLStorageStatus.
func (in *ArangoMLStorageStatus) DeepCopy() *ArangoMLStorageStatus {
	if in == nil {
		return nil
	}
	out := new(ArangoMLStorageStatus)
	in.DeepCopyInto(out)
	return out
}