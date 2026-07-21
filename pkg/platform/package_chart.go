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
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"os"
	"path"
	"sort"
	goStrings "strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/platform/pack"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	licenseData "github.com/arangodb/kube-arangodb/pkg/util/license"
)

var (
	//go:embed templates/chart.yaml
	packageChartTemplateChart util.Template[packageChartRenderInput]

	//go:embed templates/check.yaml
	packageChartTemplateCheck util.Template[packageChartRenderInput]

	//go:embed templates/values.yaml
	packageChartTemplateValues util.Template[packageChartRenderInput]

	//go:embed templates/notes.txt
	packageChartTemplateNotes util.Template[packageChartRenderInput]

	//go:embed templates/readme.md
	packageChartTemplateReadme util.Template[packageChartRenderInput]

	//go:embed templates/resource.chart.yaml
	packageChartTemplateResourceChart util.Template[packageChartRenderInputChart]

	//go:embed templates/resource.service.yaml
	packageChartTemplateResourceService util.Template[packageChartRenderInputServiceTemplate]
)

type packageChartRenderInput struct {
	Name    string
	Version string

	Charts   map[string]packageChartRenderInputChart
	Services map[string]packageChartRenderInputService
}
type packageChartRenderInputChart struct {
	Name      string
	Version   string
	ChartData []byte
	Values    map[string]interface{}
	Schema    map[string]interface{}

	// DocumentedValues flattens Values into individual override paths, described using
	// Schema where it provides descriptions.
	DocumentedValues []packageChartRenderInputValue
}

type packageChartRenderInputService struct {
	Name     string
	ChartRef string
}

// packageChartRenderInputValue is a single documented top-level value of a service.
type packageChartRenderInputValue struct {
	Key         string
	Default     string
	Description string
}

type packageChartRenderInputServiceTemplate struct {
	Name     string
	ChartRef string
}

func packageChart() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "chart [flags] output ... packages"
	cmd.Short = "Generates the Helm Chart version of the Platform Installation"

	if err := cli.RegisterFlags(&cmd, flagLicenseManager, flagRegistry, flagLicenseManagerDiscoverCredentials); err != nil {
		return nil, err
	}

	cmd.RunE = getRunner().With(packageChartRun).Run

	return &cmd, nil
}

func packageChartRun(cmd *cobra.Command, args []string) error {
	var hosts map[string]util.ModR[config.Host]

	if newHosts, err := cli.LicenseManagerRegistryHosts(cmd, flagLicenseManager, flagLicenseManager); err != nil {
		logger.Err(err).Warn("Unable to fetch credentials")
	} else {
		hosts = newHosts
	}

	reg, err := flagRegistry.Client(cmd, hosts)
	if err != nil {
		return err
	}

	endpoint, err := flagLicenseManager.Endpoint(cmd)
	if err != nil {
		return err
	}

	if len(args) < 2 {
		return errors.Errorf("Invalid arguments")
	}

	r, err := getHelmPackages(args[1:]...)
	if err != nil {
		return err
	}

	version := "1.0.0"
	if r.Version != nil && *r.Version != "" {
		version = *r.Version
	}

	input := packageChartRenderInput{
		Name:    "arango-platform-release",
		Version: version,
	}

	out, err := os.OpenFile(args[0], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	input.Charts = map[string]packageChartRenderInputChart{}
	input.Services = map[string]packageChartRenderInputService{}

	for k, v := range r.Packages {
		o, err := packageChartChart(cmd.Context(), reg, endpoint, k, v)
		if err != nil {
			return err
		}

		if o != nil {
			input.Charts[k] = *o
		}
	}

	for k, v := range r.Releases {
		o := packageChartRelease(k, v)

		if o != nil {
			input.Services[k] = *o
		}
	}

	schema := generateValuesSchema(input)

	builder := util.NewGZipBuilder(out)

	builder = builder.File(util.GZipBuilderProcessBytes(schema), "%s/values.schema.json", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateChart, input), "%s/Chart.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateValues, input), "%s/values.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateReadme, input), "%s/README.md", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateCheck, input), "%s/templates/check.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateNotes, input), "%s/templates/NOTES.txt", input.Name)

	// Generate per-chart templates and files
	for _, c := range input.Charts {
		builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateResourceChart, c), "%s/templates/charts/%s.yaml", input.Name, c.Name)
		builder = builder.File(util.GZipBuilderProcessBytes(c.ChartData), "%s/files/%s.tgz", input.Name, c.Name)
	}

	// Generate per-service templates
	for _, s := range input.Services {
		builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateResourceService, packageChartRenderInputServiceTemplate(s)), "%s/templates/services/%s.yaml", input.Name, s.Name)
	}

	builder = builder.File(util.GZipBuilderProcessBytes(licenseData.Full()), "%s/LICENSE", input.Name)

	if err := builder.Done(); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}

