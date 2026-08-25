//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package platform

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	goStrings "strings"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

// newTestChart builds a gzipped tar chart archive from the given entries.
func newTestChart(t *testing.T, files map[string]string, order []string) []byte {
	t.Helper()

	var buf bytes.Buffer

	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	for _, name := range order {
		body := files[name]
		require.NoError(t, tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(body)),
		}))
		_, err := tw.Write([]byte(body))
		require.NoError(t, err)
	}

	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())

	return buf.Bytes()
}

// Test_extractChartFile_IgnoresSubcharts ensures only the chart's own top-level file is
// read. The nested subchart entries are written FIRST so that a matcher which does not
// check the path depth would return the wrong (subchart) content.
func Test_extractChartFile_IgnoresSubcharts(t *testing.T) {
	files := map[string]string{
		"mychart/charts/sub/values.yaml":        "owner: subchart\n",
		"mychart/charts/sub/values.schema.json": `{"title":"subchart"}`,
		"mychart/values.yaml":                   "owner: parent\n",
		"mychart/values.schema.json":            `{"title":"parent"}`,
	}
	order := []string{
		"mychart/charts/sub/values.yaml",
		"mychart/charts/sub/values.schema.json",
		"mychart/values.yaml",
		"mychart/values.schema.json",
	}

	chart := newTestChart(t, files, order)

	t.Run("values.yaml", func(t *testing.T) {
		values, err := extractChartValues(chart)
		require.NoError(t, err)
		require.Equal(t, "parent", values["owner"], "must read the chart's own values.yaml, not a subchart's")
	})

	t.Run("values.schema.json", func(t *testing.T) {
		schema, err := extractChartSchema(chart)
		require.NoError(t, err)
		require.NotNil(t, schema)
		require.Equal(t, "parent", schema["title"], "must read the chart's own schema, not a subchart's")
	})
}

// Test_extractChart_Missing covers charts that ship neither file.
func Test_extractChart_Missing(t *testing.T) {
	chart := newTestChart(t, map[string]string{"mychart/Chart.yaml": "name: mychart\n"}, []string{"mychart/Chart.yaml"})

	values, err := extractChartValues(chart)
	require.NoError(t, err)
	require.Empty(t, values)

	schema, err := extractChartSchema(chart)
	require.NoError(t, err)
	require.Nil(t, schema, "charts without a schema must fall back to nil")
}

// Test_extractChartSchema_Malformed ensures an unparsable schema is surfaced as an error
// rather than silently degrading to no validation. packageChartChart propagates it and
// fails the packaging.
func Test_extractChartSchema_Malformed(t *testing.T) {
	chart := newTestChart(t,
		map[string]string{"mychart/values.schema.json": "{not json"},
		[]string{"mychart/values.schema.json"},
	)

	schema, err := extractChartSchema(chart)
	require.Error(t, err, "a chart shipping an unparsable schema must not be accepted")
	require.Nil(t, schema, "must not return a partially decoded schema")
}

// Test_sanitizeOverrideSchema verifies the override-schema relaxation.
func Test_sanitizeOverrideSchema(t *testing.T) {
	in := map[string]interface{}{
		"$schema":              "https://json-schema.org/draft/2020-12/schema",
		"$id":                  "https://example.com/values.schema.json",
		"type":                 "object",
		"required":             []interface{}{"image"},
		"additionalProperties": false,
		"properties": map[string]interface{}{
			"image": map[string]interface{}{
				"type":     "object",
				"required": []interface{}{"tag"},
				"properties": map[string]interface{}{
					"tag": map[string]interface{}{"type": "string"},
				},
			},
			// A chart value that is literally named "required" must survive.
			"required": map[string]interface{}{"type": "boolean"},
		},
	}

	out := sanitizeOverrideSchema(in, "#/properties/charts/properties/demo")

	require.NotContains(t, out, "$schema", "dialect must not be re-declared in a subschema")
	require.NotContains(t, out, "$id", "base URI must not shift")
	require.NotContains(t, out, "required", "overrides are partial documents")
	require.Equal(t, "object", out["type"])
	require.Equal(t, false, out["additionalProperties"], "typo detection must be preserved")

	props := out["properties"].(map[string]interface{})

	// Property literally named "required" is data, not a keyword.
	require.Contains(t, props, "required")
	require.Equal(t, "boolean", props["required"].(map[string]interface{})["type"])

	// Nested `required` keyword is stripped, nested properties preserved.
	image := props["image"].(map[string]interface{})
	require.NotContains(t, image, "required", "nested required must be stripped for partial overrides")
	require.Contains(t, image, "properties")

	// Input must not be mutated.
	require.Contains(t, in, "$schema")
	require.Contains(t, in, "required")
}

