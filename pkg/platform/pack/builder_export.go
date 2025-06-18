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

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/ref"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func ExportPackage(ctx context.Context, path string, m helm.ChartManager, client *regclient.RegClient, p helm.Package) error {
	out := NewBuilder(path)

	for k, v := range p.Packages {
		out = ExportChart(ctx, out, m, client, k, v.Version)
		if out.HasError() {
			return out.Done()
		}
	}

	return out.Done()
}

func ExportChart(ctx context.Context, out Builder, m helm.ChartManager, client *regclient.RegClient, name, version string) Builder {
	logger := logger.Str("chart", name).Str("version", version).Wrap(logging.WithElapsed("chartDuration"))

	repo, ok := m.Get(name)
	if !ok {
		return out.WithError(errors.Errorf("Chart `%s` not found", name))
	}

	ver, ok := repo.Get(version)
	if !ok {
		return out.WithError(errors.Errorf("Chart `%s=%s` not found", name, version))
	}

	chart, err := ver.Get(ctx)
	if err != nil {
		return out.WithError(err)
	}

	var chartProto ProtoChart

	chartProto.Images = map[string]ProtoImage{}
	chartProto.Version = version

	chartData, err := chart.Get()
	if err != nil {
		return out.WithError(err)
	}

	type valuesInterface struct {
		Images ProtoImages `json:"images,omitempty"`
	}

	protoImages, err := util.JSONRemarshal[map[string]any, valuesInterface](chartData.Chart().Values)
	if err != nil {
		return out.WithError(err)
	}

	for k, v := range protoImages.Images {
		if v.IsTest() {
			logger.Str("image", v.GetImage()).Info("Skip Test Image")
			continue
		}

		out = ExportImage(ctx, out, client, v)
		if out.HasError() {
			return out
		}

		v.Registry = nil
		v.Kind = ""

		chartProto.Images[k] = v
	}

	return out.UpdateProto(func(in Proto) Proto {
		if in.Charts == nil {
			in.Charts = map[string]ProtoChart{}
		}

		in.Charts[name] = chartProto

		return in
	}).WithChart(name, version, chart.Raw())
}

func ExportImage(ctx context.Context, out Builder, client *regclient.RegClient, image ProtoImage) Builder {
	logger := logger.Str("image", image.GetImage()).Wrap(logging.WithElapsed("duration"))

	logger.Info("Extracting image")
	defer func() {
		logger.Info("Extracted image")
	}()

	src, err := ref.New(image.GetImage())
	if err != nil {
		return out.WithError(err)
	}

	out = ExportManifest(ctx, out, client, src)

	m, err := client.ManifestGet(ctx, src)
	if err != nil {
		return out.WithError(err)
	}

	return out.UpdateProto(func(in Proto) Proto {
		if in.Manifests == nil {
			in.Manifests = map[string]string{}
		}

		in.Manifests[image.GetShortImage()] = m.GetDescriptor().Digest.Hex()

		return in
	})
}

func ExportManifest(ctx context.Context, out Builder, client *regclient.RegClient, src ref.Ref) Builder {
	logger := logger.Str("manifest", src.CommonName()).Wrap(logging.WithElapsed("duration"))

	logger.Info("Extracting manifest")
	defer func() {
		logger.Info("Extracted manifest")
	}()

	m, err := client.ManifestGet(ctx, src)
	if err != nil {
		return out.WithError(err)
	}

	if manifestIndex, ok := m.(manifest.Indexer); ok && m.IsSet() {
		manifests, err := manifestIndex.GetManifestList()
		if err != nil {
			return out.WithError(err)
		}

		for _, entry := range manifests {
			if entry.Platform == nil {
				continue
			}

			switch entry.MediaType {
			case mediatype.Docker1Manifest, mediatype.Docker1ManifestSigned,
				mediatype.Docker2Manifest, mediatype.Docker2ManifestList,
				mediatype.OCI1Manifest, mediatype.OCI1ManifestList:
				out = ExportManifest(ctx, out, client, src.SetDigest(entry.Digest.String()))
			case mediatype.Docker2ImageConfig, mediatype.OCI1ImageConfig,
				mediatype.Docker2Layer, mediatype.Docker2LayerGzip, mediatype.Docker2LayerZstd,
				mediatype.OCI1Layer, mediatype.OCI1LayerGzip, mediatype.OCI1LayerZstd,
				mediatype.BuildkitCacheConfig:
				out = ExportBlob(ctx, out, client, src, entry)
			}
		}
	}

	if manifestIndex, ok := m.(manifest.Imager); ok && m.IsSet() {
		cd, err := manifestIndex.GetConfig()
		if err != nil {
			// docker schema v1 does not have a config object, ignore if it's missing
			if !errors.Is(err, errs.ErrUnsupportedMediaType) {
				return out.WithError(fmt.Errorf("failed to get config digest for %s: %w", src.CommonName(), err))
			}
		} else {
			out = ExportBlob(ctx, out, client, src, cd)
		}

		layers, err := manifestIndex.GetLayers()
		if err != nil {
			return out.WithError(err)
		}

		for _, m := range layers {
			out = ExportBlob(ctx, out, client, src, m)
		}
	}

	return out.WithManifest(m)
}

func ExportBlob(ctx context.Context, out Builder, client *regclient.RegClient, src ref.Ref, desc descriptor.Descriptor) Builder {
	logger := logger.Str("blob", desc.Digest.Hex()).Wrap(logging.WithElapsed("duration")).Int64("size", desc.Size)

	logger.Info("Extracting blob")
	defer func() {
		logger.Info("Extracted blob")
	}()

	if o, err := client.BlobGet(ctx, src, desc); err != nil {
		return out.WithError(err)
	} else {
		return out.WithBlob(desc, o)
	}
}
