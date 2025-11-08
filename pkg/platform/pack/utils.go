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
	"fmt"
	"io"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/ref"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func ChartReference(endpoint, stage, name, version string) (ref.Ref, error) {
	if stage == "prd" {
		endpoint = fmt.Sprintf("helm.%s", endpoint)
	} else {
		endpoint = fmt.Sprintf("%s.helm.%s", stage, endpoint)
	}

	return ref.New(fmt.Sprintf("%s/%s:%s", endpoint, name, version))
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
