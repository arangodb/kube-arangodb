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

func ImportPackage(ctx context.Context, path string, client *regclient.RegClient, registry string) (*helm.Package, error) {
	importer, err := NewImporter(path)
	if err != nil {
		return nil, err
	}

	proto, err := ReadProto(importer)
	if err != nil {
		return nil, err
	}

	var pkg helm.Package

	pkg.Packages = map[string]helm.PackageSpec{}

	type valuesInterface struct {
		Images ProtoImages `json:"images,omitempty"`
	}

	for k, v := range proto.Charts {
		var pkgS helm.PackageSpec

		data, err := importer.Read("chart/%s-%s.tgz", k, v.Version)
		if err != nil {
			return nil, err
		}

		pkgS.Chart = data
		pkgS.Version = v.Version

		var versions valuesInterface

		versions.Images = map[string]ProtoImage{}

		for i, q := range v.Images {
			var img ProtoImage = q

			img.Registry = util.NewType(registry)

			manBlob, ok := proto.Manifests[q.GetImage()]
			if !ok {
				return nil, errors.Errorf("Unable to find image: %s", q.GetImage())
			}
			rq, err := ref.New(fmt.Sprintf("%s/%s", registry, q.GetImage()))
			if err != nil {
				return nil, err
			}

			if err := ImportManifest(ctx, importer, client, rq, manBlob); err != nil {
				return nil, err
			}

			versions.Images[i] = img
		}

		vData, err := helm.NewValues(versions)
		if err != nil {
			return nil, err
		}

		pkgS.Overrides = vData

		pkg.Packages[k] = pkgS
	}

	return &pkg, nil
}

func ReadProto(im Importer) (Proto, error) {
	data, err := im.Read("proto.yaml")
	if err != nil {
		return Proto{}, err
	}

	return util.JsonOrYamlUnmarshal[Proto](data)
}

func ImportManifest(ctx context.Context, reader Importer, client *regclient.RegClient, src ref.Ref, blob string) error {
	logger := logger.Str("manifest", src.CommonName()).Wrap(logging.WithElapsed("duration"))

	logger.Info("Importing manifest")
	defer func() {
		logger.Info("Imported manifest")
	}()

	data, err := reader.Read("manifest/%s", blob)
	if err != nil {
		return err
	}

	m, err := manifest.New(
		manifest.WithRaw(data),
		manifest.WithRef(src),
	)
	if err != nil {
		return err
	}

	if manifestIndex, ok := m.(manifest.Indexer); ok && m.IsSet() {
		manifests, err := manifestIndex.GetManifestList()
		if err != nil {
			return err
		}

		for _, entry := range manifests {
			if entry.Platform == nil {
				continue
			}

			switch entry.MediaType {
			case mediatype.Docker1Manifest, mediatype.Docker1ManifestSigned,
				mediatype.Docker2Manifest, mediatype.Docker2ManifestList,
				mediatype.OCI1Manifest, mediatype.OCI1ManifestList:
				if err := ImportManifest(ctx, reader, client, src.SetDigest(entry.Digest.String()), entry.Digest.Hex()); err != nil {
					return err
				}
			case mediatype.Docker2ImageConfig, mediatype.OCI1ImageConfig,
				mediatype.Docker2Layer, mediatype.Docker2LayerGzip, mediatype.Docker2LayerZstd,
				mediatype.OCI1Layer, mediatype.OCI1LayerGzip, mediatype.OCI1LayerZstd,
				mediatype.BuildkitCacheConfig:
				if err := ImportBlob(ctx, reader, client, src, entry); err != nil {
					return err
				}
			}
		}
	}

	if manifestIndex, ok := m.(manifest.Imager); ok && m.IsSet() {
		cd, err := manifestIndex.GetConfig()
		if err != nil {
			// docker schema v1 does not have a config object, ignore if it's missing
			if !errors.Is(err, errs.ErrUnsupportedMediaType) {
				return fmt.Errorf("failed to get config digest for %s: %w", src.CommonName(), err)
			}
		} else {
			if err := ImportBlob(ctx, reader, client, src, cd); err != nil {
				return err
			}
		}

		layers, err := manifestIndex.GetLayers()
		if err != nil {
			return err
		}

		for _, m := range layers {
			if err := ImportBlob(ctx, reader, client, src, m); err != nil {
				return err
			}
		}
	}

	if err := client.ManifestPut(ctx, src, m); err != nil {
		return err
	}

	return nil
}

func ImportBlob(ctx context.Context, reader Importer, client *regclient.RegClient, src ref.Ref, desc descriptor.Descriptor) error {
	logger := logger.Str("blob", desc.Digest.Hex()).Wrap(logging.WithElapsed("duration")).Int64("size", desc.Size)

	logger.Info("Importing blob")
	defer func() {
		logger.Info("Imported blob")
	}()

	// Upload
	qs, err := reader.Open("bloob/%s", desc.Digest.Hex())
	if err != nil {
		return err
	}
	if _, err := client.BlobPut(ctx, src, desc, qs); err != nil {
		return err
	}

	return nil
}
