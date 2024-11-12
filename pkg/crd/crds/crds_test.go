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
	"testing"

	"github.com/stretchr/testify/require"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	"github.com/arangodb/kube-arangodb/pkg/apis/ml"
	"github.com/arangodb/kube-arangodb/pkg/apis/platform"
	"github.com/arangodb/kube-arangodb/pkg/apis/replication"
	"github.com/arangodb/kube-arangodb/pkg/apis/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/apis/storage"
)

func ensureCRDCompliance(t *testing.T, name string, crdDef *apiextensions.CustomResourceDefinition, schemaExpected, preserveExpected bool) {
	t.Helper()

	require.NotNil(t, crdDef)
	require.Equal(t, name, crdDef.GetName())
	for _, version := range crdDef.Spec.Versions {
		t.Run(name+" "+version.Name, func(t *testing.T) {
			require.NotNil(t, version.Schema)
			require.Equal(t, "object", version.Schema.OpenAPIV3Schema.Type)

			if preserveExpected {
				require.NotNil(t, version.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
				require.True(t, *version.Schema.OpenAPIV3Schema.XPreserveUnknownFields)
			} else {
				require.Nil(t, version.Schema.OpenAPIV3Schema.XPreserveUnknownFields)

				if version.Subresources != nil {
					if version.Subresources.Status != nil {
						t.Run("Status", func(t *testing.T) {
							require.Contains(t, version.Schema.OpenAPIV3Schema.Properties, "status")
							status := version.Schema.OpenAPIV3Schema.Properties["status"]
							require.NotNil(t, status.XPreserveUnknownFields)
							require.True(t, *status.XPreserveUnknownFields)
						})
					}
				}
			}

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
		{scheduler.ArangoProfileCRDName, SchedulerProfileDefinitionWithOptions},
		{scheduler.PodCRDName, SchedulerPodDefinitionWithOptions},
		{scheduler.DeploymentCRDName, SchedulerDeploymentDefinitionWithOptions},
		{scheduler.BatchJobCRDName, SchedulerBatchJobDefinitionWithOptions},
		{scheduler.CronJobCRDName, SchedulerCronJobDefinitionWithOptions},
		{platform.ArangoPlatformStorageCRDName, PlatformStorageDefinitionWithOptions},
		{platform.ArangoPlatformChartCRDName, PlatformChartDefinitionWithOptions},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s-no-schema", tc.name), func(t *testing.T) {
			ensureCRDCompliance(t, tc.name, tc.getter().CRD, false, true)
		})
		t.Run(fmt.Sprintf("%s-with-schema", tc.name), func(t *testing.T) {
			ensureCRDCompliance(t, tc.name, tc.getter(WithSchema()).CRD, true, true)
		})
		t.Run(fmt.Sprintf("%s-with-schema-np", tc.name), func(t *testing.T) {
			ensureCRDCompliance(t, tc.name, tc.getter(WithSchema(), WithoutPreserve()).CRD, true, false)
		})
	}
}

func Test_AllDefinitionsDefined(t *testing.T) {
	registered := map[string]bool{}

	for _, def := range AllDefinitions() {
		a, b := def.Checksum()

		require.NotEmpty(t, a)
		require.NotEmpty(t, b)

		require.NotContains(t, registered, a)
		require.NotContains(t, registered, b)

		registered[a] = true
		registered[b] = true

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
		SchedulerProfileWithOptions,
		SchedulerPodWithOptions,
		SchedulerDeploymentWithOptions,
		SchedulerBatchJobWithOptions,
		SchedulerCronJobWithOptions,
		AnalyticsGAEWithOptions,
		NetworkingRouteWithOptions,
		PlatformStorageWithOptions,
		PlatformChartWithOptions,
	}
	require.Equal(t, len(AllDefinitions()), len(getters))

	for _, g := range getters {
		t.Run("no-schema", func(t *testing.T) {
			crd := g()
			require.NotNil(t, crd)
			ensureCRDCompliance(t, crd.Spec.Names.Plural+"."+crd.Spec.Group, crd, false, true)
		})

		t.Run("with-schema", func(t *testing.T) {
			crdWithSchema := g(WithSchema())
			require.NotNil(t, crdWithSchema)
			ensureCRDCompliance(t, crdWithSchema.Spec.Names.Plural+"."+crdWithSchema.Spec.Group+"", crdWithSchema, true, true)
		})
	}
}