// Test_sanitizeOverrideSchema_RewritesLocalRefs ensures a chart schema that uses internal
// `$ref`s stays resolvable once inlined. Without rewriting, `#/definitions/...` would point
// at the generated release schema - which has no `definitions` - and helm would reject the
// release chart even against its own packaged defaults.
func Test_sanitizeOverrideSchema_RewritesLocalRefs(t *testing.T) {
	in := map[string]interface{}{
		"type": "object",
		"definitions": map[string]interface{}{
			"image": map[string]interface{}{"type": "object"},
		},
		"properties": map[string]interface{}{
			"images": map[string]interface{}{
				"additionalProperties": map[string]interface{}{"$ref": "#/definitions/image"},
			},
			"list": map[string]interface{}{
				"items": map[string]interface{}{"$ref": "#/definitions/image"},
			},
			"self":     map[string]interface{}{"$ref": "#"},
			"external": map[string]interface{}{"$ref": "https://example.com/other.json#/definitions/x"},
		},
	}

	base := "#/properties/charts/properties/" + escapeJSONPointer("my-chart")
	out := sanitizeOverrideSchema(in, base)

	props := out["properties"].(map[string]interface{})

	require.Equal(t, base+"/definitions/image",
		props["images"].(map[string]interface{})["additionalProperties"].(map[string]interface{})["$ref"],
		"refs under additionalProperties must be rebased")
	require.Equal(t, base+"/definitions/image",
		props["list"].(map[string]interface{})["items"].(map[string]interface{})["$ref"],
		"refs under items must be rebased")
	require.Equal(t, base, props["self"].(map[string]interface{})["$ref"], "a bare # must become the base")
	require.Equal(t, "https://example.com/other.json#/definitions/x",
		props["external"].(map[string]interface{})["$ref"], "non-local refs must be left alone")

	// The definitions the refs point at must survive the inlining.
	require.Contains(t, out, "definitions")
	require.Contains(t, out["definitions"].(map[string]interface{}), "image")
}

// Test_escapeJSONPointer covers RFC 6901 escaping of chart names used in ref bases.
func Test_escapeJSONPointer(t *testing.T) {
	require.Equal(t, "plain-chart", escapeJSONPointer("plain-chart"))
	require.Equal(t, "a~1b", escapeJSONPointer("a/b"))
	require.Equal(t, "a~0b", escapeJSONPointer("a~b"))
	require.Equal(t, "a~01b", escapeJSONPointer("a~1b"), "~ must be escaped before /")
}

// Test_generateValuesSchema_UsesChartSchema ensures a chart's own schema is inlined when
// present, and that charts without one stay permissive.
func Test_generateValuesSchema_UsesChartSchema(t *testing.T) {
	input := packageChartRenderInput{
		Name:    "arango-platform-release",
		Version: "1.0.0",
		Charts: map[string]packageChartRenderInputChart{
			"with-schema": {
				Name:    "with-schema",
				Version: "1.2.3",
				Schema: map[string]interface{}{
					"$schema":              "https://json-schema.org/draft/2020-12/schema",
					"type":                 "object",
					"required":             []interface{}{"image"},
					"additionalProperties": false,
					"properties": map[string]interface{}{
						"image": map[string]interface{}{"type": "string"},
					},
				},
			},
			"no-schema": {Name: "no-schema", Version: "4.5.6"},
		},
	}

	var out map[string]interface{}
	require.NoError(t, json.Unmarshal(generateValuesSchema(input), &out))

	charts := out["properties"].(map[string]interface{})["charts"].(map[string]interface{})["properties"].(map[string]interface{})

	withSchema := charts["with-schema"].(map[string]interface{})
	require.Equal(t, false, withSchema["additionalProperties"], "chart schema must be enforced")
	require.NotContains(t, withSchema, "required", "override block is partial")
	require.NotContains(t, withSchema, "$schema")
	require.Contains(t, withSchema, "properties")
	require.Contains(t, withSchema["description"], "with-schema")

	noSchema := charts["no-schema"].(map[string]interface{})
	require.Equal(t, true, noSchema["additionalProperties"], "charts without a schema stay permissive")
}

