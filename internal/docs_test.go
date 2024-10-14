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
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/require"

	analyticsApi "github.com/arangodb/kube-arangodb/pkg/apis/analytics/v1alpha1"
	appsApi "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	backupApi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	mlApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1alpha1"
	mlApi "github.com/arangodb/kube-arangodb/pkg/apis/ml/v1beta1"
	networkingApi "github.com/arangodb/kube-arangodb/pkg/apis/networking/v1alpha1"
	replicationApi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	schedulerApiv1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	storageApi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

const (
	// title of docs/api/README.md page
	apiIndexPageTitle = "CRD reference"
)

func (d DocDefinitions) RenderMarkdown(t *testing.T, repositoryPath string) []byte {
	out := bytes.NewBuffer(nil)

	for i, el := range d {
		if i != 0 {
			write(t, out, "***\n\n")
		}

		write(t, out, "### %s\n\n", el.Path)
		write(t, out, "Type: `%s` <sup>[\\[ref\\]](%s/%s#L%d)</sup>\n\n", el.Type, repositoryPath, el.File, el.Line)

		if d := el.Deprecated; len(d) > 0 {
			write(t, out, "> [!WARNING]\n")
			write(t, out, "> ***DEPRECATED***\n")
			write(t, out, "> \n")
			for _, line := range d {
				write(t, out, "> **%s**\n", line)
			}
			write(t, out, "\n")
		}

		if d := el.Important; d != nil {
			write(t, out, "> [!IMPORTANT]\n")
			write(t, out, "> **%s**\n\n", *d)
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

				snip := fmt.Sprintf("`\"%s\"`", z[0])
				if id == 0 {
					snip = fmt.Sprintf("%s (default)", snip)
				}

				if len(z) == 1 {
					write(t, out, "* %s\n", snip)
				} else if len(z) == 2 {
					write(t, out, "* %s - %s\n", snip, z[1])
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

	fset := token.NewFileSet()

	type inputPackageTypes map[string]map[string]any

	type inputPackage struct {
		Types inputPackageTypes

		Shared []string
	}

	type inputPackages map[string]map[string]inputPackage

	// package path -> result doc file name -> name of the top-level field to be described -> field instance for reflection
	input := inputPackages{
		"deployment": map[string]inputPackage{
			"v1": {
				Types: inputPackageTypes{
					"ArangoDeployment.V1": {
						"Spec": deploymentApi.ArangoDeployment{}.Spec,
					},
					"ArangoMember.V1": {
						"Spec": deploymentApi.ArangoMember{}.Spec,
					},
				},
				Shared: []string{
					"shared/v1",
					"scheduler/v1beta1",
					"scheduler/v1beta1/container",
					"scheduler/v1beta1/container/resources",
					"scheduler/v1beta1/integration",
					"scheduler/v1beta1/pod",
					"scheduler/v1beta1/pod/resources",
					"scheduler/v1beta1/policy",
				},
			},
		},
		"apps": map[string]inputPackage{
			"v1": {
				Types: inputPackageTypes{
					"ArangoJob.V1": {
						"Spec": appsApi.ArangoJob{}.Spec,
					},
				},
			},
		},
		"backup": map[string]inputPackage{
			"v1": {
				Types: inputPackageTypes{
					"ArangoBackup.V1": {
						"Spec":   backupApi.ArangoBackup{}.Spec,
						"Status": backupApi.ArangoBackup{}.Status,
					},
					"ArangoBackupPolicy.V1": {
						"Spec":   backupApi.ArangoBackupPolicy{}.Spec,
						"Status": backupApi.ArangoBackupPolicy{}.Status,
					},
				},
			},
		},
		"ml": map[string]inputPackage{
			"v1alpha1": {
				Types: inputPackageTypes{
					"ArangoMLExtension.V1Alpha1": {
						"Spec":   mlApiv1alpha1.ArangoMLExtension{}.Spec,
						"Status": mlApiv1alpha1.ArangoMLExtension{}.Status,
					},
					"ArangoMLStorage.V1Alpha1": {
						"Spec":   mlApiv1alpha1.ArangoMLStorage{}.Spec,
						"Status": mlApiv1alpha1.ArangoMLStorage{}.Status,
					},
					"ArangoMLCronJob.V1Alpha1": {
						"Spec":   mlApiv1alpha1.ArangoMLCronJob{}.Spec,
						"Status": mlApiv1alpha1.ArangoMLCronJob{}.Status,
					},
					"ArangoMLBatchJob.V1Alpha1": {
						"Spec":   mlApiv1alpha1.ArangoMLBatchJob{}.Spec,
						"Status": mlApiv1alpha1.ArangoMLBatchJob{}.Status,
					},
				},
				Shared: []string{
					"shared/v1",
					"scheduler/v1alpha1",
					"scheduler/v1alpha1/container",
					"scheduler/v1alpha1/container/resources",
					"scheduler/v1alpha1/pod",
					"scheduler/v1alpha1/pod/resources",
				},
			},
			"v1beta1": {
				Types: inputPackageTypes{
					"ArangoMLExtension.V1Beta1": {
						"Spec":   mlApi.ArangoMLExtension{}.Spec,
						"Status": mlApi.ArangoMLExtension{}.Status,
					},
					"ArangoMLStorage.V1Beta1": {
						"Spec":   mlApi.ArangoMLStorage{}.Spec,
						"Status": mlApi.ArangoMLStorage{}.Status,
					},
				},
				Shared: []string{
					"shared/v1",
					"scheduler/v1beta1",
					"scheduler/v1beta1/container",
					"scheduler/v1beta1/container/resources",
					"scheduler/v1beta1/integration",
					"scheduler/v1beta1/pod",
					"scheduler/v1beta1/pod/resources",
					"scheduler/v1beta1/policy",
				},
			},
		},
		"networking": map[string]inputPackage{
			"v1alpha1": {
				Types: inputPackageTypes{
					"ArangoRoute.V1Alpha1": {
						"Spec":   networkingApi.ArangoRoute{}.Spec,
						"Status": networkingApi.ArangoRoute{}.Status,
					},
				},
				Shared: []string{
					"shared/v1",
				},
			},
		},
		"analytics": map[string]inputPackage{
			"v1alpha1": {
				Types: inputPackageTypes{
					"GraphAnalyticsEngine.V1Alpha1": {
						"Spec":   analyticsApi.GraphAnalyticsEngine{}.Spec,
						"Status": analyticsApi.GraphAnalyticsEngine{}.Status,
					},
				},
				Shared: []string{
					"shared/v1",
					"scheduler/v1beta1",
					"scheduler/v1beta1/container",
					"scheduler/v1beta1/container/resources",
					"scheduler/v1beta1/integration",
					"scheduler/v1beta1/pod",
					"scheduler/v1beta1/pod/resources",
					"scheduler/v1beta1/policy",
				},
			},
		},
		"replication": map[string]inputPackage{
			"v1": {
				Types: inputPackageTypes{
					"ArangoDeploymentReplication.V1": {
						"Spec": replicationApi.ArangoDeploymentReplication{}.Spec,
					},
				},
			},
		},
		"scheduler": map[string]inputPackage{
			"v1alpha1": {
				Types: inputPackageTypes{
					"ArangoProfile.V1Alpha1": {
						"Spec": schedulerApiv1alpha1.ArangoProfile{}.Spec,
					},
				},
				Shared: []string{
					"shared/v1",
					"scheduler/v1alpha1/container",
					"scheduler/v1alpha1/container/resources",
					"scheduler/v1alpha1/pod",
					"scheduler/v1alpha1/pod/resources",
				},
			},
			"v1beta1": {
				Types: inputPackageTypes{
					"ArangoProfile.V1Beta1": {
						"Spec": schedulerApi.ArangoProfile{}.Spec,
					},
				},
				Shared: []string{
					"shared/v1",
					"scheduler/v1beta1/container",
					"scheduler/v1beta1/container/resources",
					"scheduler/v1beta1/integration",
					"scheduler/v1beta1/pod",
					"scheduler/v1beta1/pod/resources",
					"scheduler/v1beta1/policy",
				},
			},
		},
		"storage": map[string]inputPackage{
			"v1alpha": {
				Types: inputPackageTypes{
					"ArangoLocalStorage.V1Alpha": {
						"Spec": storageApi.ArangoLocalStorage{}.Spec,
					},
				},
			},
		},
	}

	for name, versions := range input {
		for version, docs := range versions {
			fields := parseSourceFiles(t, root, fset, path.Join(root, "pkg/apis", name, version))

			for _, p := range docs.Shared {
				sharedFields := parseSourceFiles(t, root, fset, path.Join(root, "pkg/apis", p))

				for n, f := range sharedFields {
					require.NotContains(t, fields, n)
					fields[n] = f
				}
			}

			generateDocs(t, docs.Types, fields, fset)
		}
	}
}

func writeFrontMatter(t *testing.T, out io.Writer, keyVals map[string]string) {
	fm := ""
	util.IterateSorted(keyVals, func(key, val string) {
		fm += fmt.Sprintf("%s: %s\n", key, val)
	})

	if fm != "" {
		fm = "---\n" + fm + "---\n\n"
	}

	write(t, out, fm)
}

func extractVersion(t *testing.T, root string) *semver.Version {
	if v := extractVersionFile(t, root); v != nil {
		return v
	} else {
		t.Logf("Unable to get Version from file, fallback to git")
	}

	if v := extractVersionGit(t, root); v != nil {
		return v
	} else {
		t.Logf("Unable to get Version from Git")
	}

	require.FailNow(t, "Unable to get version")

	return nil
}

func extractVersionFile(t *testing.T, root string) *semver.Version {
	data, err := os.ReadFile(path.Join(root, "VERSION"))
	require.NoError(t, err)

	v := strings.TrimSpace(string(data))
	sm, err := semver.NewVersion(v)
	require.NoError(t, err)

	if v := sm.PreRelease.Slice(); len(v) > 0 && v[0] != "" {
		return nil
	}

	return sm
}

func extractVersionGit(t *testing.T, root string) *semver.Version {
	cmd := exec.Command("git", "tag", "--list")
	cmd.Dir = root

	out, err := cmd.StdoutPipe()
	require.NoError(t, err)

	require.NoError(t, cmd.Start())

	versions := semver.Versions{}

	scanner := bufio.NewScanner(out)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		v := strings.TrimSpace(scanner.Text())
		sm, err := semver.NewVersion(v)
		if err != nil {
			t.Logf("Unable to parse: %s", v)
			continue
		}

		if v := sm.PreRelease.Slice(); len(v) > 0 && v[0] != "" {
			continue
		}

		versions = append(versions, sm)
	}

	require.NoError(t, cmd.Wait())

	if len(versions) == 0 {
		return nil
	}

	sort.Sort(versions)

	return versions[len(versions)-1]
}

func prepareGitHubTreePath(t *testing.T, root string) string {
	opVersion := extractVersion(t, root)
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

			objName := strings.ReplaceAll(objectName, ".", " ")
			writeFrontMatter(t, out, map[string]string{
				"layout": "page",
				"title":  objName,
				"parent": apiIndexPageTitle,
			})
			write(t, out, "# API Reference for %s\n\n", objName)

			util.IterateSorted(renderSections, func(name string, section []byte) {
				write(t, out, "## %s\n\n", name)

				_, err = out.Write(section)
				require.NoError(t, err)
			})
		})
	}
	return outPaths
}

func write(t *testing.T, out io.Writer, format string, args ...interface{}) {
	_, err := out.Write([]byte(fmt.Sprintf(format, args...)))
	require.NoError(t, err)
}
