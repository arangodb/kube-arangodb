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
	"io"

	"github.com/regclient/regclient/types/descriptor"
	"github.com/regclient/regclient/types/manifest"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Error struct {
	errs []error
}

func (e Error) UpdateProto(mod util.ModR[Proto]) Builder {
	return e
}

func (e Error) WithChart(name, version string, data []byte) Builder {
	return e
}

func (e Error) Manifests() map[string]string {
	return nil
}

func (e Error) HasError() bool {
	return true
}

func (e Error) WithError(err error) Builder {
	errs := make([]error, len(e.errs)+1)
	copy(errs, e.errs)
	errs[len(e.errs)] = err
	return Error{errs}
}

func (e Error) WithManifest(manifest manifest.Manifest) Builder {
	return e
}

func (e Error) WithBlob(desc descriptor.Descriptor, blob io.ReadCloser) Builder {
	return e
}

func (e Error) Done() error {
	return shared.WithErrors(e.errs...)
}
