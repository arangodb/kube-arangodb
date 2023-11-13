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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	"github.com/arangodb/kube-arangodb/pkg/apis/replication"
	"github.com/arangodb/kube-arangodb/pkg/apis/storage"
)

func ensureCRDCompliance(t *testing.T, name string, crdDef *apiextensions.CustomResourceDefinition, schemaExpected bool) {
	t.Helper()

	require.NotNil(t, crdDef)
	require.Equal(t, name, crdDef.GetName())
	for _, version := range crdDef.Spec.Versions {
		t.Run(name+" "+version.Name, func(t *testing.T) {
			require.NotNil(t, version.Schema)
			require.Equal(t, "object", version.Schema.OpenAPIV3Schema.Type)
			require.NotNil(t, version.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
			require.True(t, *version.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
			if schemaExpected {
				require.NotEmpty(t, version.Schema.OpenAPIV3Schema.Properties)
			} else {
				require.Empty(t, version.Schema.OpenAPIV3Schema.Properties)
			}
		})
	}
}

func Test_CRD(t *testing.T) {
	testCases := []struct {
		name   string
		getter func(opts ...func(options *CRDOptions)) Definition
	}{
		{apps.ArangoJobCRDName, AppsJobDefinitionWithOptions},
		{backup.ArangoBackupCRDName, BackupsBackupDefinitionWithOptions},
		{backup.ArangoBackupPolicyCRDName, BackupsBackupPolicyDefinitionWithOptions},
		{deployment.ArangoClusterSynchronizationCRDName, DatabaseClusterSynchronizationDefinitionWithOptions},
		{deployment.ArangoDeploymentCRDName, DatabaseDeploymentDefinitionWithOptions},
		{deployment.ArangoMemberCRDName, DatabaseMemberDefinitionWithOptions},
		{deployment.ArangoTaskCRDName, DatabaseTaskDefinitionWithOptions},
		{replication.ArangoDeploymentReplicationCRDName, ReplicationDeploymentReplicationDefinitionWithOptions},
		{storage.ArangoLocalStorageCRDName, StorageLocalStorageDefinitionWithOptions},
		{ml.ArangoMLExtensionCRDName, MLExtensionDefinitionWithOptions},
		{ml.ArangoMLStorageCRDName, MLStorageDefinitionWithOptions},
		{ml.ArangoMLCronJobCRDName, MLCronJobDefinitionWithOptions},
		{ml.ArangoMLBatchJobCRDName, MLBatchJobDefinitionWithOptions},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-no-schema", tc.name), func(t *testing.T) {
			ensureCRDCompliance(t, tc.name, tc.getter().CRD, false)
		})
		t.Run(fmt.Sprintf("%s-with-schema", tc.name), func(t *testing.T) {
			ensureCRDCompliance(t, tc.name, tc.getter(WithSchema()).CRD, true)
		})
	}
}

func Test_AllDefinitionsDefined(t *testing.T) {
	for _, def := range AllDefinitions() {
		require.NotEmpty(t, def.Version)
		require.NotNil(t, def.CRD)
	}
}

func Test_CRDGetters(t *testing.T) {
	// getters are exposed for the usage by customers
	getters := []func(opts ...func(*CRDOptions)) *apiextensions.CustomResourceDefinition{
		AppsJobWithOptions,
		BackupsBackupWithOptions,
		BackupsBackupPolicyPolicyWithOptions,
		DatabaseClusterSynchronizationWithOptions,
		DatabaseDeploymentWithOptions,
		DatabaseMemberWithOptions,
		DatabaseTaskWithOptions,
		MLExtensionWithOptions,
		MLBatchJobWithOptions,
		MLCronJobWithOptions,
		MLStorageWithOptions,
		ReplicationDeploymentReplicationWithOptions,
		StorageLocalStorageWithOptions,
	}
	require.Equal(t, len(AllDefinitions()), len(getters))

	for _, g := range getters {
		t.Run("no-schema", func(t *testing.T) {
			crd := g()
			require.NotNil(t, crd)
			ensureCRDCompliance(t, crd.Spec.Names.Plural+"."+crd.Spec.Group, crd, false)
		})

		t.Run("with-schema", func(t *testing.T) {
			crdWithSchema := g(WithSchema())
			require.NotNil(t, crdWithSchema)
			ensureCRDCompliance(t, crdWithSchema.Spec.Names.Plural+"."+crdWithSchema.Spec.Group+"", crdWithSchema, true)
		})
	}
}