// Test_documentedValues covers the per-chart value documentation: flattening to full
// override paths, descriptions pulled from the chart schema, and stable order.
func Test_documentedValues(t *testing.T) {
	chartValues := map[string]interface{}{
		"replicas": 1,
		"image":    "registry.example.com/demo:1.0.0",
		"empty":    "",
		"resources": map[string]interface{}{
			"requests": map[string]interface{}{"cpu": "100m", "memory": "128Mi"},
		},
	}

	chartSchema := map[string]interface{}{
		"properties": map[string]interface{}{
			"replicas": map[string]interface{}{"description": "Number of replicas"},
			"image":    map[string]interface{}{"description": "Container image | with a pipe"},
			// "empty" intentionally has no description
			"resources": map[string]interface{}{
				"description": "Compute resources",
				"properties": map[string]interface{}{
					"requests": map[string]interface{}{
						// no description on this level - it must not be listed
						"properties": map[string]interface{}{
							"cpu": map[string]interface{}{"description": "CPU request"},
							// "memory" intentionally has no description
						},
					},
				},
			},
		},
	}

	values := documentedValues(chartValues, chartSchema)
	require.NotEmpty(t, values)

	byKey := map[string]packageChartRenderInputValue{}
	keys := make([]string, 0, len(values))
	for _, v := range values {
		byKey[v.Key] = v
		keys = append(keys, v.Key)
	}

	// Nested values are expanded to full dotted paths, never collapsed into a JSON blob.
	require.Equal(t, []string{
		"empty",
		"image",
		"replicas",
		"resources",
		"resources.requests.cpu",
		"resources.requests.memory",
	}, keys, "values must be flattened and sorted for stable output")

	require.Equal(t, "1", byKey["replicas"].Default)
	require.Equal(t, "Number of replicas", byKey["replicas"].Description)

	// Pipes are escaped so they cannot break the Markdown table.
	require.Equal(t, `Container image \| with a pipe`, byKey["image"].Description)

	// Empty string is rendered visibly rather than as a blank cell.
	require.Equal(t, `""`, byKey["empty"].Default)

	// Leaves carry their own value and description from the nested schema.
	require.Equal(t, "100m", byKey["resources.requests.cpu"].Default)
	require.Equal(t, "CPU request", byKey["resources.requests.cpu"].Description)
	require.Equal(t, "128Mi", byKey["resources.requests.memory"].Default)

	// A documented intermediate object is listed so its description survives, but it has
	// no default of its own - its children carry the values.
	require.Empty(t, byKey["resources"].Default)
	require.Equal(t, "Compute resources", byKey["resources"].Description)

	// No value may render as JSON.
	for _, v := range values {
		require.NotContains(t, v.Default, "{", "nested values must be flattened, not JSON: %s", v.Key)
	}

	t.Run("chart with no values yields nothing", func(t *testing.T) {
		require.Nil(t, documentedValues(nil, nil))
	})

	t.Run("chart without schema still documents defaults", func(t *testing.T) {
		v := documentedValues(map[string]interface{}{"a": 1}, nil)
		require.Len(t, v, 1)
		require.Equal(t, "1", v[0].Default)
		require.Empty(t, v[0].Description)
	})
}

