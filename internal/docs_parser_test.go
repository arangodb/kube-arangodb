//
// DISCLAIMER
//
// Copyright 2023-2025 ArangoDB GmbH, Cologne, Germany
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
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	goStrings "strings"
	"testing"

	"github.com/stretchr/testify/require"
	openapi "k8s.io/kube-openapi/pkg/common"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	rootPackageName = "github.com/arangodb/kube-arangodb"
)

func parseDocDefinition(t *testing.T, root, path, typ string, field *ast.Field, fs *token.FileSet) DocDefinition {
	def := DocDefinition{
		Path: path,
		Type: typ,
	}

	require.NotNil(t, field)

	if links, ok := extract(field, "link"); ok {
		def.Links = links
	}

	if d, ok := extract(field, "default"); ok {
		def.Default = util.NewType[string](d[0])
	}

	if example, ok := extract(field, "example"); ok {
		def.Example = example
	}

	if enum, ok := extract(field, "enum"); ok {
		def.Enum = enum
	}

	if immutable, ok := extract(field, "immutable"); ok {
		def.Immutable = util.NewType[string](immutable[0])
	}

	if important, ok := extract(field, "important"); ok {
		def.Important = util.NewType[string](important[0])
	}

	if docs := extractDocumentation(field); len(docs) > 0 {
		def.Docs = docs
	}

	if grade, err := extractGrade(field); err != nil {
		require.Fail(t, fmt.Sprintf("Error while getting grade for %s: %s", path, err.Error()))
	} else {
		def.Grade = grade
	}

	file := fs.File(field.Pos())

	filePath, err := filepath.Rel(root, file.Name())
	require.NoError(t, err)

	def.File = filePath
	def.Line = file.Line(field.Pos())

	return def
}

type typeInfo struct {
	path string
	typ  string
}

func iterateOverObject(t *testing.T, fields map[string]*ast.Field, name string, object reflect.Type, path string) map[typeInfo]*ast.Field {
	r := map[typeInfo]*ast.Field{}
	t.Run(name, func(t *testing.T) {
		for k, v := range iterateOverObjectDirect(t, fields, name, object, path) {
			r[k] = v
		}
	})

	return r
}

func iterateOverObjectDirect(t *testing.T, fields map[string]*ast.Field, name string, object reflect.Type, path string) map[typeInfo]*ast.Field {
	if n, _, simple := isSimpleType(object); simple {
		return map[typeInfo]*ast.Field{
			typeInfo{
				path: fmt.Sprintf("%s.%s", path, name),
				typ:  n,
			}: nil,
		}
	}

	r := map[typeInfo]*ast.Field{}

	switch object.Kind() {
	case reflect.Array, reflect.Slice:
		if _, _, simple := isSimpleType(object.Elem()); simple {
			return map[typeInfo]*ast.Field{
				typeInfo{
					path: fmt.Sprintf("%s.%s", path, name),
					typ:  "array",
				}: nil,
			}
		}

		for k, v := range iterateOverObjectDirect(t, fields, fmt.Sprintf("%s\\[int\\]", name), object.Elem(), path) {
			r[k] = v
		}
	case reflect.Map:
		if _, _, simple := isSimpleType(object.Elem()); simple {
			return map[typeInfo]*ast.Field{
				typeInfo{
					path: fmt.Sprintf("%s.%s", path, name),
					typ:  "object",
				}: nil,
			}
		}

		for k, v := range iterateOverObjectDirect(t, fields, fmt.Sprintf("%s.\\<%s\\>", name, object.Key().Kind().String()), object.Elem(), path) {
			r[k] = v
		}
	case reflect.Struct:
		for field := 0; field < object.NumField(); field++ {
			f := object.Field(field)

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

			fullFieldName := fmt.Sprintf("%s.%s.%s", object.PkgPath(), object.Name(), f.Name)

			doc, ok := fields[fullFieldName]
			if !ok && !f.Anonymous {
				require.True(t, ok, "field %s was not parsed from source", fullFieldName)
			}

			if !f.Anonymous {
				if t, ok := extractType(doc); ok {
					info := typeInfo{
						path: fmt.Sprintf("%s.%s.%s", path, name, n),
						typ:  t[0],
					}
					r[info] = doc
					continue
				}
			}

			// inline and anonymous field (embedded)
			if inline && n == "" {
				if doc != nil {
					if t, ok := extractType(doc); ok {
						info := typeInfo{
							path: fmt.Sprintf("%s.%s", path, name),
							typ:  t[0],
						}
						r[info] = doc
						continue
					}
				}
			}

			if inline {
				for k, v := range iterateOverObjectDirect(t, fields, name, f.Type, path) {
					if v == nil {
						v = doc
					}
					r[k] = v
				}
			} else {

				for k, v := range iterateOverObject(t, fields, n, f.Type, fmt.Sprintf("%s.%s", path, name)) {
					if v == nil {
						v = doc
					}
					r[k] = v
				}
			}
		}
	case reflect.Pointer:
		for k, v := range iterateOverObjectDirect(t, fields, name, object.Elem(), path) {
			r[k] = v
		}
	default:
		require.Failf(t, "unsupported type", "%s for %s at %s", object.String(), name, path)
	}

	return r
}

func extractType(n *ast.Field) ([]string, bool) {
	return extract(n, "type")
}

func extract(n *ast.Field, tag string) ([]string, bool) {
	if n.Doc == nil {
		return nil, false
	}

	var ret []string

	for _, c := range n.Doc.List {
		if goStrings.HasPrefix(c.Text, fmt.Sprintf("// +doc/%s: ", tag)) {
			ret = append(ret, goStrings.TrimPrefix(c.Text, fmt.Sprintf("// +doc/%s: ", tag)))
		}
	}

	return ret, len(ret) > 0
}

