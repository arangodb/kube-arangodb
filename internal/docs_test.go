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
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type DocDefinitions []DocDefinition

func (d DocDefinitions) Render(t *testing.T) []byte {
	out := bytes.NewBuffer(nil)

	for _, el := range d {

		write(t, out, "### %s: %s\n\n", el.Path, el.Type)

		if d := el.Important; d != nil {
			write(t, out, "**Important**: %s\n\n", *d)
		}

		if len(el.Docs) > 0 {
			for _, doc := range el.Docs {
				write(t, out, "%s\n", doc)
			}
			write(t, out, "\n")
		}

		if len(el.Links) > 0 {
			write(t, out, "Links:\n")

			for _, link := range el.Links {
				z := strings.Split(link, "|")
				if len(z) == 1 {
					write(t, out, "* [Documentation](%s)\n", z[0])
				} else if len(z) == 2 {
					write(t, out, "* [%s](%s)\n", z[0], z[1])
				} else {
					require.Fail(t, "Invalid link format")
				}
			}

			write(t, out, "\n")
		}

		if len(el.Example) > 0 {
			write(t, out, "Example:\n")
			write(t, out, "```yaml\n")
			for _, example := range el.Example {
				write(t, out, "%s\n", example)
			}
			write(t, out, "```\n\n")
		}

		if len(el.Enum) > 0 {
			write(t, out, "Possible Values: \n")
			for id, enum := range el.Enum {
				z := strings.Split(enum, "|")

				if id == 0 {
					z[0] = fmt.Sprintf("%s (default)", z[0])
				}

				if len(z) == 1 {
					write(t, out, "* %s\n", z[0])
				} else if len(z) == 2 {
					write(t, out, "* %s - %s\n", z[0], z[1])
				} else {
					require.Fail(t, "Invalid enum format")
				}
			}
			write(t, out, "\n")
		} else {
			if d := el.Default; d != nil {
				write(t, out, "Default Value: %s\n\n", *d)
			}
		}

		if d := el.Immutable; d != nil {
			write(t, out, "This field is **immutable**: %s\n\n", *d)
		}

		write(t, out, "[Code Reference](/%s#L%d)\n\n", el.File, el.Line)
	}

	return out.Bytes()
}

type DocDefinition struct {
	Path string
	Type string

	File string
	Line int

	Docs []string

	Links []string

	Important *string

	Enum []string

	Immutable *string

	Default *string
	Example []string
}

func Test_GenerateAPIDocs(t *testing.T) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	docs := map[string]map[string]interface{}{
		"ArangoDeployment.V1": {
			"Spec": api.ArangoDeployment{}.Spec,
		},
		"ArangoMember.V1": {
			"Spec": api.ArangoMember{}.Spec,
		},
	}
	resultPaths := generateDocs(t, docs, fmt.Sprintf("%s/pkg/apis/deployment/v1", root))

	generateIndex(t, resultPaths)
}

func generateDocs(t *testing.T, objects map[string]map[string]interface{}, paths ...string) map[string]string {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	docs, fs := getDocs(t, paths...)
	outPaths := make(map[string]string)

	for object, sections := range objects {
		t.Run(object, func(t *testing.T) {
			renderSections := map[string][]byte{}
			for section, data := range sections {
				t.Run(section, func(t *testing.T) {

					res := iterateOverObject(t, docs, strings.ToLower(section), reflect.TypeOf(data), "")

					var elements []string

					for k := range res {
						elements = append(elements, k)
					}

					sort.Slice(elements, func(i, j int) bool {
						if a, b := strings.ToLower(elements[i]), strings.ToLower(elements[j]); a == b {
							return elements[i] < elements[j]
						} else {
							return a < b
						}
					})

					defs := make(DocDefinitions, len(elements))

					for id, k := range elements {
						field := res[k]

						var def DocDefinition

						def.Path = strings.Split(k, ":")[0]
						def.Type = strings.Split(k, ":")[1]

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

						defs[id] = def
					}

					renderSections[section] = defs.Render(t)
				})
			}

			fileName := fmt.Sprintf("%s.md", object)
			outPaths[object] = fileName
			outPath := path.Join(root, "docs/api", fmt.Sprintf("%s.md", object))
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			require.NoError(t, err)

			defer func() {
				require.NoError(t, out.Close())
			}()

			write(t, out, "# API Reference for %s\n\n", strings.ReplaceAll(object, ".", " "))

			for name, section := range renderSections {
				write(t, out, "## %s\n\n", name)

				_, err = out.Write(section)
				require.NoError(t, err)
			}
		})
	}
	return outPaths
}