func packageChartRelease(name string, packageSpec helm.PackageRelease) *packageChartRenderInputService {
	return &packageChartRenderInputService{
		Name:     name,
		ChartRef: packageSpec.Package,
	}
}

func packageChartChart(ctx context.Context, reg *regclient.RegClient, endpoint string, name string, packageSpec helm.PackageSpec) (*packageChartRenderInputChart, error) {
	chart, err := pack.ResolvePackageSpec(ctx, endpoint, name, packageSpec, reg, nil)
	if err != nil {
		return nil, err
	}

	defaults, err := extractChartValues(chart)
	if err != nil {
		logger.Err(err).Str("chart", name).Warn("Unable to extract chart default values")
	}

	// Remove internal platform keys not meant for user configuration
	delete(defaults, "arangodb_platform")

	// Merge platform.yaml overrides on top of chart defaults
	if len(packageSpec.Overrides) > 0 {
		var overrides map[string]interface{}
		if err := json.Unmarshal(packageSpec.Overrides, &overrides); err == nil {
			for k, v := range overrides {
				defaults[k] = v
			}
		}
	}

	// A chart that ships a values.schema.json we cannot parse is a chart bug: silently
	// degrading would drop override validation without any signal, so fail the packaging.
	// A chart shipping no schema at all is fine and stays permissive.
	schema, err := extractChartSchema(chart)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid values.schema.json in chart %s", name)
	}

	return &packageChartRenderInputChart{
		Name:             name,
		Version:          packageSpec.Version,
		ChartData:        chart,
		Values:           defaults,
		Schema:           schema,
		DocumentedValues: documentedValues(defaults, schema),
	}, nil
}

// extractChartFile reads a chart's own top-level file (`<chart>/<name>`) from a gzipped
// tar chart archive. Files belonging to bundled subcharts (`<chart>/charts/<sub>/<name>`)
// are ignored. Returns nil data when the file is not present.
func extractChartFile(chartData []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(chartData))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Tar entries always use forward slashes, so split on "/" explicitly. Note that
		// filepath.SplitList must not be used here: it splits on the OS path-list
		// separator (":" on Linux), so it never matches and would let nested subchart
		// files through.
		parts := goStrings.Split(path.Clean(hdr.Name), "/")
		if len(parts) == 2 && parts[1] == name {
			data, err := io.ReadAll(tr)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
	}

	return nil, nil
}

// extractChartValues reads values.yaml from a gzipped tar chart archive.
func extractChartValues(chartData []byte) (map[string]interface{}, error) {
	data, err := extractChartFile(chartData, "values.yaml")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return map[string]interface{}{}, nil
	}

	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	if values == nil {
		return map[string]interface{}{}, nil
	}

	return values, nil
}

// extractChartSchema reads values.schema.json from a gzipped tar chart archive.
// Returns nil when the chart does not ship a schema.
func extractChartSchema(chartData []byte) (map[string]interface{}, error) {
	data, err := extractChartFile(chartData, "values.schema.json")
	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, err
	}

	return schema, nil
}

