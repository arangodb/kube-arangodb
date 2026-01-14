//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package http

import (
	"crypto/tls"
	goHttp "net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func WithTransport(mods ...util.Mod[goHttp.Transport]) util.Mod[goHttp.Client] {
	return func(in *goHttp.Client) {
		in.Transport = Transport(mods...)
	}
}

func Transport(mods ...util.Mod[goHttp.Transport]) goHttp.RoundTripper {
	var c goHttp.Transport

	util.ApplyMods[goHttp.Transport](&c, mods...)

	return &c
}

func WithTransportTLS(mods ...util.Mod[tls.Config]) util.Mod[goHttp.Transport] {
	return func(in *goHttp.Transport) {
		if in.TLSClientConfig == nil {
			in.TLSClientConfig = &tls.Config{}
		}

		util.ApplyMods(in.TLSClientConfig, mods...)
	}
}
