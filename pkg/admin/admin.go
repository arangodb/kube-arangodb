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
	"time"

	driver "github.com/arangodb/go-driver"
	dapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/rs/zerolog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
)

// Dependencies holds dependent services for a DatabaseAdmin
type Dependencies struct {
	Log                zerolog.Logger
	KubeCli            kubernetes.Interface
	DatabaseAdminCRCli versioned.Interface
	EventRecorder      record.EventRecorder
}

// DeploymentProvider provides clients and api objects for deployments
type DeploymentProvider interface {
	GetClient(ctx context.Context, deployment, namespace string) (driver.Client, error)
	GetAPIObject(deployment, namespace string) (*dapi.ArangoDeployment, error)
}

// DatabaseAdmin contains all information about the resources in the current cluster
type DatabaseAdmin struct {
	Databases    map[string]Database
	Namespace    string
	Dependencies Dependencies
}

// NewDatabaseAdmin creates a new DatabaseAdmin
func NewDatabaseAdmin(Namespace string, deps Dependencies) *DatabaseAdmin {
	return &DatabaseAdmin{
		Databases:    make(map[string]Database),
		Namespace:    Namespace,
		Dependencies: deps,
	}
}

// GetAPIObject return the ArangoDeployment API object
func (da *DatabaseAdmin) GetAPIObject(deployment, namespace string) (*dapi.ArangoDeployment, error) {
	return da.Dependencies.DatabaseAdminCRCli.DatabaseV1alpha().ArangoDeployments(namespace).Get(deployment, v1.GetOptions{})
}

// GetClient returns a database client for the given deployment
func (da *DatabaseAdmin) GetClient(ctx context.Context, deployment, namespace string) (driver.Client, error) {
	if apiObject, err := da.GetAPIObject(deployment, namespace); err == nil {
		return arangod.CreateArangodDatabaseClient(ctx, da.Dependencies.KubeCli.CoreV1(), apiObject, false)
	} else {
		return nil, err
	}
}

// CheckResources ensures all resources
func (da *DatabaseAdmin) CheckResources() {
	for name, db := range da.Databases {
		if err := db.Ensure(da); err != nil {
			da.Dependencies.Log.Error().Str("name", name).Err(err).Msg("Failed to ensure database")
		}
	}
}

// Run runs the database admin
func (da *DatabaseAdmin) Run(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		case <-time.After(5 * time.Second):
			da.Dependencies.Log.Debug().Msg("Hello there! Inspecting your deployments...")
			da.CheckResources()
			da.Dependencies.Log.Debug().Msg("Finished inspection.")
		}
	}
}
