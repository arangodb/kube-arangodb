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
	"context"
	_ "embed"
	"os"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/config"
	"github.com/spf13/cobra"

	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
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
	packageChartTemplateResourceChart util.Template[packageChartRenderInput]

	//go:embed templates/resource.service.yaml
	packageChartTemplateResourceService util.Template[packageChartRenderInput]
)

type packageChartRenderInput struct {
	Name string

	Charts   map[string]packageChartRenderInputChart
	Services map[string]packageChartRenderInputService
}
type packageChartRenderInputChart struct {
	Name    string
	Version string
	Spec    platformApi.ArangoPlatformChartSpec
}

type packageChartRenderInputService struct {
	Name string
	Spec platformApi.ArangoPlatformServiceSpec
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

	input := packageChartRenderInput{
		Name: "arango-platform-release",
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

	builder := util.NewGZipBuilder(out)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateChart, input), "%s/Chart.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateValues, input), "%s/values.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateCheck, input), "%s/templates/check.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateNotes, input), "%s/templates/NOTES.txt", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateResourceChart, input), "%s/templates/ArangoPlatformChart.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessTemplate(packageChartTemplateResourceService, input), "%s/templates/ArangoPlatformService.yaml", input.Name)

	builder = builder.File(util.GZipBuilderProcessBytes(licenseData.Full()), "%s/LICENSE", input.Name)

	//	builder.YAMLFile("values.yaml", map[string]any{
	//		"deployment": "",
	//	})
	//
	//	builder.File("templates/check.yaml", []byte(`
	//{{- if not .Values.deployment }}
	//{{- fail "Deployment needs to be defined" }}
	//{{- end }}
	//`))
	//

	//
	//	for k, v := range r.Releases {
	//		rel := packageChartRelease(k, v)
	//		builder.YAMLFile(fmt.Sprintf("templates/release/%s.yaml", rel.GetName()), rel)
	//		lines = append(lines, fmt.Sprintf("ArangoPlatformService %s", rel.GetName()))
	//	}
	//
	//	builder.File("templates/NOTES.txt", []byte(fmt.Sprintf(`
	//Arango Platform Release has been installed!
	//
	//Components:
	//%s
	//`, goStrings.Join(lines, "\n"))))

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
		Name: name,
		Spec: platformApi.ArangoPlatformServiceSpec{
			Deployment: &sharedApi.Object{
				Name: "{{ .Values.deployment }}",
			},
			Chart: &sharedApi.Object{
				Name: packageSpec.Package,
			},
			Values: sharedApi.Any(packageSpec.Overrides),
		},
	}
}

func packageChartChart(ctx context.Context, reg *regclient.RegClient, endpoint string, name string, packageSpec helm.PackageSpec) (*packageChartRenderInputChart, error) {
	chart, err := pack.ResolvePackageSpec(ctx, endpoint, name, packageSpec, reg, nil)
	if err != nil {
		return nil, err
	}

	return &packageChartRenderInputChart{
		Name:    name,
		Version: packageSpec.Version,
		Spec: platformApi.ArangoPlatformChartSpec{
			Definition: sharedApi.Data(chart),
			Overrides:  sharedApi.Any(packageSpec.Overrides),
		},
	}, nil
}
