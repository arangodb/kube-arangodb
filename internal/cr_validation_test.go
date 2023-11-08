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

package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	appsv1 "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	backupv1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	deploymentv1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	deploymentv2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v2alpha1"
	mlv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	replicationv1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	replicationv2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v2alpha1"
	storagev1alpha "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
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
		obj interface{}
	}

	// CR file prefix -> packages to parse -> versions -> obj
	input := map[string]map[string]map[string]genSpec{
		"apps-job": {
			fmt.Sprintf("%s/pkg/apis/apps/v1", root): {
				"v1": {
					appsv1.ArangoJob{}.Spec,
				},
			},
		},
		"backups-backup": {
			fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
				"v1": {
					backupv1.ArangoBackup{}.Spec,
				},
				"v1alpha": {
					backupv1.ArangoBackup{}.Spec,
				},
			},
		},
		"backups-backuppolicy": {
			fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
				"v1": {
					backupv1.ArangoBackupPolicy{}.Spec,
				},
				"v1alpha": {
					backupv1.ArangoBackupPolicy{}.Spec,
				},
			},
		},
		"database-deployment": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					deploymentv1.ArangoDeployment{}.Spec,
				},
				"v1alpha": {
					deploymentv1.ArangoDeployment{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					deploymentv2alpha1.ArangoDeployment{}.Spec,
				},
			},
		},
		"database-member": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					deploymentv1.ArangoMember{}.Spec,
				},
				"v1alpha": {
					deploymentv1.ArangoMember{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					deploymentv2alpha1.ArangoMember{}.Spec,
				},
			},
		},
		"database-clustersynchronization": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					deploymentv1.ArangoClusterSynchronization{}.Spec,
				},
				"v1alpha": {
					deploymentv1.ArangoClusterSynchronization{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					deploymentv2alpha1.ArangoClusterSynchronization{}.Spec,
				},
			},
		},
		"database-task": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					deploymentv1.ArangoTask{}.Spec,
				},
				"v1alpha": {
					deploymentv1.ArangoTask{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					deploymentv2alpha1.ArangoTask{}.Spec,
				},
			},
		},
		"replication-deploymentreplication": {
			fmt.Sprintf("%s/pkg/apis/replication/v1", root): {
				"v1": {
					replicationv1.ArangoDeploymentReplication{}.Spec,
				},
				"v1alpha": {
					replicationv1.ArangoDeploymentReplication{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/replication/v2alpha1", root): {
				"v2alpha1": {
					replicationv2alpha1.ArangoDeploymentReplication{}.Spec,
				},
			},
		},
		"storage-localstorage": {
			fmt.Sprintf("%s/pkg/apis/storage/v1alpha", root): {
				"v1alpha": {
					storagev1alpha.ArangoLocalStorage{}.Spec,
				},
			},
		},
		"ml-extension": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					mlv1alpha1.ArangoMLExtension{}.Spec,
				},
			},
		},
		"ml-storage": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					mlv1alpha1.ArangoMLStorage{}.Spec,
				},
			},
		},
		"ml-job-cron": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					mlv1alpha1.ArangoMLCronJob{}.Spec,
				},
			},
		},
		"ml-job-batch": {
			fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
				"v1alpha1": {
					mlv1alpha1.ArangoMLBatchJob{}.Spec,
				},
			},
		},
	}

	for filePrefix, packagesToVersion := range input {
		validationPerVersion := make(map[string]apiextensions.CustomResourceValidation, len(packagesToVersion))
		for apiDir, versionMap := range packagesToVersion {
			fields, fileSets := parseSourceFiles(t, apiDir)

			for version, generationSpec := range versionMap {
				sb := newSchemaBuilder(root, fields, fileSets)
				s := sb.TypeToSchema(t, reflect.TypeOf(generationSpec.obj), ".spec")
				require.NotNil(t, s, version)

				validationPerVersion[version] = apiextensions.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
						Type:                   "object",
						Properties:             map[string]apiextensions.JSONSchemaProps{"spec": *s},
						XPreserveUnknownFields: util.NewType(true),
					},
				}
			}
		}

		outPath := path.Join(root, "pkg/crd/crds", fmt.Sprintf("%s.schema.generated.json", filePrefix))
		out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, out.Close())
		}()

		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		err = enc.Encode(validationPerVersion)
		require.NoError(t, err)
	}
}