// documentedValues flattens a chart's effective values (its packaged defaults with the
// platform.yaml overrides already applied) into individual override paths, described using
// the chart's own values.schema.json where it provides descriptions. Values belong to the
// chart rather than to a service, since a chart may back several services - or none.
func documentedValues(values, schema map[string]interface{}) []packageChartRenderInputValue {
	if len(values) == 0 {
		return nil
	}

	var out []packageChartRenderInputValue

	flattenValues("", values, schema, &out, 0)

	return out
}

// documentedValuesMaxDepth bounds how deep nested values are expanded, so a pathological
// chart cannot produce an unreadable table.
const documentedValuesMaxDepth = 6

// flattenValues expands nested values into dotted key paths so every leaf is documented
// with its own default, rather than collapsing a subtree into an opaque JSON blob. An
// intermediate object is listed only when the schema documents it, so its description is
// not lost; its leaves follow underneath.
func flattenValues(prefix string, values map[string]interface{}, schema map[string]interface{}, out *[]packageChartRenderInputValue, depth int) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		path := k
		if prefix != "" {
			path = prefix + "." + k
		}

		child := schemaProperty(schema, k)
		description := schemaDescription(child)

		// Only descend into non-empty objects; everything else (scalar, list, empty
		// object) is a leaf and gets its own row.
		if nested, ok := values[k].(map[string]interface{}); ok && len(nested) > 0 && depth < documentedValuesMaxDepth {
			if description != "" {
				*out = append(*out, packageChartRenderInputValue{
					Key:         path,
					Description: description,
				})
			}

			flattenValues(path, nested, child, out, depth+1)

			continue
		}

		*out = append(*out, packageChartRenderInputValue{
			Key:         path,
			Default:     formatValueDefault(values[k]),
			Description: description,
		})
	}
}

// schemaProperty returns the subschema documenting a property of the given schema node.
func schemaProperty(schema map[string]interface{}, key string) map[string]interface{} {
	props, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return nil
	}

	prop, _ := props[key].(map[string]interface{})

	return prop
}

// schemaDescription returns the description a schema node documents, if any.
func schemaDescription(schema map[string]interface{}) string {
	description, _ := schema["description"].(string)

	return escapeTableCell(description)
}

// formatValueDefault renders a value compactly for a Markdown table cell. Nested objects
// and lists are shown as truncated JSON, since the full tree belongs in values.yaml.
func formatValueDefault(v interface{}) string {
	if s, ok := v.(string); ok {
		if s == "" {
			return `""`
		}
		return escapeTableCell(s)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}

	s := string(data)
	const max = 48
	if len(s) > max {
		s = s[:max] + "..."
	}

	return escapeTableCell(s)
}

// escapeJSONPointer escapes a single JSON Pointer reference token (RFC 6901).
func escapeJSONPointer(in string) string {
	in = goStrings.ReplaceAll(in, "~", "~0")

	return goStrings.ReplaceAll(in, "/", "~1")
}

// escapeTableCell keeps a value from breaking out of a Markdown table row.
func escapeTableCell(s string) string {
	s = goStrings.ReplaceAll(s, "|", "\\|")
	return goStrings.ReplaceAll(s, "\n", " ")
}

