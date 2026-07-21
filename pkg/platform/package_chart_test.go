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

	out := sanitizeOverrideSchema(in)

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
