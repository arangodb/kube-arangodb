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

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/admin/v1alpha"
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/util"
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

func (coll *Collection) GetDeploymentName() string {
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

// GetFinalizerName returns the name of the finalizer for this collection
func (coll *Collection) GetFinalizerName() string {
	return "database-admin-collection-" + coll.Spec.GetName()
}

// Reconcile updates the Collection resource to the given spec
func (coll *Collection) Reconcile(ctx context.Context, admin ReconcileContext) (bool, error) {

	dbn := coll.GetDatabaseName()
	finalizerName := coll.GetFinalizerName()

	if coll.GetDeletionTimestamp() != nil {
		removeFinalizers := false
		defer func() {
			if removeFinalizers {
				admin.RemoveFinalizer(coll)
				if dbr, ok := admin.GetDatabaseResourceByDatabaseName(coll, dbn); ok {
					admin.RemoveResourceFinalizer(dbr, finalizerName)
				}
			}
		}()

		// Collection is marked to be deleted
		client, err := admin.GetArangoDatabaseClient(ctx, coll, coll.GetDatabaseName())
		if driver.IsNotFound(err) {
			removeFinalizers = true // Database gone!
			return true, nil
		} else if err != nil {
			return false, errors.Wrap(err, "Could not connect to deployment")
		}
		acoll, err := client.Collection(ctx, coll.Spec.GetName())
		if driver.IsNotFound(err) {
			// Collection not found - great!
			removeFinalizers = true
			return false, nil
		} else if err == nil {
			// Check if the collection was created by the operator
			if admin.GetCreatedAt(coll) != nil {
				// Delete the collection
				if err := acoll.Remove(ctx); err != nil {
					admin.ReportError(coll, "Remove collection", err)
					return false, errors.Wrap(err, "Could not remove collection")
				}
				admin.ReportEvent(coll, "Reconciliation", "Collection deleted")
			}
			removeFinalizers = true
			return false, nil
		}
		return false, errors.Wrap(err, "Could not access collection")
	}

	if !admin.HasFinalizer(coll) {
		admin.AddFinalizer(coll)
	}

	if dbr, ok := admin.GetDatabaseResourceByDatabaseName(coll, dbn); ok {
		admin.AddResourceFinalizer(dbr, finalizerName)
	}

	// Collection is not delete
	client, err := admin.GetArangoDatabaseClient(ctx, coll, dbn)
	if err != nil {
		return false, errors.Wrap(err, "Could not connect to deployment")
	}
	acoll, err := client.Collection(ctx, coll.Spec.GetName())
	if driver.IsNotFound(err) {

		if admin.GetCreatedAt(coll) != nil {
			admin.ReportWarning(coll, "Collection lost", "The collection was lost and will be recreated")
		}

		// Collection is not there
		_, err := client.CreateCollection(ctx, coll.Spec.GetName(), coll.getCreateOptions())
		if err != nil {
			return false, errors.Wrap(err, "Could not create collection")
		}
		admin.SetCreatedAtNow(coll)
		admin.ReportEvent(coll, "Reconciliation", "Collection created")
		return true, nil
	} else if err != nil {
		return false, errors.Wrap(err, "Could not access collection")
	}

	props, err := acoll.Properties(ctx)
	if err != nil {
		return false, errors.Wrap(err, "Could not get collection properties")
	}

	update, updateRequired, err := coll.getUpdateProperties(props)
	if err != nil {
		return false, errors.Wrap(err, "Can not update properties")
	}

	if updateRequired {
		if err = acoll.SetProperties(ctx, update); err != nil {
			return false, errors.Wrap(err, "Could not update collection properties")
		}
		admin.ReportEvent(coll, "Reconciliation", "Collection properties updated")
	}

	// Collection is there
	return true, nil
}

func (coll *Collection) getUpdateProperties(props driver.CollectionProperties) (driver.SetCollectionPropertiesOptions, bool, error) {
	spec := coll.Spec
	var opts driver.SetCollectionPropertiesOptions
	updateRequired := false

	if spec.GetReplicationFactor() != props.ReplicationFactor {
		opts.ReplicationFactor = spec.GetReplicationFactor()
		updateRequired = true
	}

	if spec.GetWaitForSync() != props.WaitForSync {
		opts.WaitForSync = util.NewBool(spec.GetWaitForSync())
		updateRequired = true
	}

	return opts, updateRequired, nil
}

func (coll *Collection) getCreateOptions() *driver.CreateCollectionOptions {
	spec := coll.Spec
	return &driver.CreateCollectionOptions{
		ReplicationFactor: spec.GetReplicationFactor(),
		WaitForSync:       spec.GetWaitForSync(),
		NumberOfShards:    spec.GetNumberOfShards(),
	}
}
