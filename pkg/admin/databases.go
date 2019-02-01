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
	"strings"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/admin/v1alpha"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Database stores information about a arangodb database
type Database struct {
	api.ArangoDatabase
}

func (db *Database) GetAPIObject() ArangoResource {
	return db
}

func (db *Database) AsRuntimeObject() runtime.Object {
	return &db.ArangoDatabase
}

func (db *Database) SetAPIObject(obj api.ArangoDatabase) {
	db.ArangoDatabase = obj
}

func (db *Database) Load(kube KubeClient) (runtime.Object, error) {
	return kube.ArangoDatabases(db.GetNamespace()).Get(db.GetName(), metav1.GetOptions{})
}

func (db *Database) Update(kube KubeClient) error {
	new, err := kube.ArangoDatabases(db.GetNamespace()).Update(&db.ArangoDatabase)
	if err != nil {
		return err
	}
	db.SetAPIObject(*new)
	return nil
}

func (db *Database) UpdateStatus(kube KubeClient) error {
	_, err := kube.ArangoDatabases(db.GetNamespace()).UpdateStatus(&db.ArangoDatabase)
	return err
}

func (db *Database) GetDeploymentName() string {
	return db.ArangoDatabase.GetDeploymentName()
}

func NewDatabaseFromObject(object runtime.Object) (*Database, error) {
	if adb, ok := object.(*api.ArangoDatabase); ok {
		adb.Spec.SetDefaults(adb.GetName())
		if err := adb.Spec.Validate(); err != nil {
			return nil, err
		}
		return &Database{
			ArangoDatabase: *adb,
		}, nil
	}

	return nil, fmt.Errorf("Not a ArangoDatabase")
}

func allCollectionsSystem(cols []driver.Collection) bool {
	for _, c := range cols {
		if !strings.HasPrefix(c.Name(), "_") {
			return false
		}
	}

	return true
}

// Reconcile updates the database resource to the given spec
func (db *Database) Reconcile(ctx context.Context, admin ReconcileContext) (bool, error) {
	dbname := db.Spec.GetName()

	if db.GetDeletionTimestamp() != nil {
		arango, err := admin.GetArangoClient(ctx, db)
		if err != nil {
			return false, errors.Wrap(err, "Could not connect to deployment")
		}

		adb, err := arango.Database(ctx, dbname)
		if driver.IsNotFound(err) {
			// database is not there
			// remove finalizer from deployment and the cr
			admin.RemoveDeploymentFinalizer(db)
			admin.RemoveFinalizer(db)

			// Resource is not ready, but no error
			return false, nil
		} else if err != nil {
			return false, errors.Wrap(err, "Could not access database")
		}

		if cols, err := adb.Collections(ctx); err != nil {
			// report error
			return false, errors.Wrap(err, "Could not access database")
		} else if !allCollectionsSystem(cols) {
			// Add event
			admin.ReportWarning(db, "Database not empty", "The database contains collections and is therefore not deleted")
			// Database is ready, no error
			return true, nil
		}

		if admin.GetCreatedAt(db) != nil {
			if err := adb.Remove(ctx); err != nil {
				return false, errors.Wrap(err, "Failed to remove database")
			}

			admin.ReportEvent(db, "Reconciliation", "Database deleted")
		}

		return false, nil
	}
	if !admin.HasFinalizer(db) {
		admin.AddFinalizer(db)
	}

	arango, err := admin.GetArangoClient(ctx, db)
	if err != nil {
		return false, errors.Wrap(err, "Could not connect to deployment")
	}

	_, err = arango.Database(ctx, dbname)
	if driver.IsNotFound(err) {
		// check if database was created before
		if admin.GetCreatedAt(db) != nil {
			admin.ReportWarning(db, "Database lost", "Database was created before and is now lost")
		}

		// create the database
		_, err := arango.CreateDatabase(ctx, dbname, nil)
		if err != nil {
			// record create error
			return false, errors.Wrap(err, "Create database failed")
		}

		admin.ReportEvent(db, "Reconciliation", "Database created")
		admin.SetCreatedAtNow(db)
		return true, nil

	} else if err != nil {
		// Generic error
		return false, errors.Wrap(err, "Could not access database")
	}

	return true, nil
}
