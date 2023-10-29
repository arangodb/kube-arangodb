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
	"go/ast"
	"go/token"
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
	replicationv1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	replicationv2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/replication/v2alpha1"
	storagev1alpha "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (def DocDefinition) ApplyToSchema(docName string, s *apiextensions.JSONSchemaProps) {
	for _, e := range def.Enum {
		z := strings.Split(e, "|")
		s.Enum = append(s.Enum, apiextensions.JSON{
			Raw: []byte("\"" + z[0] + "\""),
		})
	}

	if def.Immutable != nil {
		s.XValidations = append(s.XValidations, apiextensions.ValidationRule{
			Rule:    fmt.Sprintf("self%s == oldSelf%s", def.Path, def.Path),
			Message: fmt.Sprintf("field %s is immutable", strings.TrimPrefix(def.Path, ".")),
		})
	}

	if len(docName) > 0 {
		s.ExternalDocs = &apiextensions.ExternalDocumentation{
			URL: getDocsLinkForField(docName, def),
		}
	}
	s.Description = strings.Join(def.Docs, "\n")
}

func Test_GenerateCRValidationSchemas(t *testing.T) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	type genSpec struct {
		docName string
		obj     interface{}
	}

	// CR file prefix -> packages to parse -> versions -> docName and obj
	input := map[string]map[string]map[string]genSpec{ // TODO: consider moving this into new YAML file which will describe CRD metadata
		"apps-job": {
			fmt.Sprintf("%s/pkg/apis/apps/v1", root): {
				"v1": {
					"",
					appsv1.ArangoJob{}.Spec,
				},
			},
		},
		"backups-backup": {
			fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
				"v1": {
					"ArangoBackup.V1",
					backupv1.ArangoBackup{}.Spec,
				},
				"v1alpha": {
					"ArangoBackup.V1",
					backupv1.ArangoBackup{}.Spec,
				},
			},
		},
		"backups-backuppolicy": {
			fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
				"v1": {
					"ArangoBackupPolicy.V1",
					backupv1.ArangoBackupPolicy{}.Spec,
				},
				"v1alpha": {
					"ArangoBackupPolicy.V1",
					backupv1.ArangoBackupPolicy{}.Spec,
				},
			},
		},
		"database-deployment": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					"ArangoDeployment.V1",
					deploymentv1.ArangoDeployment{}.Spec,
				},
				"v1alpha": {
					"ArangoDeployment.V1",
					deploymentv1.ArangoDeployment{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					"",
					deploymentv2alpha1.ArangoDeployment{}.Spec,
				},
			},
		},
		"database-member": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					"ArangoMember.V1",
					deploymentv1.ArangoMember{}.Spec,
				},
				"v1alpha": {
					"ArangoMember.V1",
					deploymentv1.ArangoMember{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					"",
					deploymentv2alpha1.ArangoMember{}.Spec,
				},
			},
		},
		"database-clustersynchronization": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					"",
					deploymentv1.ArangoClusterSynchronization{}.Spec,
				},
				"v1alpha": {
					"",
					deploymentv1.ArangoClusterSynchronization{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					"",
					deploymentv2alpha1.ArangoClusterSynchronization{}.Spec,
				},
			},
		},
		"database-task": {
			fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
				"v1": {
					"",
					deploymentv1.ArangoTask{}.Spec,
				},
				"v1alpha": {
					"",
					deploymentv1.ArangoTask{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/deployment/v2alpha1", root): {
				"v2alpha1": {
					"",
					deploymentv2alpha1.ArangoTask{}.Spec,
				},
			},
		},
		"replication-deploymentreplication": {
			fmt.Sprintf("%s/pkg/apis/replication/v1", root): {
				"v1": {
					"ArangoDeploymentReplication.V1",
					replicationv1.ArangoDeploymentReplication{}.Spec,
				},
				"v1alpha": {
					"ArangoDeploymentReplication.V1",
					replicationv1.ArangoDeploymentReplication{}.Spec,
				},
			},
			fmt.Sprintf("%s/pkg/apis/replication/v2alpha1", root): {
				"v2alpha1": {
					"",
					replicationv2alpha1.ArangoDeploymentReplication{}.Spec,
				},
			},
		},
		"storage-localstorage": {
			fmt.Sprintf("%s/pkg/apis/storage/v1alpha", root): {
				"v1alpha": {
					"ArangoLocalStorage.V1Alpha",
					storagev1alpha.ArangoLocalStorage{}.Spec,
				},
			},
		},
	}

	// TODO: consider using "sigs.k8s.io/controller-tools/pkg/crd" for parsing instead
	for filePrefix, packagesToVersion := range input {
		validationPerVersion := make(map[string]apiextensions.CustomResourceValidation, len(packagesToVersion))
		for apiDir, versionMap := range packagesToVersion {
			fields, fileSets := parseSourceFiles(t, apiDir)

			for version, generationSpec := range versionMap {
				sb := newSchemaBuilder(root, generationSpec.docName, fields, fileSets)
				s := sb.TypeToSchema(t, reflect.TypeOf(generationSpec.obj), ".spec")
				require.NotNil(t, s, version)

				validationPerVersion[version] = apiextensions.CustomResourceValidation{
					OpenAPIV3Schema: &apiextensions.JSONSchemaProps{
						Type:                   "object",
						XPreserveUnknownFields: util.NewType(true),
						Properties:             map[string]apiextensions.JSONSchemaProps{"spec": *s},
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

type schemaBuilder struct {
	root    string
	docName string
	fields  map[string]*ast.Field
	fs      *token.FileSet
}

func newSchemaBuilder(root, docName string, fields map[string]*ast.Field, fs *token.FileSet) *schemaBuilder {
	return &schemaBuilder{
		root:    root,
		docName: docName,
		fields:  fields,
		fs:      fs,
	}
}

func (b *schemaBuilder) TypeToSchema(t *testing.T, obj reflect.Type, path string) *apiextensions.JSONSchemaProps {
	var schema *apiextensions.JSONSchemaProps
	t.Run(obj.Name(), func(t *testing.T) {
		switch obj.Kind() {
		case reflect.Pointer:
			schema = b.TypeToSchema(t, obj.Elem(), path)
		case reflect.Struct:
			schema = b.StructToSchema(t, obj, path)
		case reflect.Array, reflect.Slice:
			schema = b.ArrayToSchema(t, obj.Elem(), path)
		case reflect.Map:
			schema = b.MapToSchema(t, obj, path)
		default:
			if typ, frmt, simple := b.getTypeFormat(obj); simple {
				schema = &apiextensions.JSONSchemaProps{
					Type:   typ,
					Format: frmt,
				}
			} else {
				// TODO: consider using https://kubernetesjsonschema.dev/ for k8s resources validation

				t.Fatalf("Unsupported obj kind: %s", obj.Kind())
				return
			}
		}
	})
	return schema
}

func (b *schemaBuilder) lookupDefinition(t *testing.T, fullName, path string) *DocDefinition {
	f := b.fields[fullName]
	if f == nil {
		return nil
	}

	d := parseDocDefinition(t, b.root, path, "", f, b.fs)
	return &d
}

func (b *schemaBuilder) ArrayToSchema(t *testing.T, elemObj reflect.Type, path string) *apiextensions.JSONSchemaProps {
	isByteArray := elemObj.Kind() == reflect.Uint8
	if isByteArray {
		return &apiextensions.JSONSchemaProps{
			Type:   "string",
			Format: "byte",
		}
	}

	return &apiextensions.JSONSchemaProps{
		Type: "array",
		Items: &apiextensions.JSONSchemaPropsOrArray{
			Schema: b.TypeToSchema(t, elemObj, path),
		},
	}
}

func (b *schemaBuilder) MapToSchema(t *testing.T, mapObj reflect.Type, path string) *apiextensions.JSONSchemaProps {
	require.Equal(t, reflect.String, mapObj.Key().Kind(), "only string keys for map are supported %s", path)

	return &apiextensions.JSONSchemaProps{
		Type: "object",
		AdditionalProperties: &apiextensions.JSONSchemaPropsOrBool{
			Schema: b.TypeToSchema(t, mapObj.Elem(), path),
			Allows: true, /* set automatically by serialization, but useful for testing */
		},
	}
}

func (b *schemaBuilder) StructToSchema(t *testing.T, structObj reflect.Type, path string) *apiextensions.JSONSchemaProps {
	schema := &apiextensions.JSONSchemaProps{
		Type:       "object",
		Properties: make(map[string]apiextensions.JSONSchemaProps),
	}

	for field := 0; field < structObj.NumField(); field++ {
		f := structObj.Field(field)

		if !f.IsExported() {
			continue
		}

		tag, ok := f.Tag.Lookup("json")
		if !ok {
			if f.Anonymous {
				tag = ",inline"
			}
		}

		n, inline := extractTag(tag)
		if n == "-" {
			continue
		}

		if !inline {
			path = fmt.Sprintf("%s.%s", path, n)
		}

		s := b.TypeToSchema(t, f.Type, path)
		require.NotNil(t, s, path)

		fullFieldName := fmt.Sprintf("%s.%s", structObj.String(), f.Name)
		def := b.lookupDefinition(t, fullFieldName, path)
		if def != nil {
			def.ApplyToSchema(b.docName, s)
		}

		schema.Properties[n] = *s
	}
	return schema
}

func (b *schemaBuilder) getTypeFormat(obj reflect.Type) (string, string, bool) {
	switch obj.Kind() {
	case reflect.String:
		return "string", "", true
	case reflect.Bool:
		return "boolean", "", true
	case reflect.Int, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16:
		return "integer", "int32", true
	case reflect.Uint64:
		return "integer", "int32", true
	case reflect.Float32:
		return "number", "float", true
	}

	return "", "", false
}
