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
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"
	"sigs.k8s.io/yaml"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/closer"
)

type Builder interface {
	Done() error

	WithBlob(desc descriptor.Descriptor, blob io.ReadCloser) Builder
	WithManifest(manifest manifest.Manifest) Builder
	WithChart(name, version string, data []byte) Builder

	Manifests() map[string]string

	UpdateProto(mod util.ModR[Proto]) Builder

	WithError(err error) Builder
	HasError() bool
}

func NewBuilder(path string) Builder {
	c := closer.NewMultiCloser()

	out, err := os.Create(path)
	if err != nil {
		return Error{errs: []error{err}}
	}

	c = c.With(out)

	tw := zip.NewWriter(out)

	c = c.With(tw)

	return &builder{
		closer:    c,
		zip:       tw,
		bloobs:    map[string]bool{},
		manifests: map[string]string{},
	}
}

type builder struct {
	lock sync.Mutex

	zip *zip.Writer

	closer closer.Close

	proto Proto

	bloobs    map[string]bool
	manifests map[string]string
}

func (b *builder) UpdateProto(mod util.ModR[Proto]) Builder {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.proto = mod(b.proto)

	return b
}

func (b *builder) Manifests() map[string]string {
	b.lock.Lock()
	defer b.lock.Unlock()

	r := make(map[string]string, len(b.manifests))

	for k, v := range b.manifests {
		r[k] = v
	}

	return r
}

func (b *builder) HasError() bool {
	return false
}

func (b *builder) WithError(err error) Builder {
	return Error{errs: []error{err}}
}

func (b *builder) WithChart(name, version string, data []byte) Builder {
	b.lock.Lock()
	defer b.lock.Unlock()

	out, err := b.zip.Create(fmt.Sprintf("chart/%s-%s.tgz", name, version))
	if err != nil {
		return b.WithError(err)
	}

	if _, err := out.Write(data); err != nil {
		return b.WithError(err)
	}

	return b
}

func (b *builder) WithManifest(manifest manifest.Manifest) Builder {
	b.lock.Lock()
	defer b.lock.Unlock()

	if _, ok := b.manifests[manifest.GetRef().CommonName()]; ok {
		return b
	}

	b.manifests[manifest.GetRef().CommonName()] = manifest.GetDescriptor().Digest.Hex()

	data, err := manifest.MarshalJSON()
	if err != nil {
		return b.WithError(err)
	}

	out, err := b.zip.Create(fmt.Sprintf("manifest/%s", manifest.GetDescriptor().Digest.Hex()))
	if err != nil {
		return b.WithError(err)
	}

	if _, err := out.Write(data); err != nil {
		return b.WithError(err)
	}

	return b
}

func (b *builder) WithBlob(desc descriptor.Descriptor, blob io.ReadCloser) Builder {
	b.lock.Lock()
	defer b.lock.Unlock()

	if _, ok := b.bloobs[desc.Digest.Hex()]; ok {
		if err := blob.Close(); err != nil {
			return b.WithError(err)
		}
		return b
	}

	b.bloobs[desc.Digest.Hex()] = true

	out, err := b.zip.Create(fmt.Sprintf("bloob/%s", desc.Digest.Hex()))
	if err != nil {
		return b.WithError(err)
	}

	if _, err := io.Copy(out, blob); err != nil {
		return b.WithError(err)
	}

	if err := blob.Close(); err != nil {
		return b.WithError(err)
	}

	return b
}

func (b *builder) Done() error {
	b.lock.Lock()
	defer b.lock.Unlock()

	out, err := b.zip.Create("proto.yaml")
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(b.proto)
	if err != nil {
		return err
	}

	if _, err := out.Write(data); err != nil {
		return err
	}

	return b.closer.Close()
}
