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

package reconcile

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	goHttp "net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/arangodb-helper/go-certificates"
	adbDriverV2Connection "github.com/arangodb/go-driver/v2/connection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/crypto"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

type tlsConfig struct {
	Client, Server *tls.Config
}

func withTLSServerConfig(cfg *tls.Config, mods ...util.Mod[tls.Config]) util.Mod[tlsConfig] {
	return func(in *tlsConfig) {
		cfg := cfg.Clone()
		util.ApplyMods(cfg, mods...)
		in.Server = cfg
	}
}

func startHTTPServer(t *testing.T, mods ...util.Mod[tlsConfig]) adbDriverV2Connection.Connection {
	ctx, c := context.WithCancel(t.Context())
	t.Cleanup(c)

	var cfg tlsConfig

	util.ApplyMods(&cfg, mods...)

	s := tests.NewHTTPServer(ctx, t, func(in *goHttp.Server, p1 context.Context) error {
		in.TLSConfig = cfg.Server
		return nil
	})

	transport := operatorHTTP.RoundTripper(func(in *goHttp.Transport) {
		in.TLSClientConfig = cfg.Client
	})

	return adbDriverV2Connection.NewHttpConnection(adbDriverV2Connection.HttpConfiguration{
		Transport:          transport,
		Endpoint:           adbDriverV2Connection.NewRoundRobinEndpoints([]string{fmt.Sprintf("https://%s", s)}),
		DontFollowRedirect: true,
	})
}

func generateSelfSignedCA(t *testing.T, mods ...util.Mod[ktls.KeyfileInput]) (*x509.CertPool, tls.Certificate) {
	cert, key, err := ktls.CreateTLSCACertificate("Test Root Certificate")
	require.NoError(t, err)

	ca, err := certificates.LoadCAFromPEM(cert, key)
	require.NoError(t, err)

	var in ktls.KeyfileInput

	in.AltNames = []string{"example"}

	util.ApplyMods(&in, mods...)

	serverCert, serverKey, err := ktls.CreateTLSServerCertificate(cert, key, in)
	require.NoError(t, err)

	serverCertObject, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	require.NoError(t, err)
	return crypto.Certificates(ca.Certificate).AsCertPool(), serverCertObject
}

func generateSelfSignedConfig(t *testing.T, mods ...util.Mod[ktls.KeyfileInput]) (*tls.Config, *tls.Config) {
	ca, server := generateSelfSignedCA(t, mods...)

	return &tls.Config{ClientCAs: ca}, &tls.Config{Certificates: []tls.Certificate{server}}
}

func Test_Server_TLS(t *testing.T) {
	_, server := generateSelfSignedConfig(t)

	conn := startHTTPServer(t, withTLSServerConfig(server))

	_, err := arangod.GetRequest[any](t.Context(), conn).Do(t.Context()).HTTPResponse()
	v, ok := errors.ExtractCause[*url.Error](err)
	require.True(t, ok)
	require.True(t, isCertificateVerificationError(v.Err))
}
