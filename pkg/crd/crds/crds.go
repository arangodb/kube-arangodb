//
// DISCLAIMER
//
// Copyright 2023-2024 ArangoDB GmbH, Cologne, Germany
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
	"fmt"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Definition struct {
	DefinitionData
	CRD *apiextensions.CustomResourceDefinition
}

type DefinitionData struct {
	definition       []byte
	schemaDefinition []byte
}

func (d DefinitionData) definitionLoader() util.Loader[apiextensions.CustomResourceDefinition] {
	return util.NewYamlLoader[apiextensions.CustomResourceDefinition](d.definition)
}

func (d DefinitionData) schemaDefinitionLoader() util.Loader[crdSchemas] {
	return util.NewYamlLoader[crdSchemas](d.schemaDefinition)
}

func (d DefinitionData) Checksum() (definition, schema string) {
	if len(d.definition) > 0 {
		definition = util.MD5(d.definition)
	}
	if len(d.schemaDefinition) > 0 {
		schema = util.MD5(d.schemaDefinition)
	}
	return
}

func AllDefinitions() []Definition {
	return []Definition{
		// Deployment
		DatabaseDeploymentDefinitionWithOptions(),
		DatabaseMemberDefinitionWithOptions(),

		// ACS
		DatabaseClusterSynchronizationDefinitionWithOptions(),

		// ArangoSync
		ReplicationDeploymentReplicationDefinitionWithOptions(),

		// Storage
		StorageLocalStorageDefinitionWithOptions(),

		// Apps
		AppsJobDefinitionWithOptions(),
		DatabaseTaskDefinitionWithOptions(),

		// Backups
		BackupsBackupDefinitionWithOptions(),
		BackupsBackupPolicyDefinitionWithOptions(),

		// ML
		MLExtensionDefinitionWithOptions(),
		MLStorageDefinitionWithOptions(),
		MLCronJobDefinitionWithOptions(),
		MLBatchJobDefinitionWithOptions(),

		// Scheduler
		SchedulerProfileDefinitionWithOptions(),

		// Analytics
		AnalyticsGAEDefinitionWithOptions(),

		// Networking
		NetworkingRouteDefinitionWithOptions(),
	}
}

type crdSchemas map[string]apiextensions.CustomResourceValidation

func DefaultCRDOptions() *CRDOptions {
	return &CRDOptions{
		WithSchema:   false,
		WithPreserve: true,
	}
}

type CRDOptions struct {
	WithSchema   bool
	WithPreserve bool
}

func (o *CRDOptions) GetWithSchema() bool {
	if o == nil {
		return false
	}

	return o.WithSchema
}

func (o *CRDOptions) GetWithPreserve() bool {
	if o == nil {
		return false
	}

	return o.WithPreserve
}

func (o *CRDOptions) AsFunc() func(*CRDOptions) {
	return func(opts *CRDOptions) {
		if opts == nil {
			return
		}

		if o != nil {
			opts.WithSchema = o.GetWithSchema()
		}

		if o != nil {
			opts.WithPreserve = o.GetWithPreserve()
		}
	}
}

func WithSchema() func(*CRDOptions) {
	return func(o *CRDOptions) {
		o.WithSchema = true
	}
}

func WithoutSchema() func(*CRDOptions) {
	return func(o *CRDOptions) {
		o.WithSchema = false
	}
}

func WithPreserve() func(*CRDOptions) {
	return func(o *CRDOptions) {
		o.WithPreserve = true
	}
}

func WithoutPreserve() func(*CRDOptions) {
	return func(o *CRDOptions) {
		o.WithPreserve = false
	}
}

func getCRD(data DefinitionData, opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition {
	o := DefaultCRDOptions()
	for _, fn := range opts {
		fn(o)
	}

	crd := data.definitionLoader().MustGet()

	if o.WithSchema {
		crdWithSchema := crd.DeepCopy()

		schemas := data.schemaDefinitionLoader().MustGet()

		for i, v := range crdWithSchema.Spec.Versions {
			schema, ok := schemas[v.Name]
			if !ok {
				panic(fmt.Sprintf("Validation schema is not defined for version %s of %s", v.Name, crd.Name))
			}
			crdWithSchema.Spec.Versions[i].Schema = schema.DeepCopy()
			if s := crdWithSchema.Spec.Versions[i].Schema.OpenAPIV3Schema; s != nil {
				if o.GetWithPreserve() {
					s.XPreserveUnknownFields = util.NewType(true)
				}
			}
		}

		return crdWithSchema
	}
	return crd.DeepCopy()
}
