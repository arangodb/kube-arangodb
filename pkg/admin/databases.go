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
	"fmt"

	driver "github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/admin/v1alpha"
)

// Database stores information about a arangodb database
type Database struct {
	// ResourceName is the name of the kubernetes resource
	ResourceName string
	// Namespace is the namespace this resource is located in
	Namespace string
	// Name is the ArangoDB Database Name
	Name string
	// DeploymentName is the name of the deployment is database is located in
	DeploymentName string
}

// Ensure ensures that the database is present in the deployment
func (db *Database) Ensure(p DeploymentProvider) error {
	client, err := p.GetClient(nil, db.DeploymentName, db.Namespace)
	if err != nil {
		return err
	}

	if exists, err := client.DatabaseExists(nil, db.Name); err != nil {
		return err
	} else if !exists {
		if _, err = client.CreateDatabase(nil, db.Name, &driver.CreateDatabaseOptions{}); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes the database from the deployment
func (db *Database) Delete(p DeploymentProvider) error {
	client, err := p.GetClient(nil, db.DeploymentName, db.Namespace)
	if err != nil {
		return err
	}

	if db, err := client.Database(nil, db.Name); !driver.IsNotFound(err) {
		return err
	} else if err == nil {
		if err := db.Remove(nil); err != nil {
			return err
		}
	}

	return nil
}

func (da *DatabaseAdmin) UpdateDatabase(apiObject api.ArangoDatabase) error {
	apiName := apiObject.ObjectMeta.Name
	if _, ok := da.Databases[apiName]; !ok {

		var deployment string
		if deployment, ok := apiObject.Labels["deployment"]; !ok {
			return fmt.Errorf("Missing deployment label on %s", apiName)
		}

		da.Databases[apiName] = Database{
			ResourceName:   apiName,
			Namespace:      apiObject.ObjectMeta.Namespace,
			Name:           apiObject.Spec.GetName(),
			DeploymentName: deployment,
		}
	}

	return nil
}

func (da *DatabaseAdmin) DeleteDatabase(apiObject api.ArangoDatabase) error {

}
