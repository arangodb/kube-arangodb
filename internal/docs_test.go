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
	"io"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func generateDocs(t *testing.T, objects map[string]map[string]interface{}, fields map[string]*ast.Field, fs *token.FileSet) map[string]string {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	outPaths := make(map[string]string)

	for objectName, sections := range objects {
		t.Run(objectName, func(t *testing.T) {
			renderSections := map[string][]byte{}
			for section, fieldInstance := range sections {
				t.Run(section, func(t *testing.T) {

					sectionParsed := iterateOverObject(t, fields, strings.ToLower(section), reflect.TypeOf(fieldInstance), "")

					defs := parseDocDefinitions(t, sectionParsed, fs)

					renderSections[section] = defs.RenderMarkdown(t)
				})
			}

			fileName := fmt.Sprintf("%s.md", objectName)
			outPaths[objectName] = fileName
			outPath := path.Join(root, "docs/api", fmt.Sprintf("%s.md", objectName))
			out, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			require.NoError(t, err)

			defer func() {
				require.NoError(t, out.Close())
			}()

			write(t, out, "# API Reference for %s\n\n", strings.ReplaceAll(objectName, ".", " "))

			util.IterateSorted(renderSections, func(name string, section []byte) {
				write(t, out, "## %s\n\n", name)

				_, err = out.Write(section)
				require.NoError(t, err)
			})
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

	util.IterateSorted(apiDocs, func(name string, filePath string) {
		write(t, out, " - [%s](./%s)\n", name, filePath)
	})
	write(t, out, "\n")
}

func write(t *testing.T, out io.Writer, format string, args ...interface{}) {
	_, err := out.Write([]byte(fmt.Sprintf(format, args...)))
	require.NoError(t, err)
}
