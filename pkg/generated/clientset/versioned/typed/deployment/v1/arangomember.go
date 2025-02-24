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

	v1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	scheme "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// ArangoMembersGetter has a method to return a ArangoMemberInterface.
// A group's client should implement this interface.
type ArangoMembersGetter interface {
	ArangoMembers(namespace string) ArangoMemberInterface
}

// ArangoMemberInterface has methods to work with ArangoMember resources.
type ArangoMemberInterface interface {
	Create(ctx context.Context, arangoMember *v1.ArangoMember, opts metav1.CreateOptions) (*v1.ArangoMember, error)
	Update(ctx context.Context, arangoMember *v1.ArangoMember, opts metav1.UpdateOptions) (*v1.ArangoMember, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, arangoMember *v1.ArangoMember, opts metav1.UpdateOptions) (*v1.ArangoMember, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.ArangoMember, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.ArangoMemberList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ArangoMember, err error)
	ArangoMemberExpansion
}

// arangoMembers implements ArangoMemberInterface
type arangoMembers struct {
	*gentype.ClientWithList[*v1.ArangoMember, *v1.ArangoMemberList]
}

// newArangoMembers returns a ArangoMembers
func newArangoMembers(c *DatabaseV1Client, namespace string) *arangoMembers {
	return &arangoMembers{
		gentype.NewClientWithList[*v1.ArangoMember, *v1.ArangoMemberList](
			"arangomembers",
			c.RESTClient(),
			scheme.ParameterCodec,
			namespace,
			func() *v1.ArangoMember { return &v1.ArangoMember{} },
			func() *v1.ArangoMemberList { return &v1.ArangoMemberList{} }),
	}
}
