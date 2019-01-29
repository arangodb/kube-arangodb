//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
package v1alpha

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoUserList is a list of ArangoDB users.
type ArangoUserList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata
	// More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArangoUser `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ArangoUser contains the entire Kubernetes info for an ArangoDB database deployment.
type ArangoUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UserSpec       `json:"spec"`
	Status            ResourceStatus `json:"status"`
}

// UserSpec is
type UserSpec struct {
	// Name of the user in arangodb. Default is resource name
	Name *string `json:"name,omitempty"`
	// Secret name of the password secret, default is <deployment-name>-<user-name>-password
	PasswordSecretName *string `json:"passwordSecretName,omitempty"`
	// Name of the deployment this is user is part of
	DeploymentName string `json:"deploymentName,omitempty"`
}

func (as *ArangoUser) GetDeploymentName() string {
	return as.Spec.DeploymentName
}

// GetStatus returns the resource status of the database
func (as *ArangoUser) GetStatus() *ResourceStatus {
	return &as.Status
}

func (as *ArangoUser) GetMeta() *metav1.ObjectMeta {
	return &as.ObjectMeta
}

func (as *ArangoUser) AsOwner() metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: SchemeGroupVersion.String(),
		Kind:       ArangoDatabaseResourceKind,
		Name:       as.Name,
		UID:        as.UID,
	}
}

// GetName returns the name of the database or empty string
func (us *UserSpec) GetName() string {
	return util.StringOrDefault(us.Name)
}

// GetDeploymentName returns the name of the deployment
func (us *UserSpec) GetDeploymentName() string {
	return us.DeploymentName
}

// GetPasswordSecretName returns the password secret name or empty string
func (us *UserSpec) GetPasswordSecretName() string {
	return util.StringOrDefault(us.PasswordSecretName)
}

// Validate validates a UserSpec
func (us *UserSpec) Validate() error {
	return nil
}

func defaultPasswordSecretName(deploymentName, username string) string {
	return deploymentName + "-" + username + "-password"
}

// SetDefaults sets the default values for a DatabaseSpec
func (us *UserSpec) SetDefaults(resourceName string) {
	if us.Name == nil {
		us.Name = util.NewString(resourceName)
	}
	if us.PasswordSecretName == nil {
		us.PasswordSecretName = util.NewString(defaultPasswordSecretName(us.DeploymentName, us.GetName()))
	}
}

// SetDefaultsFrom fills in the values not specified with the values form source
func (us *UserSpec) SetDefaultsFrom(source *UserSpec) {
	if us.Name == nil {
		us.Name = util.NewStringOrNil(source.Name)
	}
	if us.PasswordSecretName == nil {
		us.PasswordSecretName = util.NewStringOrNil(source.PasswordSecretName)
	}
	if us.DeploymentName == "" {
		us.DeploymentName = source.DeploymentName
	}
}

// ResetImmutableFields replaces all immutable fields in the given target with values from the source spec.
// It returns a list of fields that have been reset.
// Field names are relative to `spec.`.
func (us *UserSpec) ResetImmutableFields(target *UserSpec) []string {
	var resetFields []string
	if us.GetName() != target.GetName() {
		target.Name = util.NewStringOrNil(us.Name)
		resetFields = append(resetFields, "Name")
	}
	if us.GetPasswordSecretName() != target.GetPasswordSecretName() {
		target.PasswordSecretName = util.NewStringOrNil(us.PasswordSecretName)
		resetFields = append(resetFields, "PasswordSecretName")
	}
	if us.GetDeploymentName() != target.GetDeploymentName() {
		target.DeploymentName = us.DeploymentName
		resetFields = append(resetFields, "DeploymentName")
	}
	return resetFields
}
