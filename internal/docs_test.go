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
	"go/token"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/require"

	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	replicationApi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	storageApi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func (d DocDefinitions) RenderMarkdown(t *testing.T, repositoryPath string) []byte {
	out := bytes.NewBuffer(nil)

	for i, el := range d {
		if i != 0 {
			write(t, out, "***\n\n")
		}

		write(t, out, "### %s\n\n", el.Path)
		write(t, out, "Type: `%s` <sup>[\\[ref\\]](%s/%s#L%d)</sup>\n\n", el.Type, repositoryPath, el.File, el.Line)

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
				write(t, out, "Default Value: `%s`\n\n", *d)
			}
		}

		if d := el.Immutable; d != nil {
			write(t, out, "This field is **immutable**: %s\n\n", *d)
		}
	}

	return out.Bytes()
}

func Test_GenerateAPIDocs(t *testing.T) {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	// package path -> result doc file name -> name of the top-level field to be described -> field instance for reflection
	input := map[string]map[string]map[string]interface{}{
		fmt.Sprintf("%s/pkg/apis/deployment/v1", root): {
			"ArangoDeployment.V1": {
				"Spec": deploymentApi.ArangoDeployment{}.Spec,
			},
			"ArangoMember.V1": {
				"Spec": deploymentApi.ArangoMember{}.Spec,
			},
		},
		fmt.Sprintf("%s/pkg/apis/backup/v1", root): {
			"ArangoBackup.V1": {
				"Spec":   backupApi.ArangoBackup{}.Spec,
				"Status": backupApi.ArangoBackup{}.Status,
			},
			"ArangoBackupPolicy.V1": {
				"Spec":   backupApi.ArangoBackupPolicy{}.Spec,
				"Status": backupApi.ArangoBackupPolicy{}.Status,
			},
		},
		fmt.Sprintf("%s/pkg/apis/ml/v1alpha1", root): {
			"ArangoMLExtension.V1Alpha1": {
				"Spec":   mlApi.ArangoMLExtension{}.Spec,
				"Status": mlApi.ArangoMLExtension{}.Status,
			},
			"ArangoMLStorage.V1Alpha1": {
				"Spec":   mlApi.ArangoMLStorage{}.Spec,
				"Status": mlApi.ArangoMLStorage{}.Status,
			},
			"ArangoMLCronJob.V1Alpha1": {
				"Spec":   mlApi.ArangoMLCronJob{}.Spec,
				"Status": mlApi.ArangoMLCronJob{}.Status,
			},
			"ArangoMLBatchJob.V1Alpha1": {
				"Spec":   mlApi.ArangoMLBatchJob{}.Spec,
				"Status": mlApi.ArangoMLBatchJob{}.Status,
			},
		},
		fmt.Sprintf("%s/pkg/apis/replication/v1", root): {
			"ArangoDeploymentReplication.V1": {
				"Spec": replicationApi.ArangoDeploymentReplication{}.Spec,
			},
		},
		fmt.Sprintf("%s/pkg/apis/storage/v1alpha", root): {
			"ArangoLocalStorage.V1Alpha": {
				"Spec": storageApi.ArangoLocalStorage{}.Spec,
			},
		},
	}

	resultPaths := make(map[string]string)
	for apiDir, docs := range input {
		fields, fileSets := parseSourceFiles(t, apiDir)
		util.CopyMap(resultPaths, generateDocs(t, docs, fields, fileSets))
	}
	generateIndex(t, resultPaths)
}

func prepareGitHubTreePath(t *testing.T, root string) string {
	vStr, err := os.ReadFile(filepath.Join(root, "VERSION"))
	require.NoError(t, err, "failed to read VERSION file")
	opVersion, err := semver.NewVersion(string(vStr))
	require.NoError(t, err)

	ref := fmt.Sprintf("%d.%d.%d", opVersion.Major, opVersion.Minor, opVersion.Patch)
	return fmt.Sprintf("https://github.com/arangodb/kube-arangodb/blob/%s", ref)
}

func generateDocs(t *testing.T, objects map[string]map[string]interface{}, fields map[string]*ast.Field, fs *token.FileSet) map[string]string {
	root := os.Getenv("ROOT")
	require.NotEmpty(t, root)

	outPaths := make(map[string]string)

	repositoryPath := prepareGitHubTreePath(t, root)

	for objectName, sections := range objects {
		t.Run(objectName, func(t *testing.T) {
			renderSections := map[string][]byte{}
			for section, fieldInstance := range sections {
				t.Run(section, func(t *testing.T) {

					sectionParsed := iterateOverObject(t, fields, strings.ToLower(section), reflect.TypeOf(fieldInstance), "")

					defs := make(DocDefinitions, 0, len(sectionParsed))
					for k, f := range sectionParsed {
						defs = append(defs, parseDocDefinition(t, root, k.path, k.typ, f, fs))
					}
					defs.Sort()

					renderSections[section] = defs.RenderMarkdown(t, repositoryPath)
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
