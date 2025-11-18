//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package pack

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	goHttp "net/http"
	"os"
	goStrings "strings"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/ref"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func ResolvePackageSpec(ctx context.Context, endpoint, name string, in helm.PackageSpec, reg *regclient.RegClient, chttp *goHttp.Client) (helm.Chart, error) {
	if err := in.Validate(); err != nil {
		return helm.Chart{}, err
	}

	switch in.PackageType() {
	case helm.PackageTypePlatform:
		if in.GetStage() == "prd" {
			endpoint = fmt.Sprintf("helm.%s", endpoint)
		} else {
			endpoint = fmt.Sprintf("%s.helm.%s", in.GetStage(), endpoint)
		}

		r, err := ref.New(fmt.Sprintf("%s/%s:%s", endpoint, name, in.Version))
		if err != nil {
			return helm.Chart{}, err
		}

		return ExportChart(ctx, reg, r)
	case helm.PackageTypeInline:
		return base64.StdEncoding.DecodeString(util.WithDefault(in.Chart))
	case helm.PackageTypeRemote:
		e := util.WithDefault(in.Chart)

		if chttp == nil {
			chttp = goHttp.DefaultClient
		}

		req, err := goHttp.NewRequestWithContext(ctx, "GET", e, nil)
		if err != nil {
			return helm.Chart{}, err
		}

		resp, err := chttp.Do(req)
		if err != nil {
			return helm.Chart{}, err
		}

		defer resp.Body.Close()

		return io.ReadAll(resp.Body)

	case helm.PackageTypeFile:
		e := util.WithDefault(in.Chart)

		e = goStrings.TrimPrefix(e, "file://")

		return os.ReadFile(e)

	case helm.PackageTypeOCI:
		e := util.WithDefault(in.Chart)
		e = goStrings.TrimPrefix(e, "oci://")

		r, err := ref.New(e)
		if err != nil {
			return helm.Chart{}, err
		}

		return ExportChart(ctx, reg, r)
	case helm.PackageTypeIndex:
		e := util.WithDefault(in.Chart)
		e = goStrings.TrimPrefix(e, "index://")

		hm, err := helm.NewChartManager(ctx, chttp, "https://%s/index.yaml", e)
		if err != nil {
			return helm.Chart{}, err
		}

		p, ok := hm.Get(name)
		if !ok {
			return nil, errors.Errorf("Package %s not found", name)
		}

		v, ok := p.Get(in.Version)
		if !ok {
			return nil, errors.Errorf("Package %s version %s not found", name, in.Version)
		}

		return v.Get(ctx)
	default:
		return helm.Chart{}, fmt.Errorf("invalid package type")
	}
}

func ExportChart(ctx context.Context, client *regclient.RegClient, src ref.Ref) (helm.Chart, error) {
	m, err := client.ManifestGet(ctx, src)
	if err != nil {
		return nil, err
	}

	if m.GetMediaType() != mediatype.OCI1Manifest {
		return nil, errors.Errorf("Manifest is not %s, got %s", mediatype.OCI1Manifest, m.GetMediaType())
	}

	layers, err := m.GetLayers()
	if err != nil {
		return nil, err
	}

	if len(layers) != 1 {
		return nil, errors.Errorf("Expected one layer in the OCI")
	}

	layer := layers[0]

	if layer.MediaType != "application/vnd.cncf.helm.chart.content.v1.tar+gzip" {
		return nil, errors.Errorf("Manifest is not %s, got %s", "application/vnd.cncf.helm.chart.content.v1.tar+gzip", layer.MediaType)
	}

	o, err := client.BlobGet(ctx, src, layer)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(o)
	if err != nil {
		return nil, err
	}

	if err := o.Close(); err != nil {
		return nil, err
	}

	return data, nil
}
