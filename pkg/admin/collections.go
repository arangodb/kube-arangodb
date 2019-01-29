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

package admin

import (
	"context"
	"fmt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/admin/v1alpha"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Collection stores information about a arangodb Collection
type Collection struct {
	api.ArangoCollection
}

func (coll *Collection) GetAPIObject() ArangoResource {
	return coll
}

func (coll *Collection) AsRuntimeObject() runtime.Object {
	return &coll.ArangoCollection
}

func (coll *Collection) SetAPIObject(obj api.ArangoCollection) {
	coll.ArangoCollection = obj
}

func (coll *Collection) Load(kube KubeClient) (runtime.Object, error) {
	return kube.ArangoCollections(coll.GetNamespace()).Get(coll.GetName(), metav1.GetOptions{})
}

func (coll *Collection) Update(kube KubeClient) error {
	new, err := kube.ArangoCollections(coll.GetNamespace()).Update(&coll.ArangoCollection)
	if err != nil {
		return err
	}
	coll.SetAPIObject(*new)
	return nil
}

func (coll *Collection) UpdateStatus(kube KubeClient) error {
	_, err := kube.ArangoCollections(coll.GetNamespace()).UpdateStatus(&coll.ArangoCollection)
	return err
}

func (coll *Collection) GetDeploymentName(resolv DeploymentNameResolver) string {
	return coll.ArangoCollection.GetDeploymentName()
}

func NewCollectionFromObject(object runtime.Object) (*Collection, error) {
	if acoll, ok := object.(*api.ArangoCollection); ok {
		acoll.Spec.SetDefaults(acoll.GetName())
		if err := acoll.Spec.Validate(); err != nil {
			return nil, err
		}
		return &Collection{
			ArangoCollection: *acoll,
		}, nil
	}

	return nil, fmt.Errorf("Not a ArangoCollection")
}

// Reconcile updates the Collection resource to the given spec
func (coll *Collection) Reconcile(ctx context.Context, admin ReconcileContext) {

	if coll.GetDeletionTimestamp() != nil {
		// Collection is marked to be deleted
	}
}
