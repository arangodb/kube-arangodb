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

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Definition struct {
	Version driver.Version
	CRD     *apiextensions.CustomResourceDefinition
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
	}
}

type crdSchemas map[string]apiextensions.CustomResourceValidation

type CRDOptions struct {
	WithSchema bool
}

func (o *CRDOptions) AsFunc() func(*CRDOptions) {
	return func(opts *CRDOptions) {
		if o == nil || opts == nil {
			opts = &CRDOptions{}
		} else {
			opts.WithSchema = o.WithSchema
		}
	}
}

func WithSchema() func(*CRDOptions) {
	return func(o *CRDOptions) {
		o.WithSchema = true
	}
}

func getCRD(crdLoader util.Loader[apiextensions.CustomResourceDefinition], schemasLoader util.Loader[crdSchemas], opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition {
	o := &CRDOptions{}
	for _, fn := range opts {
		fn(o)
	}

	crd := crdLoader.MustGet()

	if o.WithSchema {
		crdWithSchema := crd.DeepCopy()

		schemas := schemasLoader.MustGet()

		for i, v := range crdWithSchema.Spec.Versions {
			schema, ok := schemas[v.Name]
			if !ok {
				panic(fmt.Sprintf("Validation schema is not defined for version %s of %s", v.Name, crd.Name))
			}
			crdWithSchema.Spec.Versions[i].Schema = schema.DeepCopy()
		}

		return crdWithSchema
	}
	return crd.DeepCopy()
}