// sanitizeOverrideSchema adapts a chart's own values.schema.json so it can be inlined as
// the schema of an override block. Two adjustments are needed:
//   - `required` is dropped: an override block is a partial document merged on top of the
//     chart defaults, so a value being mandatory for the chart does not mean the user must
//     restate it here.
//   - `$schema`/`$id` are dropped: the schema becomes a subschema rather than a document
//     root, and keeping them would re-declare the dialect / shift the base URI.
//
// Recursion only descends into positions that actually hold schemas, so a chart value
// legitimately named "required" (a key under `properties`) is preserved.
//
// refBase is the JSON Pointer at which the schema is being inlined. Local `$ref`s are
// rewritten onto it, because once inlined `#/` no longer means the chart's own schema but
// the generated release schema - a chart referencing `#/definitions/image` would otherwise
// produce a release chart that fails validation against its own defaults.
func sanitizeOverrideSchema(in map[string]interface{}, refBase string) map[string]interface{} {
	out := make(map[string]interface{}, len(in))

	for k, v := range in {
		switch k {
		case "required", "$schema", "$id":
			continue
		case "$ref":
			if ref, ok := v.(string); ok && goStrings.HasPrefix(ref, "#") {
				out[k] = refBase + goStrings.TrimPrefix(ref, "#")
				continue
			}
			out[k] = v
		case "properties", "patternProperties", "$defs", "definitions":
			// Maps keyed by property name - keys are data, values are schemas.
			if m, ok := v.(map[string]interface{}); ok {
				sub := make(map[string]interface{}, len(m))
				for name, s := range m {
					if sm, ok := s.(map[string]interface{}); ok {
						sub[name] = sanitizeOverrideSchema(sm, refBase)
					} else {
						sub[name] = s
					}
				}
				out[k] = sub
				continue
			}
			out[k] = v
		case "allOf", "anyOf", "oneOf", "prefixItems":
			// Lists of schemas.
			if l, ok := v.([]interface{}); ok {
				sub := make([]interface{}, len(l))
				for i, s := range l {
					if sm, ok := s.(map[string]interface{}); ok {
						sub[i] = sanitizeOverrideSchema(sm, refBase)
					} else {
						sub[i] = s
					}
				}
				out[k] = sub
				continue
			}
			out[k] = v
		case "additionalProperties", "items", "not", "if", "then", "else", "contains", "propertyNames":
			// Single nested schema.
			if sm, ok := v.(map[string]interface{}); ok {
				out[k] = sanitizeOverrideSchema(sm, refBase)
				continue
			}
			out[k] = v
		default:
			out[k] = v
		}
	}

	return out
}

// generateValuesSchema builds a JSON Schema for values.yaml.
// A chart entry is validated against the chart's own values.schema.json when it ships one
// (relaxed for partial overrides, see sanitizeOverrideSchema); charts without a schema fall
// back to a permissive entry (additionalProperties: true). Service entries stay permissive.
// The deployment field is required.
func generateValuesSchema(input packageChartRenderInput) []byte {
	chartProps := map[string]interface{}{}
	for _, c := range input.Charts {
		description := "Overrides for chart " + c.Name + " (version " + c.Version + ")"

		if len(c.Schema) > 0 {
			s := sanitizeOverrideSchema(c.Schema, "#/properties/charts/properties/"+escapeJSONPointer(c.Name))
			s["description"] = description
			chartProps[c.Name] = s
			continue
		}

		chartProps[c.Name] = map[string]interface{}{
			"type":                 "object",
			"description":          description,
			"additionalProperties": true,
		}
	}

	serviceProps := map[string]interface{}{}
	for _, s := range input.Services {
		serviceProps[s.Name] = map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"values": map[string]interface{}{
					"type":                 "object",
					"description":          "Value overrides for service " + s.Name,
					"additionalProperties": true,
				},
			},
			"additionalProperties": false,
		}
	}

	schema := map[string]interface{}{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"type":        "object",
		"description": "Values schema for " + input.Name + " " + input.Version,
		"required":    []string{"deployment"},
		"properties": map[string]interface{}{
			"deployment": map[string]interface{}{
				"type":        "string",
				"description": "Name of the ArangoDeployment to target",
				"minLength":   1,
			},
			"charts": map[string]interface{}{
				"type":                 "object",
				"description":          "Per-chart value overrides",
				"properties":           chartProps,
				"additionalProperties": false,
			},
			"services": map[string]interface{}{
				"type":                 "object",
				"description":          "Per-service value overrides",
				"properties":           serviceProps,
				"additionalProperties": false,
			},
		},
		"additionalProperties": false,
	}

	data, _ := json.MarshalIndent(schema, "", "  ")
	return data
}