// Test_packageChartTemplateReadme renders the generated chart README.
func Test_packageChartTemplateReadme(t *testing.T) {
	t.Run("with charts and services", func(t *testing.T) {
		input := packageChartRenderInput{
			Name:    "arango-platform-release",
			Version: "1.0.0",
			Charts: map[string]packageChartRenderInputChart{
				"gral": {
					Name: "gral", Version: "1.2.3",
					Schema:           map[string]interface{}{"type": "object"},
					DocumentedValues: []packageChartRenderInputValue{{Key: "replicas", Default: "2", Description: "Replica count"}},
				},
				"no-schem": {Name: "no-schem", Version: "4.5.6"},
			},
			Services: map[string]packageChartRenderInputService{
				"gral": {Name: "gral", ChartRef: "arangodb-gral"},
			},
		}

		out, err := packageChartTemplateReadme.RenderBytes(input)
		require.NoError(t, err)

		readme := string(out)
		t.Logf("rendered README.md:\n%s", readme)

		require.Contains(t, readme, "# arango-platform-release")
		require.Contains(t, readme, "`1.0.0`")
		// Charts table
		require.Contains(t, readme, "| `gral` | `1.2.3` |")
		require.Contains(t, readme, "| `no-schem` | `4.5.6` |")
		// Schema presence is reflected per chart
		require.Contains(t, readme, "validated against the chart's own `values.schema.json`")
		require.Contains(t, readme, "not validated - chart ships no schema")
		// Services table
		require.Contains(t, readme, "| `gral` | `arangodb-gral` |")
		// Values are documented per chart, using the full override path
		require.Contains(t, readme, "| `charts.gral.replicas` | `2` | Replica count |")
		require.Contains(t, readme, "This chart exposes no configurable values.")
		require.NotContains(t, readme, "### Service values")
		// No unrendered template directives leak through
		require.NotContains(t, readme, "{{")
		require.NotContains(t, readme, "<no value>")

		// A single, complete values.yaml example - deployment plus every chart and service.
		// A services-only or charts-only snippet would not be a valid values file, since
		// `deployment` is required.
		example := readmeExample(t, readme)
		require.Contains(t, example, "deployment: my-deployment")
		require.Contains(t, example, "charts:")
		require.Contains(t, example, "gral: {}")
		require.Contains(t, example, "no-schem: {}")
		require.Contains(t, example, "services:")
		require.Contains(t, example, "values: {}")
		require.Equal(t, 1, goStrings.Count(readme, "A complete `values.yaml`"), "exactly one example block")
	})

	t.Run("empty release", func(t *testing.T) {
		out, err := packageChartTemplateReadme.RenderBytes(packageChartRenderInput{
			Name:    "empty-release",
			Version: "0.0.1",
		})
		require.NoError(t, err)

		readme := string(out)
		require.Contains(t, readme, "This release bundles no charts.")
		require.Contains(t, readme, "This release bundles no services.")
		require.NotContains(t, readme, "{{")

		// With nothing bundled the example must not emit empty `charts:`/`services:` keys.
		example := readmeExample(t, readme)
		require.Contains(t, example, "deployment: my-deployment")
		require.NotContains(t, example, "charts:")
		require.NotContains(t, example, "services:")
	})
}

// readmeExample extracts the fenced YAML of the README's complete values.yaml example.
func readmeExample(t *testing.T, readme string) string {
	t.Helper()

	_, after, found := goStrings.Cut(readme, "A complete `values.yaml`")
	require.True(t, found, "README must contain the complete values.yaml example")

	_, after, found = goStrings.Cut(after, "```yaml\n")
	require.True(t, found, "example must be a fenced yaml block")

	example, _, found := goStrings.Cut(after, "```")
	require.True(t, found, "example block must be closed")

	return example
}

// Test_packageChartTemplateValues_EmptySections ensures a release with no charts or no
// services still emits valid values.yaml. A bare `services:` key parses as null, which the
// generated schema rejects as "Expected: object, given: null" - making the release chart
// uninstallable against its own defaults.
func Test_packageChartTemplateValues_EmptySections(t *testing.T) {
	for name, input := range map[string]packageChartRenderInput{
		"no services": {
			Name: "r", Version: "1",
			Charts: map[string]packageChartRenderInputChart{"c": {Name: "c", Version: "1"}},
		},
		"no charts": {Name: "r", Version: "1", Services: map[string]packageChartRenderInputService{"s": {Name: "s"}}},
		"empty":     {Name: "r", Version: "1"},
	} {
		t.Run(name, func(t *testing.T) {
			out, err := packageChartTemplateValues.RenderBytes(input)
			require.NoError(t, err)

			var doc map[string]interface{}
			require.NoError(t, yaml.Unmarshal(out, &doc), "values.yaml must parse")

			for _, key := range []string{"charts", "services"} {
				require.Contains(t, doc, key)
				require.NotNil(t, doc[key], "%s must be an object, never null: %s", key, string(out))
			}
		})
	}
}

