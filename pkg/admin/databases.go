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
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Database stores information about a arangodb database
type Database struct {
	apiObject api.ArangoDatabase
	// log
	// Dont like that updateRequired
	updateRequired bool
}

func (db *Database) GetAPIObject() ArangoResource {
	return &db.apiObject
}

func (db *Database) SetAPIObject(obj api.ArangoDatabase) {
	db.apiObject = obj
}

func (db *Database) GetNamespace() string {
	return db.apiObject.GetNamespace()
}

func (db *Database) GetName() string {
	return "db-" + db.apiObject.Spec.GetName()
}

func (db *Database) Load(kube KubeClient) (runtime.Object, error) {
	return kube.ArangoDatabases(db.apiObject.GetNamespace()).Get(db.apiObject.GetName(), metav1.GetOptions{})
}

func (db *Database) Update(kube KubeClient) error {
	new, err := kube.ArangoDatabases(db.GetNamespace()).Update(&db.apiObject)
	if err != nil {
		return err
	}
	db.apiObject = *new
	return nil
}

func (db *Database) UpdateStatus(kube KubeClient) error {
	_, err := kube.ArangoDatabases(db.GetNamespace()).UpdateStatus(&db.apiObject)
	return err
}

func (db *Database) SetUpdateRequired() {
	db.updateRequired = true
}

func (db *Database) GetDeploymentName(resolv DeploymentNameResolver) string {
	return db.apiObject.GetDeploymentName()
}

func NewDatabaseFromObject(object runtime.Object) (*Database, error) {
	if adb, ok := object.(*api.ArangoDatabase); ok {
		adb.Spec.SetDefaults(adb.GetName())
		if err := adb.Spec.Validate(); err != nil {
			return nil, err
		}
		return &Database{
			apiObject: *adb,
		}, nil
	}

	return nil, fmt.Errorf("Not a ArangoDatabase")
}

func (db *Database) ModifyObject(context ModifyObjectContext, object runtime.Object) {
	if adb, ok := object.(*api.ArangoDatabase); ok {
		adb.Spec.SetDefaultsFrom(&db.apiObject.Spec)
		if fields := db.apiObject.Spec.ResetImmutableFields(&adb.Spec); len(fields) > 0 {
			context.ResetImmutableFields(fields)
		} else {
			if err := adb.Spec.Validate(); err != nil {
				context.ValidationError(err)
			} else {
				// All got, take that one
				db.apiObject = *adb
			}
		}
	}
}

// Reconcile updates the database resource to the given spec
func (db *Database) Reconcile(ctx context.Context, admin ReconcileContext) {
	apiObject := db.apiObject
	dbname := db.apiObject.Spec.GetName()

	if apiObject.GetDeletionTimestamp() != nil {
		arango, err := admin.GetArangoClient(ctx, db)
		if err != nil {
			admin.ReportError(db, "Could not connect to deployment", err.Error())
		}

		adb, err := arango.Database(ctx, dbname)
		if driver.IsNotFound(err) {
			// database is not there
			// remove finalizer from deployment and the cr
			admin.RemoveDeploymentFinalizer(db)
			admin.RemoveFinalizer(db)

			// Finally delete the database from the internal
			admin.RemoveDatabase(db)
			return
		}

		if cols, err := adb.Collections(ctx); err != nil {
			// report error
			admin.ReportError(db, "Failed to access database", err.Error())
		} else if len(cols) > 0 {
			// Add event
			admin.ReportWarning(db, "Database not empty", "The database contains collections and is therefore not deleted")
		}

		if err := adb.Remove(ctx); err != nil {
			admin.ReportError(db, "Failed to remove database", err.Error())
		}
	} else {
		if !admin.HasFinalizer(db) {
			admin.AddFinalizer(db)
		}

		arango, err := admin.GetArangoClient(ctx, db)
		if err != nil {
			admin.ReportWarning(db, "Connection failed", "Could not connect to deployment")
			return
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
				admin.ReportError(db, "Create database failed", err.Error())
			}

			admin.SetCreatedAtNow(db)
		} else if err != nil {
			// Generic error
			admin.ReportError(db, "Failed to access deployment", err.Error())
		}

		// Database is there, everything good, set ready condition
		admin.SetCondition(db, api.ConditionTypeReady, v1.ConditionTrue, "Database ready", "Database is ready")
	}
}
