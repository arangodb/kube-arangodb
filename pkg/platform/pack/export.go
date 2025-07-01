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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/errs"
	"github.com/regclient/regclient/types/manifest"
	"github.com/regclient/regclient/types/mediatype"
	"github.com/regclient/regclient/types/ref"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func Export(ctx context.Context, path string, m helm.ChartManager, client *regclient.RegClient, p helm.Package, images ...ProtoImage) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}

	tw := zip.NewWriter(out)

	var r = exportPackageSet{
		m:         m,
		images:    images,
		client:    client,
		wr:        tw,
		existence: map[string]bool{},
	}

	if err := executor.Run(ctx, logger, 8, r.run(p)); err != nil {
		return err
	}

	if err := r.saveProto(); err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := out.Close(); err != nil {
		return err
	}

	return nil
}

type exportPackageSet struct {
	lock sync.Mutex

	proto Proto

	m      helm.ChartManager
	client *regclient.RegClient

	images []ProtoImage

	existence map[string]bool

	wr *zip.Writer
}

func (r *exportPackageSet) run(p helm.Package) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		for k, v := range p.Packages {
			h.RunAsync(ctx, r.exportPackage(k, v))
		}

		for _, image := range r.images {
			h.RunAsync(ctx, r.exportImage(image))
		}

		return nil
	}
}

func (r *exportPackageSet) exportPackage(name string, spec helm.PackageSpec) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("chart", name).Str("version", spec.Version).Wrap(logging.WithElapsed("chartDuration"))

		log.Info("Extracting chart")
		defer func() {
			log.Info("Extracted chart")
		}()

		var chart helm.Chart

		if spec.Chart.IsZero() {
			repo, ok := r.m.Get(name)
			if !ok {
				return errors.Errorf("Chart `%s` not found", name)
			}

			ver, ok := repo.Get(spec.Version)
			if !ok {
				return errors.Errorf("Chart `%s=%s` not found", name, spec.Version)
			}

			c, err := ver.Get(ctx)
			if err != nil {
				return err
			}

			chart = c
		} else {
			chart = helm.Chart(spec.Chart)
		}

		var chartProto ProtoChart

		chartProto.Images = map[string]ProtoImage{}
		chartProto.Version = spec.Version

		chartData, err := chart.Get()
		if err != nil {
			return err
		}

		type valuesInterface struct {
			Images ProtoImages `json:"images,omitempty"`
		}

		protoImages, err := util.JSONRemarshal[map[string]any, valuesInterface](chartData.Chart().Values)
		if err != nil {
			return err
		}

		for k, v := range protoImages.Images {
			if v.IsTest() {
				logger.Str("image", v.GetImage()).Info("Skip Test Image")
				continue
			}

			h.RunAsync(ctx, r.exportImage(v))

			v.Registry = nil
			v.Kind = ""

			chartProto.Images[k] = v
		}

		h.WaitForSubThreads(t)

		r.withProto(func(in Proto) Proto {
			if in.Charts == nil {
				in.Charts = map[string]ProtoChart{}
			}

			in.Charts[name] = chartProto

			return in
		})

		return r.writeOutFile(bytes.NewReader(chart.Raw()), "chart/%s-%s.tgz", name, spec.Version)
	}
}

func (r *exportPackageSet) exportImage(image ProtoImage) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("image", image.GetImage()).Wrap(logging.WithElapsed("duration"))

		log.Info("Extracting image")
		defer func() {
			log.Info("Extracted image")
		}()

		src, err := ref.New(image.GetImage())
		if err != nil {
			return err
		}

		h.RunAsync(ctx, r.exportManifest(src))

		h.WaitForSubThreads(t)

		m, err := r.client.ManifestGet(ctx, src)
		if err != nil {
			return err
		}

		r.withProto(func(in Proto) Proto {
			if in.Manifests == nil {
				in.Manifests = map[string]string{}
			}

			in.Manifests[image.GetShortImage()] = m.GetDescriptor().Digest.Hex()

			return in
		})

		return nil
	}
}