// Test_imagesFromValues ensures container images are derived ONLY from the root `images` map,
// following the ArangoDB chart convention `images.<name>`. Image specs anywhere else in values are
// inherited upstream defaults rather than images the chart declares, and must stay out of the
// air-gapped list; the chart's own images.yaml is not scanned.
func Test_imagesFromValues(t *testing.T) {
	t.Run("derives the root images map", func(t *testing.T) {
		got := imagesFromValues(map[string]interface{}{
			// ArangoDB convention: images.<name>.{image,registry,tag}
			"images": map[string]interface{}{
				"application": map[string]interface{}{
					"image":    "gral/engine",
					"registry": "registry.license.arango.ai",
					"tag":      "v1.1.12",
				},
				// "repository" is accepted as an alias for "image".
				"reloader": map[string]interface{}{
					"repository": "platform-monitoring/prometheus-config-reloader",
					"registry":   "registry.license.arango.ai",
					"tag":        "v0.0.6",
				},
			},
			// Not an image spec - must be ignored.
			"imagePullPolicy": "IfNotPresent",
			"replicas":        3,
		})

		// Ordered by key so the generated list is reproducible.
		require.Equal(t, []packageChartRenderInputImage{
			{OverridePath: "images.application", Image: "registry.license.arango.ai/gral/engine:v1.1.12"},
			{OverridePath: "images.reloader", Image: "registry.license.arango.ai/platform-monitoring/prometheus-config-reloader:v0.0.6"},
		}, got)
	})

	t.Run("ignores image specs outside the root images map", func(t *testing.T) {
		// The shape of an upstream chart fork: the chart declares one image it owns, and carries
		// upstream's own specs for optional components it merely inherits.
		got := imagesFromValues(map[string]interface{}{
			"images": map[string]interface{}{
				"grafana": map[string]interface{}{
					"image":    "platform-monitoring/grafana",
					"registry": "registry.license.arango.ai",
					"tag":      "v0.0.6",
				},
			},
			"initChownData": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   "docker.io",
					"repository": "library/busybox",
					"tag":        "1.31.1",
				},
			},
			"sidecar": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   "quay.io",
					"repository": "kiwigrid/k8s-sidecar",
					"tag":        "1.30.10",
				},
			},
			"imageRenderer": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   "docker.io",
					"repository": "grafana/grafana-image-renderer",
					"tag":        "latest",
				},
			},
			"testFramework": map[string]interface{}{
				"image": map[string]interface{}{
					"registry":   "docker.io",
					"repository": "bats/bats",
					"tag":        "v1.4.1",
				},
			},
			"downloadDashboardsImage": map[string]interface{}{
				"registry":   "docker.io",
				"repository": "curlimages/curl",
				"tag":        "8.9.1",
			},
		})

		require.Equal(t, []packageChartRenderInputImage{
			{OverridePath: "images.grafana", Image: "registry.license.arango.ai/platform-monitoring/grafana:v0.0.6"},
		}, got)
	})

	t.Run("excludes test images", func(t *testing.T) {
		got := imagesFromValues(map[string]interface{}{
			"images": map[string]interface{}{
				"application": map[string]interface{}{
					"image":    "gral/engine",
					"registry": "registry.license.arango.ai",
					"tag":      "v1.1.12",
				},
				// Excluded: sits under the "test" key.
				"test": map[string]interface{}{
					"image":    "gral/engine-test",
					"registry": "registry.license.arango.ai",
					"tag":      "v1.1.12",
				},
				// Excluded: repository ends with -test even though the key is not "test".
				"extra": map[string]interface{}{
					"image":    "foo/bar-test",
					"registry": "docker.io",
					"tag":      "1.0",
				},
			},
		})

		require.Equal(t, []packageChartRenderInputImage{
			{OverridePath: "images.application", Image: "registry.license.arango.ai/gral/engine:v1.1.12"},
		}, got)
	})

	t.Run("composes a repo that already carries its registry, and a numeric tag", func(t *testing.T) {
		got := imagesFromValues(map[string]interface{}{
			"images": map[string]interface{}{
				"other": map[string]interface{}{
					"image":    "docker.io/library/busybox",
					"registry": "docker.io",
					"tag":      1.31,
				},
			},
		})

		require.Equal(t, []packageChartRenderInputImage{
			{OverridePath: "images.other", Image: "docker.io/library/busybox:1.31"},
		}, got)
	})

	t.Run("deduplicates by image reference", func(t *testing.T) {
		got := imagesFromValues(map[string]interface{}{
			"images": map[string]interface{}{
				"b": map[string]interface{}{"image": "arangodb/enterprise", "tag": "3.12.5"},
				"a": map[string]interface{}{"image": "arangodb/enterprise", "tag": "3.12.5"},
			},
		})

		// The first key in sorted order wins, so the kept entry is deterministic.
		require.Equal(t, []packageChartRenderInputImage{
			{OverridePath: "images.a", Image: "arangodb/enterprise:3.12.5"},
		}, got)
	})

	t.Run("no usable root images map yields nothing", func(t *testing.T) {
		require.Nil(t, imagesFromValues(map[string]interface{}{"replicas": 3}))
		require.Nil(t, imagesFromValues(map[string]interface{}{"images": "not-a-map"}))
		// A non-map entry under images is skipped rather than failing the packaging.
		require.Empty(t, imagesFromValues(map[string]interface{}{
			"images": map[string]interface{}{"application": "not-a-map"},
		}))
	})
}

