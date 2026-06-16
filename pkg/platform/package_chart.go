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
	"path/filepath"

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
}

type packageChartRenderInputService struct {
	Name     string
	ChartRef string
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

	return &packageChartRenderInputChart{
		Name:      name,
		Version:   packageSpec.Version,
		ChartData: chart,
		Values:    defaults,
	}, nil
}

// extractChartValues reads values.yaml from a gzipped tar chart archive.
func extractChartValues(chartData []byte) (map[string]interface{}, error) {
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

		if filepath.Base(hdr.Name) == "values.yaml" {
			// Only match top-level values.yaml (e.g. chartname/values.yaml)
			parts := filepath.SplitList(hdr.Name)
			if len(parts) <= 2 || filepath.Dir(hdr.Name) == filepath.Dir(filepath.Dir(hdr.Name)) {
				data, err := io.ReadAll(tr)
				if err != nil {
					return nil, err
				}
				var values map[string]interface{}
				if err := yaml.Unmarshal(data, &values); err != nil {
					return nil, err
				}
				return values, nil
			}
		}
	}

	return map[string]interface{}{}, nil
}

// generateValuesSchema builds a JSON Schema for values.yaml.
// Charts and services entries use permissive schemas (additionalProperties: true)
// so users can override any value. The deployment field is required.
func generateValuesSchema(input packageChartRenderInput) []byte {
	chartProps := map[string]interface{}{}
	for _, c := range input.Charts {
		chartProps[c.Name] = map[string]interface{}{
			"type":                 "object",
			"description":          "Overrides for chart " + c.Name + " (version " + c.Version + ")",
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
