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

package internal

import (
	"fmt"
	"go/token"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"

	analyticsApi "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
	appsv1 "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	backupv1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	deploymentv1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	deploymentv2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v2alpha1"
	mlApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1alpha1"
	replicationv1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	replicationv2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v2alpha1"
	schedulerApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	storagev1alpha "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
)

func (def DocDefinition) ApplyToSchema(s *apiextensions.JSONSchemaProps) {
	for _, e := range def.Enum {
		z := strings.Split(e, "|")
		s.Enum = append(s.Enum, apiextensions.JSON{
			Raw: []byte("\"" + z[0] + "\""),
		})
	}

	s.Description = strings.Join(def.Docs, "\n")
}

// Test_GenerateCRValidationSchemas generates validation schema JSONs for each CRD referenced in `input` (see impl)
func Test_GenerateCRValidationSchemas(t *testing.T) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	type genSpec struct {
		objects map[string]interface{}
	}

	fset := token.NewFileSet()

	sharedFields := parseSourceFiles(t, root, fset, fmt.Sprintf("%s/pkg/apis/shared/v1", root))

	// CR file prefix -> packages to parse -> versions -> obj
	input := map[string]map[string]map[string]genSpec{
		"apps-job": {
			fmt.Sprintf("%s/pkg/apis/apps/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": appsv1.ArangoJob{}.Spec,
					},
				},
			},
		},
		"backups-backup": {
			fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": backupv1.ArangoBackup{}.Spec,
					},
				},
				"v1alpha": {
					objects: map[string]interface{}{
						"spec": backupv1.ArangoBackup{}.Spec,
					},
				},
			},
		},
		"backups-backuppolicy": {
			fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": backupv1.ArangoBackupPolicy{}.Spec,
					},
				},
				"v1alpha": {
					objects: map[string]interface{}{
						"spec": backupv1.ArangoBackupPolicy{}.Spec,
					},
				},
			},
		},
		"database-deployment": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": deploymentv1.ArangoDeployment{}.Spec,
					},
				},
				"v1alpha": {
					objects: map[string]interface{}{
						"spec": deploymentv1.ArangoDeployment{}.Spec,
					},
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					objects: map[string]interface{}{
						"spec": deploymentv2alpha1.ArangoDeployment{}.Spec,
					},
				},
			},
		},
		"database-member": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": deploymentv1.ArangoMember{}.Spec,
					},
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					objects: map[string]interface{}{
						"spec": deploymentv2alpha1.ArangoMember{}.Spec,
					},
				},
			},
		},
		"database-clustersynchronization": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": deploymentv1.ArangoClusterSynchronization{}.Spec,
					},
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					objects: map[string]interface{}{
						"spec": deploymentv2alpha1.ArangoClusterSynchronization{}.Spec,
					},
				},
			},
		},
		"database-task": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": deploymentv1.ArangoTask{}.Spec,
					},
				},
				"v1alpha": {
					objects: map[string]interface{}{
						"spec": deploymentv1.ArangoTask{}.Spec,
					},
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					objects: map[string]interface{}{
						"spec": deploymentv2alpha1.ArangoTask{}.Spec,
					},
				},
			},
		},
		"replication-deploymentreplication": {
			fmt.Sprintf("%s/pkg/apis/replication/v1", root): {
				"v1": {
					objects: map[string]interface{}{
						"spec": replicationv1.ArangoDeploymentReplication{}.Spec,
					},
				},
				"v1alpha": {
					objects: map[string]interface{}{
						"spec": replicationv1.ArangoDeploymentReplication{}.Spec,
					},
				},
			},
			fmt.Sprintf("%s/pkg/apis/replication/v2alpha1", root): {
				"v2alpha1": {
					objects: map[string]interface{}{
						"spec": replicationv2alpha1.ArangoDeploymentReplication{}.Spec,
					},
				},
			},
		},
		"storage-localstorage": {
			fmt.Sprintf("%s/pkg/apis/storage/v1alpha", root): {
				"v1alpha": {
					objects: map[string]interface{}{
						"spec": storagev1alpha.ArangoLocalStorage{}.Spec,
					},
				},
			},
		},
		"scheduler-profile": {
			fmt.Sprintf("%s/pkg/apis/scheduler/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": schedulerApiv1alpha1.ArangoProfile{}.Spec,
					},
				},
				"v1beta1": {
					objects: map[string]interface{}{
						"spec": schedulerApi.ArangoProfile{}.Spec,
					},
				},
			},
		},
		"scheduler-pod": {
			fmt.Sprintf("%s/pkg/apis/scheduler/v1alpha1", root): {
				"v1beta1": {
					objects: map[string]interface{}{
						"spec": schedulerApi.ArangoSchedulerPod{}.Spec,
					},
				},
			},
		},
		"scheduler-deployment": {
			fmt.Sprintf("%s/pkg/apis/scheduler/v1alpha1", root): {
				"v1beta1": {
					objects: map[string]interface{}{
						"spec": schedulerApi.ArangoSchedulerDeployment{}.Spec,
					},
				},
			},
		},
		"scheduler-batchjob": {
			fmt.Sprintf("%s/pkg/apis/scheduler/v1alpha1", root): {
				"v1beta1": {
					objects: map[string]interface{}{
						"spec": schedulerApi.ArangoSchedulerBatchJob{}.Spec,
					},
				},
			},
		},
		"scheduler-cronjob": {
			fmt.Sprintf("%s/pkg/apis/scheduler/v1alpha1", root): {
				"v1beta1": {
					objects: map[string]interface{}{
						"spec": schedulerApi.ArangoSchedulerCronJob{}.Spec,
					},
				},
			},
		},
		"ml-extension": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": mlApiv1alpha1.ArangoMLExtension{}.Spec,
					},
				},
				"v1beta1": {
					objects: map[string]interface{}{
						"spec": mlApi.ArangoMLExtension{}.Spec,
					},
				},
			},
		},
		"ml-storage": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": mlApiv1alpha1.ArangoMLStorage{}.Spec,
					},
				},
				"v1beta1": {
					objects: map[string]interface{}{
						"spec": mlApi.ArangoMLStorage{}.Spec,
					},
				},
			},
		},
		"ml-job-cron": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": mlApiv1alpha1.ArangoMLCronJob{}.Spec,
					},
				},
			},
		},
		"ml-job-batch": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": mlApiv1alpha1.ArangoMLBatchJob{}.Spec,
					},
				},
			},
		},
		"analytics-graphanalyticsengine": {
			fmt.Sprintf("%s/pkg/apis/analytics/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": analyticsApi.GraphAnalyticsEngine{}.Spec,
					},
				},
			},
		},
		"networking-route": {
			fmt.Sprintf("%s/pkg/apis/networking/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": networkingApi.ArangoRoute{}.Spec,
					},
				},
			},
		},
		"platform-storage": {
			fmt.Sprintf("%s/pkg/apis/platform/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": platformApi.ArangoPlatformStorage{}.Spec,
					},
				},
			},
		},
		"platform-chart": {
			fmt.Sprintf("%s/pkg/apis/platform/v1alpha1", root): {
				"v1alpha1": {
					objects: map[string]interface{}{
						"spec": platformApi.ArangoPlatformChart{}.Spec,
					},
				},
			},
		},
	}

	for filePrefix, packagesToVersion := range input {
		t.Run(filePrefix, func(t *testing.T) {
			// Preload Definition
			data, err := os.ReadFile(fmt.Sprintf("%s/pkg/crd/crds/%s.yaml", root, filePrefix))
			require.NoError(t, err)

			var crd apiextensions.CustomResourceDefinition

			require.NoError(t, yaml.Unmarshal(data, &crd))

			validationPerVersion := make(map[string]apiextensions.CustomResourceValidation, len(packagesToVersion))
			for apiDir, versionMap := range packagesToVersion {
				fields := parseSourceFiles(t, root, fset, apiDir)

				for n, f := range sharedFields {
					require.NotContains(t, fields, n)
					fields[n] = f
				}

				for version, generationSpec := range versionMap {
					crdVersion := findCRDVersion(t, crd, version)

					t.Log(crdVersion.Schema)

					if _, ok := generationSpec.objects["status"]; !ok {
						generationSpec.objects["status"] = allowAnyType{}
					}

					sb := newSchemaBuilder(root, fields, fset)

					objects := map[string]apiextensions.JSONSchemaProps{}

					for k, obj := range generationSpec.objects {
						s := sb.TypeToSchema(t, reflect.TypeOf(obj), fmt.Sprintf(".%s", k))
						require.NotNil(t, s, version)

						objects[k] = *s
					}

					validationPerVersion[version] = apiextensions.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
							Type:       "object",
							Properties: objects,
						},
					}
				}
			}

			yamlRaw, err := yaml.Marshal(validationPerVersion)
			require.NoError(t, err)

			outPath := path.Join(root, "pkg/crd/crds", fmt.Sprintf("%s.schema.generated.yaml", filePrefix))
			err = os.WriteFile(outPath, yamlRaw, 0644)
			require.NoError(t, err)
		})
	}
}

func findCRDVersion(t *testing.T, crd apiextensions.CustomResourceDefinition, version string) apiextensions.CustomResourceDefinitionVersion {
	for _, v := range crd.Spec.Versions {
		if v.Name == version {
			return v
		}
	}

	require.FailNowf(t, "Unable to find version", "Trying to find %s/%s", crd.GetName(), version)
	return apiextensions.CustomResourceDefinitionVersion{}
}
