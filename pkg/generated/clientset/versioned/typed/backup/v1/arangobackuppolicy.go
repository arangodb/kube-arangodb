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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"

	v1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	scheme "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// ArangoBackupPoliciesGetter has a method to return a ArangoBackupPolicyInterface.
// A group's client should implement this interface.
type ArangoBackupPoliciesGetter interface {
	ArangoBackupPolicies(namespace string) ArangoBackupPolicyInterface
}

// ArangoBackupPolicyInterface has methods to work with ArangoBackupPolicy resources.
type ArangoBackupPolicyInterface interface {
	Create(ctx context.Context, arangoBackupPolicy *v1.ArangoBackupPolicy, opts metav1.CreateOptions) (*v1.ArangoBackupPolicy, error)
	Update(ctx context.Context, arangoBackupPolicy *v1.ArangoBackupPolicy, opts metav1.UpdateOptions) (*v1.ArangoBackupPolicy, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, arangoBackupPolicy *v1.ArangoBackupPolicy, opts metav1.UpdateOptions) (*v1.ArangoBackupPolicy, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ArangoBackupPolicy, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.ArangoBackupPolicyList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ArangoBackupPolicy, err error)
	ArangoBackupPolicyExpansion
}

// arangoBackupPolicies implements ArangoBackupPolicyInterface
type arangoBackupPolicies struct {
	*gentype.ClientWithList[*v1.ArangoBackupPolicy, *v1.ArangoBackupPolicyList]
}

// newArangoBackupPolicies returns a ArangoBackupPolicies
func newArangoBackupPolicies(c *BackupV1Client, namespace string) *arangoBackupPolicies {
	return &arangoBackupPolicies{
		gentype.NewClientWithList[*v1.ArangoBackupPolicy, *v1.ArangoBackupPolicyList](
			"arangobackuppolicies",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.ArangoBackupPolicy { return &v1.ArangoBackupPolicy{} },
			func() *v1.ArangoBackupPolicyList { return &v1.ArangoBackupPolicyList{} }),
	}
}
