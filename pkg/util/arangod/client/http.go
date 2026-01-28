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

package client

import (
	"context"
	goHttp "net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type HTTPClient interface {
	GetClient(ctx context.Context) (goHttp.RoundTripper, error)
}

type HTTPClientFunc func(ctx context.Context) (goHttp.RoundTripper, error)

func (f HTTPClientFunc) GetClient(ctx context.Context) (goHttp.RoundTripper, error) {
	return f(ctx)
}

func HTTPClientFactory(mods ...util.Mod[goHttp.Transport]) HTTPClient {
	return HTTPClientFunc(func(ctx context.Context) (goHttp.RoundTripper, error) {
		var c goHttp.Transport

		util.ApplyMods(&c, mods...)

		return &c, nil
	})
}