func generateIndex(t *testing.T, apiDocs map[string]string) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)
	outPath := path.Join(root, "docs/api/README.md")

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, out.Close())
	}()

	write(t, out, "# Custom Resources API Reference\n\n")

	for name, filePath := range apiDocs {
		write(t, out, " - [%s](./%s)\n", name, filePath)
	}
	write(t, out, "\n")
}

func write(t *testing.T, out io.Writer, format string, args ...interface{}) {
	_, err := out.Write([]byte(fmt.Sprintf(format, args...)))
	require.NoError(t, err)
}

func iterateOverObject(t *testing.T, docs map[string]*ast.Field, name string, object reflect.Type, path string) map[string]*ast.Field {
	r := map[string]*ast.Field{}
	t.Run(name, func(t *testing.T) {
		for k, v := range iterateOverObjectDirect(t, docs, name, object, path) {
			r[k] = v
		}
	})

	return r
}

func iterateOverObjectDirect(t *testing.T, docs map[string]*ast.Field, name string, object reflect.Type, path string) map[string]*ast.Field {
	if n, simple := isSimpleType(object); simple {
		return map[string]*ast.Field{
			fmt.Sprintf("%s.%s:%s", path, name, n): nil,
		}
	}

	r := map[string]*ast.Field{}

	switch object.Kind() {
	case reflect.Array, reflect.Slice:
		if n, simple := isSimpleType(object.Elem()); simple {
			return map[string]*ast.Field{
				fmt.Sprintf("%s.%s:[]%s", path, name, n): nil,
			}
		}

		for k, v := range iterateOverObjectDirect(t, docs, fmt.Sprintf("%s\\[int\\]", name), object.Elem(), path) {
			r[k] = v
		}
	case reflect.Map:
		if n, simple := isSimpleType(object.Elem()); simple {
			return map[string]*ast.Field{
				fmt.Sprintf("%s.%s:map[%s]%s", path, name, object.Key().String(), n): nil,
			}
		}

		for k, v := range iterateOverObjectDirect(t, docs, fmt.Sprintf("%s.\\<%s\\>", name, object.Key().Kind().String()), object.Elem(), path) {
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

			docName := fmt.Sprintf("%s.%s", object.String(), f.Name)

			doc, ok := docs[docName]
			if !ok && !f.Anonymous {
				require.True(t, ok, docName, f.Name)
			}

			if !f.Anonymous {
				if t, ok := extractType(doc); ok {
					r[fmt.Sprintf("%s.%s.%s:%s", path, name, n, t[0])] = doc
					continue
				}
			}

			if inline {
				for k, v := range iterateOverObjectDirect(t, docs, name, f.Type, path) {
					if v == nil {
						v = doc
					}
					r[k] = v
				}
			} else {

				for k, v := range iterateOverObject(t, docs, n, f.Type, fmt.Sprintf("%s.%s", path, name)) {
					if v == nil {
						v = doc
					}
					r[k] = v
				}
			}
		}
	case reflect.Pointer:
		for k, v := range iterateOverObjectDirect(t, docs, name, object.Elem(), path) {
			r[k] = v
		}
	default:
		require.Fail(t, object.String())
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

func isSimpleType(obj reflect.Type) (string, bool) {
	switch obj.Kind() {
	case reflect.String, reflect.Int64, reflect.Bool, reflect.Int, reflect.Uint16, reflect.Int32:
		return obj.Kind().String(), true
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

func getDocs(t *testing.T, paths ...string) (map[string]*ast.Field, *token.FileSet) {
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
			_, ok := r[k]
			require.False(t, ok)
			r[k] = v
		}
	}

	return r, fset
}