func extractGrade(n *ast.Field) (*DocDefinitionGradeDefinition, error) {
	deprecatedGrade, err := extractDeprecated(n)
	if err != nil {
		return nil, err
	}

	var grade *DocDefinitionGradeDefinition

	if v, ok := extract(n, "grade"); ok {
		grade, err = NewDocDefinitionGradeDefinition(v...)
		if err != nil {
			return nil, err
		}
	}

	if deprecatedGrade != nil && grade != nil {
		return nil, errors.Errorf("Only one way of defining grade should be visible")
	}

	if deprecatedGrade != nil {
		return deprecatedGrade, nil
	}

	if grade != nil {
		return grade, nil
	}

	return nil, nil
}

func extractDeprecated(n *ast.Field) (*DocDefinitionGradeDefinition, error) {
	if n == nil || n.Doc == nil {
		return nil, nil
	}
	for _, c := range n.Doc.List {
		if goStrings.HasPrefix(c.Text, "// ") {
			if goStrings.HasPrefix(c.Text, "// Deprecated") {
				if !goStrings.HasPrefix(c.Text, "// Deprecated: ") {
					return nil, errors.Errorf("Invalid deprecated field")
				}
			}
			if goStrings.HasPrefix(c.Text, "// Deprecated:") {
				return &DocDefinitionGradeDefinition{
					Grade:   DocDefinitionGradeDeprecated,
					Message: []string{goStrings.TrimSpace(goStrings.TrimPrefix(c.Text, "// Deprecated:"))},
				}, nil
			}
		}
	}

	return nil, nil
}

func extractDocumentation(n *ast.Field) []string {
	if n.Doc == nil {
		return nil
	}

	var ret []string

	for _, c := range n.Doc.List {
		if goStrings.HasPrefix(c.Text, "// ") {
			if goStrings.HasPrefix(c.Text, "// +doc/") {
				continue
			}
			if goStrings.HasPrefix(c.Text, "// Deprecated") {
				continue
			}
			ret = append(ret, goStrings.TrimSpace(goStrings.TrimPrefix(c.Text, "// ")))
		}
	}

	return ret
}

// isSimpleType returns the OpenAPI-compatible type name, type format and boolean indicating if this is simple type or not
func isSimpleType(obj reflect.Type) (string, string, bool) {
	typ, frmt := openapi.OpenAPITypeFormat(obj.Kind().String())
	return typ, frmt, typ != "" || frmt != ""
}

func extractTag(tag string) (string, bool) {
	parts := goStrings.Split(tag, ",")

	if len(parts) == 1 {
		return parts[0], false
	}

	if parts[1] == "inline" {
		return parts[0], true
	}

	return parts[0], false
}

// parseSourceFiles returns map of <path to field in structure> -> AST for structure Field and the token inspector for all files in package
func parseSourceFiles(t *testing.T, root string, fset *token.FileSet, path string) map[string]*ast.Field {
	d := parseMultipleDirs(t, root, fset, parser.ParseComments, path)

	r := map[string]*ast.Field{}

	for k, f := range d {
		var ct *ast.TypeSpec
		var nt *ast.TypeSpec

		ast.Inspect(f, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.TypeSpec, *ast.FuncDecl, *ast.Field, *ast.Package, *ast.File, *ast.Ident, *ast.StructType:
			default:
				if x == nil {
					return true
				}
				return true
			}

			switch x := n.(type) {
			case *ast.TypeSpec:
				ct = x
			case *ast.StructType:
				nt = ct
			case *ast.FuncDecl:
				nt = nil
			case *ast.Field:
				if nt != nil {
					require.NotEmpty(t, nt.Name)

					if len(x.Names) > 0 {
						for _, name := range x.Names {
							r[fmt.Sprintf("%s.%s.%s", k, nt.Name, name)] = x
						}
					} else {
						// If x.Names is empty, it's an anonymous field
						if len(x.Names) == 0 {
							// first check if it's a pointer to a struct
							typeName, ok := x.Type.(*ast.StarExpr)
							if ok {
								ident, ok := typeName.X.(*ast.SelectorExpr)
								if ok {
									fieldName := ident.Sel.Name
									r[fmt.Sprintf("%s.%s.%s", k, nt.Name, fieldName)] = x
								}
							} else {
								// if it's not a pointer
								ident, ok := x.Type.(*ast.SelectorExpr)
								if ok {
									fieldName := ident.Sel.Name
									r[fmt.Sprintf("%s.%s.%s", k, nt.Name, fieldName)] = x
								}
							}
						}
					}

				}
			}

			return true
		})
	}

	return r
}

func parseMultipleDirs(t *testing.T, root string, fset *token.FileSet, mode parser.Mode, dirs ...string) map[string]*ast.Package {
	// positions are relative to fset

	r := map[string]*ast.Package{}

	for _, dir := range dirs {
		d, err := parser.ParseDir(fset, dir, func(info fs.FileInfo) bool {
			return !goStrings.HasSuffix(info.Name(), "_test.go") &&
				info.Name() != "zz_generated.deepcopy.go"
		}, mode)
		require.NoError(t, err)

		require.Len(t, d, 1)
		k := goStrings.ReplaceAll(dir, root, rootPackageName)

		for _, v := range d {
			require.NotContains(t, r, k)
			r[k] = v
		}
	}

	return r
}
