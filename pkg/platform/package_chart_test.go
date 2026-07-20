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

// Test_packageChartTemplateReadme renders the generated chart README.
func Test_packageChartTemplateReadme(t *testing.T) {
	t.Run("with charts and services", func(t *testing.T) {
		input := packageChartRenderInput{
			Name:    "arango-platform-release",
			Version: "1.0.0",
			Charts: map[string]packageChartRenderInputChart{
				"gral":     {Name: "gral", Version: "1.2.3", Schema: map[string]interface{}{"type": "object"}},
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
		// No unrendered template directives leak through
		require.NotContains(t, readme, "{{")
		require.NotContains(t, readme, "<no value>")
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
	})
}
