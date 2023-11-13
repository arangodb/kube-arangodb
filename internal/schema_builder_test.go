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
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	openapi "k8s.io/kube-openapi/pkg/common"
)

type schemaBuilder struct {
	root   string
	fields map[string]*ast.Field
	fs     *token.FileSet
}

func newSchemaBuilder(root string, fields map[string]*ast.Field, fs *token.FileSet) *schemaBuilder {
	return &schemaBuilder{
		root:   root,
		fields: fields,
		fs:     fs,
	}
}

func (b *schemaBuilder) tryGetKubeOpenAPIDefinitions(t *testing.T, obj reflect.Type) *apiextensions.JSONSchemaProps {
	if o, ok := reflect.New(obj).Interface().(openapi.OpenAPIV3DefinitionGetter); ok {
		return b.openAPIDefToSchemaPros(t, o.OpenAPIV3Definition())
	}
	if o, ok := reflect.New(obj).Interface().(openapi.OpenAPIDefinitionGetter); ok {
		return b.openAPIDefToSchemaPros(t, o.OpenAPIDefinition())
	}

	type openAPISchemaTypeGetter interface {
		OpenAPISchemaType() []string
	}
	type openAPISchemaFormatGetter interface {
		OpenAPISchemaFormat() string
	}
	var typ, frmt string
	if o, ok := reflect.New(obj).Interface().(openAPISchemaTypeGetter); ok {
		strs := o.OpenAPISchemaType()
		require.Len(t, strs, 1)
		typ = strs[0]
	}
	if o, ok := reflect.New(obj).Interface().(openAPISchemaFormatGetter); ok {
		frmt = o.OpenAPISchemaFormat()
	}
	if typ != "" || frmt != "" {
		return &apiextensions.JSONSchemaProps{
			Type:   typ,
			Format: frmt,
		}
	}
	return nil
}

func (b *schemaBuilder) openAPIDefToSchemaPros(t *testing.T, _ *openapi.OpenAPIDefinition) *apiextensions.JSONSchemaProps {
	require.Fail(t, "openAPIDefToSchemaPros is not implemented because there were no calls to this function. Add the impl if needed.")
	return nil
}

func (b *schemaBuilder) TypeToSchema(t *testing.T, obj reflect.Type, path string) *apiextensions.JSONSchemaProps {
	var schema *apiextensions.JSONSchemaProps
	t.Run(obj.Name(), func(t *testing.T) {
		// first check if type already implements a method to get OpenAPI schema:
		schema = b.tryGetKubeOpenAPIDefinitions(t, obj)
		if schema != nil {
			return
		}

		// fallback to our impl:
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
			if typ, frmt, simple := isSimpleType(obj); simple {
				schema = &apiextensions.JSONSchemaProps{
					Type:   typ,
					Format: frmt,
				}
			} else {
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
			} else {
				require.Failf(t, "field %s.%s has no valid json tag: can't build schema", path, f.Name)
			}
		}

		n, inline := extractTag(tag)
		if n == "-" {
			continue
		}

		p := path
		if !inline {
			p = fmt.Sprintf("%s.%s", path, n)
		}

		s := b.TypeToSchema(t, f.Type, p)
		require.NotNil(t, s, p)

		fullFieldName := fmt.Sprintf("%s.%s", structObj.String(), f.Name)
		def := b.lookupDefinition(t, fullFieldName, p)
		if def != nil {
			def.ApplyToSchema(s)
		}

		if inline {
			// merge into parent
			for k, v := range s.Properties {
				schema.Properties[k] = v
			}
		} else {
			require.NotEmpty(t, n, fullFieldName)
			schema.Properties[n] = *s
		}
	}
	return schema
}