// Test_aggregateImages ensures images from every chart are merged, de-duplicated by image
// reference, and ordered deterministically regardless of chart map iteration order.
func Test_aggregateImages(t *testing.T) {
	charts := map[string]packageChartRenderInputChart{
		"a": {Name: "a", Images: []packageChartRenderInputImage{
			{OverridePath: "images.operator", Image: "arangodb/kube-arangodb:1.4.4"},
			{OverridePath: "images.db", Image: "arangodb/arangodb-enterprise:3.12.5"},
		}},
		"b": {Name: "b", Images: []packageChartRenderInputImage{
			{OverridePath: "images.dup", Image: "arangodb/kube-arangodb:1.4.4"}, // duplicate image, dropped
			{OverridePath: "images.webui", Image: "arangodb/webui:2.0.0"},
			{OverridePath: "images.empty", Image: ""}, // empty image, skipped
		}},
	}

	// Sorted by image; paths chart-prefixed; the duplicate keeps chart "a" (name order).
	require.Equal(t, []packageChartRenderInputImage{
		{OverridePath: "charts.a.images.db", Image: "arangodb/arangodb-enterprise:3.12.5"},
		{OverridePath: "charts.a.images.operator", Image: "arangodb/kube-arangodb:1.4.4"},
		{OverridePath: "charts.b.images.webui", Image: "arangodb/webui:2.0.0"},
	}, aggregateImages(charts))
}

