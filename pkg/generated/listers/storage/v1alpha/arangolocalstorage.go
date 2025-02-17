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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha

import (
	v1alpha "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/listers"
	"k8s.io/client-go/tools/cache"
)

// ArangoLocalStorageLister helps list ArangoLocalStorages.
// All objects returned here must be treated as read-only.
type ArangoLocalStorageLister interface {
	// List lists all ArangoLocalStorages in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha.ArangoLocalStorage, err error)
	// Get retrieves the ArangoLocalStorage from the index for a given name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha.ArangoLocalStorage, error)
	ArangoLocalStorageListerExpansion
}

// arangoLocalStorageLister implements the ArangoLocalStorageLister interface.
type arangoLocalStorageLister struct {
	listers.ResourceIndexer[*v1alpha.ArangoLocalStorage]
}

// NewArangoLocalStorageLister returns a new ArangoLocalStorageLister.
func NewArangoLocalStorageLister(indexer cache.Indexer) ArangoLocalStorageLister {
	return &arangoLocalStorageLister{listers.New[*v1alpha.ArangoLocalStorage](indexer, v1alpha.Resource("arangolocalstorage"))}
}