func (r *exportPackageSet) exportManifest(src ref.Ref) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("manifest", src.CommonName()).Wrap(logging.WithElapsed("duration"))

		log.Info("Extracting manifest")
		defer func() {
			log.Info("Extracted manifest")
		}()

		m, err := r.client.ManifestGet(ctx, src)
		if err != nil {
			return err
		}

		if !r.once("manifests/%s", m.GetDescriptor().Digest.Hex()) {
			return nil
		}

		if manifestIndex, ok := m.(manifest.Indexer); ok && m.IsSet() {
			manifests, err := manifestIndex.GetManifestList()
			if err != nil {
				return err
			}

			for _, entry := range manifests {
				switch entry.MediaType {
				case mediatype.Docker1Manifest, mediatype.Docker1ManifestSigned,
					mediatype.Docker2Manifest, mediatype.Docker2ManifestList,
					mediatype.OCI1Manifest, mediatype.OCI1ManifestList:
					h.RunAsync(ctx, r.exportManifest(src.SetDigest(entry.Digest.String())))
				case mediatype.Docker2ImageConfig, mediatype.OCI1ImageConfig,
					mediatype.Docker2Layer, mediatype.Docker2LayerGzip, mediatype.Docker2LayerZstd,
					mediatype.OCI1Layer, mediatype.OCI1LayerGzip, mediatype.OCI1LayerZstd,
					mediatype.BuildkitCacheConfig:
					h.RunAsync(ctx, r.exportBlob(src, entry))
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
				h.RunAsync(ctx, r.exportBlob(src, cd))
			}

			layers, err := manifestIndex.GetLayers()
			if err != nil {
				return err
			}

			for _, m := range layers {
				h.RunAsync(ctx, r.exportBlob(src, m))
			}
		}

		h.WaitForSubThreads(t)

		return r.writeOutData(m.MarshalJSON, "manifests/%s", m.GetDescriptor().Digest.Hex())
	}
}

func (r *exportPackageSet) exportBlob(src ref.Ref, desc descriptor.Descriptor) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("blob", desc.Digest.Hex()).Wrap(logging.WithElapsed("duration")).Int64("size", desc.Size)

		log.Info("Extracting blob")
		defer func() {
			log.Info("Extracted blob")
		}()

		if o, err := r.client.BlobGet(ctx, src, desc); err != nil {
			return err
		} else {
			return r.exportBlobData(desc, o)
		}
	}
}

func (r *exportPackageSet) exportBlobData(desc descriptor.Descriptor, blob io.ReadCloser) error {
	if !r.once("blobs/%s", desc.Digest.Hex()) {
		return nil
	}

	f, err := os.CreateTemp("", "tmp-")
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, blob); err != nil {
		return nil
	}

	if err := blob.Close(); err != nil {
		return err
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}

	if err := r.writeOutFile(f, "blobs/%s", desc.Digest.Hex()); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Remove(f.Name())
}

func (r *exportPackageSet) writeOutData(in func() ([]byte, error), f string, args ...any) error {
	data, err := in()
	if err != nil {
		return err
	}

	return r.writeOutFile(bytes.NewReader(data), f, args...)
}

func (r *exportPackageSet) withProto(mod util.ModR[Proto]) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.proto = mod(r.proto)
}

func (r *exportPackageSet) writeOutFile(in io.Reader, f string, args ...any) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	q, err := r.wr.Create(fmt.Sprintf(f, args...))
	if err != nil {
		return err
	}

	if _, err := io.Copy(q, in); err != nil {
		return err
	}

	return nil
}

func (r *exportPackageSet) saveProto() error {
	out, err := r.wr.Create("proto.yaml")
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(r.proto)
	if err != nil {
		return err
	}

	if _, err := out.Write(data); err != nil {
		return err
	}

	return nil
}

func (r *exportPackageSet) once(f string, args ...any) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	k := fmt.Sprintf(f, args...)

	_, ok := r.existence[k]
	if !ok {
		return true
	}

	r.existence[k] = true

	return false
}
