//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	"sync"
	"time"

	"github.com/regclient/regclient"
	"github.com/regclient/regclient/types"
	"github.com/regclient/regclient/types/ref"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/executor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func Copy(ctx context.Context, registry, dest string, client *regclient.RegClient, p helm.Package, images ...ProtoImage) (*helm.Package, error) {
	var r = copyPackageSet{
		client:   client,
		images:   images,
		registry: registry,
		dest:     dest,
	}

	if err := executor.Run(ctx, logger, 1, r.run(p)); err != nil {
		return nil, err
	}

	return &r.p, nil
}

type copyPackageSet struct {
	lock sync.Mutex

	client *regclient.RegClient

	images []ProtoImage

	p helm.Package

	registry string
	dest     string
}

func (c *copyPackageSet) run(p helm.Package) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		for n, z := range p.Packages {
			h.RunAsync(ctx, c.copyPackage(n, z))
		}

		return nil
	}
}

func (r *copyPackageSet) copyPackage(name string, spec helm.PackageSpec) executor.RunFunc {
	return func(ctx context.Context, log logging.Logger, t executor.Thread, h executor.Handler) error {
		log = log.Str("chart", name).Str("version", spec.Version).Wrap(logging.WithElapsed("chartDuration"))

		log.Info("Copying chart")
		defer func() {
			log.Info("Copied chart")
		}()

		chart, err := ResolvePackageSpec(ctx, r.registry, name, spec, r.client, nil)
		if err != nil {
			return err
		}

		var pkgS helm.PackageSpec

		pkgS.Chart = util.NewType(base64.StdEncoding.EncodeToString(chart))
		pkgS.Version = spec.Version

		var versions ProtoValues

		versions.Images = map[string]ProtoImage{}

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

			var img = v

			img.Registry = util.NewType(r.dest)

			if err := r.exportImage(ctx, logger, v, img); err != nil {
				return err
			}

			versions.Images[k] = img
		}

		vData, err := helm.NewValues(versions)
		if err != nil {
			return err
		}

		pkgS.Overrides = vData

		r.withPackage(func(in helm.Package) helm.Package {
			if len(in.Packages) == 0 {
				in.Packages = map[string]helm.PackageSpec{}
			}
			in.Packages[name] = pkgS
			return in
		})

		return nil
	}
}

func (r *copyPackageSet) exportImage(ctx context.Context, log logging.Logger, src, dest ProtoImage) error {
	srcImage, err := ref.New(src.GetImage())
	if err != nil {
		return err
	}
	destImage, err := ref.New(dest.GetImage())
	if err != nil {
		return err
	}

	log = log.Str("source", srcImage.CommonName()).Str("destination", destImage.CommonName())

	ct := copyTracker{
		blobs: map[string]copyBlobTracker{},
	}

	c := ct.EnableLogger(log)
	defer c()

	if err := r.client.ImageCopy(ctx, srcImage, destImage, ct.ImageCallback()); err != nil {
		return err
	}
	return nil
}

func (c *copyPackageSet) withPackage(mod util.ModR[helm.Package]) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.p = mod(c.p)
}

type copyBlobTracker struct {
	Current int64
	Total   int64
}

func (c copyBlobTracker) Completed() float64 {
	if c.Total == 0 {
		return 0
	}
	return float64(c.Current) / float64(c.Total)
}

type copyTracker struct {
	lock sync.Mutex

	blobs map[string]copyBlobTracker
}

func (c *copyTracker) ImageCallback() regclient.ImageOpts {
	return regclient.ImageWithCallback(c.Log)
}

func (c *copyTracker) sizes() copyBlobTracker {
	c.lock.Lock()
	defer c.lock.Unlock()

	var q copyBlobTracker

	for _, v := range c.blobs {
		q.Total += v.Total
		q.Current += v.Current
	}

	return q
}

func (c *copyTracker) EnableLogger(log logging.Logger) context.CancelFunc {
	done := make(chan struct{})

	ctx, cc := context.WithCancel(context.Background())

	go func() {
		defer close(done)
		log.Info("Copy Started")

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				b := c.sizes()
				log.Int64("current", b.Current).Int64("total", b.Total).Info("In progress: %0.2f%%", b.Completed()*100)
			case <-ctx.Done():
				log.Info("Copy Completed")
				return
			}
		}
	}()

	return func() {
		cc()

		<-done
	}
}

func (c *copyTracker) Log(kind types.CallbackKind, instance string, state types.CallbackState, cur, total int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	switch kind {
	case types.CallbackManifest:
		return
	case types.CallbackBlob:
		switch state {
		case types.CallbackStarted:
			c.blobs[instance] = copyBlobTracker{
				Total: total,
			}
			return
		case types.CallbackActive:
			c.blobs[instance] = copyBlobTracker{
				Current: cur,
				Total:   total,
			}
			return
		case types.CallbackFinished:
			c.blobs[instance] = copyBlobTracker{
				Current: cur,
				Total:   total,
			}
			return
		case types.CallbackSkipped:
			c.blobs[instance] = copyBlobTracker{
				Current: total,
				Total:   total,
			}
			return
		}
	}
}