// Test_renderImagesFile ensures the aggregated images.yaml carries the informational header and the
// image entries.
func Test_renderImagesFile(t *testing.T) {
	out := string(renderImagesFile(packageChartRenderInput{
		Name: "arango-platform-release", Version: "1.0.0",
		Images: []packageChartRenderInputImage{{OverridePath: "charts.op.images.application", Image: "arangodb/kube-arangodb:1.4.4"}},
	}))

	require.Contains(t, out, "# Container images bundled by arango-platform-release 1.0.0.")
	require.Contains(t, out, "Helm does not consume this file")
	require.Contains(t, out, "overridePath: charts.op.images.application")
	require.Contains(t, out, "image: arangodb/kube-arangodb:1.4.4")

	// It must be a valid images.yaml document.
	var doc struct {
		Images []packageChartRenderInputImage `json:"images"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(out), &doc))
	require.Equal(t, []packageChartRenderInputImage{{OverridePath: "charts.op.images.application", Image: "arangodb/kube-arangodb:1.4.4"}}, doc.Images)

	// With no images the document renders an empty list, never null.
	empty := string(renderImagesFile(packageChartRenderInput{Name: "r", Version: "1"}))
	require.Contains(t, empty, "images: []")
	require.NotContains(t, empty, "images: null")
}

// Test_packageChartTemplateReadme_Images ensures the README renders the aggregated container-image
// section, and degrades cleanly when a release declares none.
func Test_packageChartTemplateReadme_Images(t *testing.T) {
	t.Run("with images", func(t *testing.T) {
		out, err := packageChartTemplateReadme.RenderBytes(packageChartRenderInput{
			Name: "arango-platform-release", Version: "1.0.0",
			Images: []packageChartRenderInputImage{
				{OverridePath: "charts.op.images.application", Image: "arangodb/kube-arangodb:1.4.4"},
				{OverridePath: "charts.db.images.application", Image: "arangodb/arangodb-enterprise:3.12.5"},
			},
		})
		require.NoError(t, err)

		readme := string(out)
		require.Contains(t, readme, "## Container Images")
		require.Contains(t, readme, "| `arangodb/kube-arangodb:1.4.4` | `charts.op.images.application` |")
		require.Contains(t, readme, "| `arangodb/arangodb-enterprise:3.12.5` | `charts.db.images.application` |")
		require.Contains(t, readme, "images.yaml")
		require.NotContains(t, readme, "{{")
		require.NotContains(t, readme, "<no value>")
	})

	t.Run("no images", func(t *testing.T) {
		out, err := packageChartTemplateReadme.RenderBytes(packageChartRenderInput{Name: "r", Version: "1"})
		require.NoError(t, err)
		require.Contains(t, string(out), "This release declares no container images.")
		require.NotContains(t, string(out), "{{")
	})
}

// Test_mergeChartValues_DeepMerge ensures chart overrides are deep-merged on top of the chart
// defaults (nested maps preserved, overrides win), empty overrides are a no-op, and a malformed
// override document is surfaced rather than silently dropped.
func Test_mergeChartValues_DeepMerge(t *testing.T) {
	defaults := map[string]interface{}{
		"image": map[string]interface{}{
			"repository": "default-repo",
			"tag":        "1.0",
		},
		"replicas": float64(1),
	}

	t.Run("deep merge keeps sibling keys", func(t *testing.T) {
		out, err := mergeChartValues(defaults, helm.Values(`{"image":{"repository":"custom-repo"},"replicas":3}`))
		require.NoError(t, err)

		img := out["image"].(map[string]interface{})
		require.EqualValues(t, "custom-repo", img["repository"], "override applied")
		require.EqualValues(t, "1.0", img["tag"], "sibling key preserved by deep merge")
		require.EqualValues(t, float64(3), out["replicas"])
	})

	t.Run("empty overrides is a no-op", func(t *testing.T) {
		out, err := mergeChartValues(defaults, nil)
		require.NoError(t, err)
		require.Equal(t, defaults, out)
	})

	t.Run("malformed overrides surface an error", func(t *testing.T) {
		_, err := mergeChartValues(defaults, helm.Values(`{not valid json`))
		require.Error(t, err)
	})
}

// Test_packageChartTemplateValues_Overrides ensures release (service) overrides and chart overrides
// are rendered into the generated values.yaml, rather than being dropped as an empty object.
func Test_packageChartTemplateValues_Overrides(t *testing.T) {
	input := packageChartRenderInput{
		Name: "r", Version: "1",
		Charts: map[string]packageChartRenderInputChart{
			"c": {Name: "c", Version: "1", Values: map[string]interface{}{"image": map[string]interface{}{"repository": "custom-repo"}}},
		},
		Services: map[string]packageChartRenderInputService{
			"withValues": {Name: "withValues", ChartRef: "c", Values: map[string]interface{}{"foo": "bar", "nested": map[string]interface{}{"x": float64(1)}}},
			"empty":      {Name: "empty", ChartRef: "c"},
		},
	}

	out, err := packageChartTemplateValues.RenderBytes(input)
	require.NoError(t, err)

	var doc map[string]interface{}
	require.NoError(t, yaml.Unmarshal(out, &doc), "values.yaml must parse: %s", string(out))

	charts := doc["charts"].(map[string]interface{})
	cVals := charts["c"].(map[string]interface{})
	require.EqualValues(t, "custom-repo", cVals["image"].(map[string]interface{})["repository"], "chart override must be rendered")

	services := doc["services"].(map[string]interface{})

	withValues := services["withValues"].(map[string]interface{})
	vals := withValues["values"].(map[string]interface{})
	require.EqualValues(t, "bar", vals["foo"], "service override must be rendered, not dropped")
	require.EqualValues(t, float64(1), vals["nested"].(map[string]interface{})["x"])

	empty := services["empty"].(map[string]interface{})
	require.NotNil(t, empty["values"], "service without overrides keeps an empty-object values")
	require.Empty(t, empty["values"], "service without overrides renders {}")
}
