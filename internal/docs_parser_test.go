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
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func parseDocDefinitions(t *testing.T, res map[typeInfo]*ast.Field, fs *token.FileSet) DocDefinitions {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	defs := make(DocDefinitions, 0, len(res))

	for info, field := range res {
		def := DocDefinition{
			Path: info.path,
			Type: info.typ,
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

		if docs, ok := extractNotTags(field); !ok {
			println(def.Path, " is missing documentation!")
		} else {
			def.Docs = docs
		}

		file := fs.File(field.Pos())

		filePath, err := filepath.Rel(root, file.Name())
		require.NoError(t, err)

		def.File = filePath
		def.Line = file.Line(field.Pos())

		defs = append(defs, def)
	}
	return defs
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
	if n, simple := isSimpleType(object); simple {
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
		if _, simple := isSimpleType(object.Elem()); simple {
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
		if _, simple := isSimpleType(object.Elem()); simple {
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

			fullFieldName := fmt.Sprintf("%s.%s", object.String(), f.Name)

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
		if strings.HasPrefix(c.Text, fmt.Sprintf("// +doc/%s: ", tag)) {
			ret = append(ret, strings.TrimPrefix(c.Text, fmt.Sprintf("// +doc/%s: ", tag)))
		}
	}

	return ret, len(ret) > 0
}

func extractNotTags(n *ast.Field) ([]string, bool) {
	if n.Doc == nil {
		return nil, false
	}

	var ret []string

	for _, c := range n.Doc.List {
		if strings.HasPrefix(c.Text, "// ") {
			if !strings.HasPrefix(c.Text, "// +doc/") {
				ret = append(ret, strings.TrimPrefix(c.Text, "// "))
			}
		}
	}

	return ret, len(ret) > 0
}

// isSimpleType returns the OpenAPI-compatible type name and boolean indicating if this is simple type or not
func isSimpleType(obj reflect.Type) (string, bool) {
	switch obj.Kind() {
	case reflect.String:
		return "string", true
	case reflect.Bool:
		return "boolean", true
	case reflect.Int, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16:
		return "integer", true
	case reflect.Int64, reflect.Uint64:
		return "integer", true
	case reflect.Float32:
		return "number", true
	}
	return "", false
}

func extractTag(tag string) (string, bool) {
	parts := strings.SplitN(tag, ",", 2)

	if len(parts) == 1 {
		return parts[0], false
	}

	if parts[1] == "inline" {
		return parts[0], true
	}

	return parts[0], false
}

// parseSourceFiles returns map of <path to field in structure> -> AST for structure Field and the token inspector for all files in package
func parseSourceFiles(t *testing.T, paths ...string) (map[string]*ast.Field, *token.FileSet) {
	d, fs := parseMultipleDirs(t, parser.ParseComments, paths...)

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

					for _, name := range x.Names {
						r[fmt.Sprintf("%s.%s.%s", k, nt.Name, name)] = x
					}
				}
			}

			return true
		})
	}

	return r, fs
}

func parseMultipleDirs(t *testing.T, mode parser.Mode, dirs ...string) (map[string]*ast.Package, *token.FileSet) {
	fset := token.NewFileSet() // positions are relative to fset

	r := map[string]*ast.Package{}

	for _, dir := range dirs {
		d, err := parser.ParseDir(fset, dir, func(info fs.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, mode)
		require.NoError(t, err)

		for k, v := range d {
			require.NotContains(t, r, k)
			r[k] = v
		}
	}

	return r, fset
}
