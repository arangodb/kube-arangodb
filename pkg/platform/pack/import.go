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
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/ref"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func Import(ctx context.Context, path string, client *regclient.RegClient, registry string) (Proto, *helm.Package, error) {
	out, err := os.Open(path)
	if err != nil {
		return Proto{}, nil, err
	}

	stat, err := out.Stat()
	if err != nil {
		return Proto{}, nil, err
	}

	in, err := zip.NewReader(out, stat.Size())
	if err != nil {
		return Proto{}, nil, err
	}

	i := &importPackageSet{
		in:       in,
		client:   client,
		registry: registry,
	}

	data, err := i.Read("proto.yaml")
	if err != nil {
		return Proto{}, nil, err
	}

	proto, err := util.JsonOrYamlUnmarshal[Proto](data)
	if err != nil {
		return Proto{}, nil, err
	}

	if err := executor.Run(ctx, logger, 8, i.run(proto)); err != nil {
		return Proto{}, nil, err
	}

	return proto, &i.p, nil
}

type importPackageSet struct {
	lock sync.Mutex

	in *zip.Reader

	client *regclient.RegClient

	p        helm.Package
	registry string
}

func (i *importPackageSet) run(p Proto) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		for k, v := range p.Manifests {
			src, err := ref.New(fmt.Sprintf("%s/%s", i.registry, k))
			if err != nil {
				return err
			}

			h.RunAsync(ctx, i.importManifest(src, v))
		}

		h.WaitForSubThreads(t)

		for k, v := range p.Charts {
			var pkgS helm.PackageSpec

			data, err := i.Read("chart/%s-%s.tgz", k, v.Version)
			if err != nil {
				return err
			}

			pkgS.Chart = data
			pkgS.Version = v.Version

			var versions ProtoValues

			versions.Images = map[string]ProtoImage{}

			for z, q := range v.Images {
				var img = q

				img.Registry = util.NewType(i.registry)

				versions.Images[z] = img
			}

			vData, err := helm.NewValues(versions)
			if err != nil {
				return err
			}

			pkgS.Overrides = vData

			i.withPackage(func(in helm.Package) helm.Package {
				if in.Packages == nil {
					in.Packages = map[string]helm.PackageSpec{}
				}

				in.Packages[k] = pkgS

				return in
			})
		}

		return nil
	}
}

func (i *importPackageSet) importManifest(src ref.Ref, blob string) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("manifest", src.CommonName()).Wrap(logging.WithElapsed("duration"))

		log.Info("Importing manifest")
		defer func() {
			log.Info("Imported manifest")
		}()

		data, err := i.Read("manifests/%s", blob)
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
					h.RunAsync(ctx, i.importManifest(src.SetDigest(entry.Digest.String()), entry.Digest.Hex()))

				case mediatype.Docker2ImageConfig, mediatype.OCI1ImageConfig,
					mediatype.Docker2Layer, mediatype.Docker2LayerGzip, mediatype.Docker2LayerZstd,
					mediatype.OCI1Layer, mediatype.OCI1LayerGzip, mediatype.OCI1LayerZstd,
					mediatype.BuildkitCacheConfig:
					h.RunAsync(ctx, i.importBlob(src, entry))
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
				h.RunAsync(ctx, i.importBlob(src, cd))
			}

			layers, err := manifestIndex.GetLayers()
			if err != nil {
				return err
			}

			for _, m := range layers {
				h.RunAsync(ctx, i.importBlob(src, m))
			}
		}

		h.WaitForSubThreads(t)

		if err := i.client.ManifestPut(ctx, src, m); err != nil {
			return err
		}

		return nil
	}
}

func (i *importPackageSet) importBlob(src ref.Ref, desc descriptor.Descriptor) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("blob", desc.Digest.Hex()).Wrap(logging.WithElapsed("duration")).Int64("size", desc.Size)

		log.Info("Importing blob")
		defer func() {
			log.Info("Imported blob")
		}()

		// Upload
		qs, err := i.Open("blobs/%s", desc.Digest.Hex())
		if err != nil {
			return err
		}

		if _, err := i.client.BlobPut(ctx, src, desc, qs); err != nil {
			return err
		}

		return nil
	}
}

func (i *importPackageSet) Open(name string, args ...any) (io.Reader, error) {
	return i.in.Open(fmt.Sprintf(name, args...))
}

func (i *importPackageSet) Read(name string, args ...any) ([]byte, error) {
	reader, err := i.in.Open(fmt.Sprintf(name, args...))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(reader)
}

func (i *importPackageSet) withPackage(mod util.ModR[helm.Package]) {
	i.lock.Lock()
	defer i.lock.Unlock()

	i.p = mod(i.p)
}
