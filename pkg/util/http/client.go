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
	"crypto/x509"
	goHttp "net/http"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func NewHTTPClient(mods ...util.Mod[goHttp.Client]) HTTPClient {
	var c goHttp.Client

	util.ApplyMods(&c, mods...)

	return &c
}

type HTTPClient Client[*goHttp.Request, *goHttp.Response]

type Client[Req, Resp any] interface {
	Do(req Req) (Resp, error)
}

func RoundTripper(mods ...util.Mod[goHttp.Transport]) goHttp.RoundTripper {
	df := append([]util.Mod[goHttp.Transport]{
		configuration.DefaultTransport,
	}, mods...)

	return Transport(df...)
}

func RoundTripperWithShortTransport(mods ...util.Mod[goHttp.Transport]) goHttp.RoundTripper {
	df := append([]util.Mod[goHttp.Transport]{
		configuration.ShortTransport,
	}, mods...)

	return Transport(df...)
}

func DoNotFollowRedirects(client *goHttp.Client) {
	client.CheckRedirect = func(req *goHttp.Request, via []*goHttp.Request) error {
		return goHttp.ErrUseLastResponse // Do not wrap, standard library will not understand
	}
}

func Insecure(in *tls.Config) {
	in.InsecureSkipVerify = true
}

func WithRootCA(ca *x509.CertPool) util.Mod[tls.Config] {
	return func(in *tls.Config) {
		in.RootCAs = ca
	}
}
