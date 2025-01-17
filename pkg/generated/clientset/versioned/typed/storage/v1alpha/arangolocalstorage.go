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

package v1alpha

import (
	"context"

	v1alpha "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	scheme "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// ArangoLocalStoragesGetter has a method to return a ArangoLocalStorageInterface.
// A group's client should implement this interface.
type ArangoLocalStoragesGetter interface {
	ArangoLocalStorages() ArangoLocalStorageInterface
}

// ArangoLocalStorageInterface has methods to work with ArangoLocalStorage resources.
type ArangoLocalStorageInterface interface {
	Create(ctx context.Context, arangoLocalStorage *v1alpha.ArangoLocalStorage, opts v1.CreateOptions) (*v1alpha.ArangoLocalStorage, error)
	Update(ctx context.Context, arangoLocalStorage *v1alpha.ArangoLocalStorage, opts v1.UpdateOptions) (*v1alpha.ArangoLocalStorage, error)
	// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
	UpdateStatus(ctx context.Context, arangoLocalStorage *v1alpha.ArangoLocalStorage, opts v1.UpdateOptions) (*v1alpha.ArangoLocalStorage, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha.ArangoLocalStorage, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha.ArangoLocalStorageList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha.ArangoLocalStorage, err error)
	ArangoLocalStorageExpansion
}

// arangoLocalStorages implements ArangoLocalStorageInterface
type arangoLocalStorages struct {
	*gentype.ClientWithList[*v1alpha.ArangoLocalStorage, *v1alpha.ArangoLocalStorageList]
}

// newArangoLocalStorages returns a ArangoLocalStorages
func newArangoLocalStorages(c *StorageV1alphaClient) *arangoLocalStorages {
	return &arangoLocalStorages{
		gentype.NewClientWithList[*v1alpha.ArangoLocalStorage, *v1alpha.ArangoLocalStorageList](
			"arangolocalstorages",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *v1alpha.ArangoLocalStorage { return &v1alpha.ArangoLocalStorage{} },
			func() *v1alpha.ArangoLocalStorageList { return &v1alpha.ArangoLocalStorageList{} }),
	}
}
