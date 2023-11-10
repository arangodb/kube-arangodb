//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package crds

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/go-driver"
)

type Definition struct {
	Version driver.Version
	CRD     *apiextensions.CustomResourceDefinition
}

func AllDefinitions() []Definition {
	return []Definition{
		// Deployment
		DatabaseDeploymentDefinition(),
		DatabaseMemberDefinition(),

		// ACS
		DatabaseClusterSynchronizationDefinition(),

		// ArangoSync
		ReplicationDeploymentReplicationDefinition(),

		// Storage
		StorageLocalStorageDefinition(),

		// Apps
		AppsJobDefinition(),
		DatabaseTaskDefinition(),

		// Backups
		BackupsBackupDefinition(),
		BackupsBackupPolicyDefinition(),

		// ML
		MLExtensionDefinition(),
		MLStorageDefinition(),

		MLCronJobDefinition(),
		MLBatchJobDefinition(),
	}
}

type CRDSchemas map[string]apiextensions.CustomResourceValidation

type CRDOptions struct {
	WithSchema *bool
}

func (c *CRDOptions) GetWithSchema() bool {
	if c == nil {
		return false
	}

	if c.WithSchema == nil {
		return false
	}

	return *c.WithSchema
}

func (c *CRDOptions) Merge(in CRDOptions) CRDOptions {
	if c == nil {
		return in
	}

	if c.WithSchema == nil {
		c.WithSchema = in.WithSchema
	}

	return *c
}

func extendCRDWithSchema(crd *apiextensions.CustomResourceDefinition, opts CRDOptions, schemas CRDSchemas) *apiextensions.CustomResourceDefinition {
	// We are already working on the copy
	for v := range crd.Spec.Versions {
		if schema, ok := schemas[crd.Spec.Versions[v].Name]; opts.GetWithSchema() && ok {
			// We have found schema, lets merge it!
			crd.Spec.Versions[v].Schema = schema.DeepCopy()
		} else {
			// Lets put default schema
			crd.Spec.Versions[v].Schema = &apiextensions.CustomResourceValidation{
				OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
					Type:                   "object",
					XPreserveUnknownFields: util.NewType(true),
				},
			}
		}
	}

	return crd
}
