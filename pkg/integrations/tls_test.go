//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/tgrpc"
)

func Test_TLSCases(t *testing.T) {
	directory := t.TempDir()

	ca1 := path.Join(directory, "CA1.keyfile")
	ca1Pem := path.Join(directory, "CA1.pem")
	server1 := path.Join(directory, "server1.keyfile")

	ca2 := path.Join(directory, "CA2.keyfile")
	ca2Pem := path.Join(directory, "CA2.pem")
	server2 := path.Join(directory, "server2.keyfile")

	t.Run("Arrange CA 1", func(t *testing.T) {
		caCert, caKey, err := ktls.CreateTLSCACertificate("Test Root Certificate")

		require.NoError(t, err)

		require.NoError(t, os.WriteFile(ca1, []byte(ktls.AsKeyfile(caCert, caKey)), 0644))

		require.NoError(t, os.WriteFile(ca1Pem, []byte(caCert), 0644))

		serverCert, serverKey, err := ktls.CreateTLSServerCertificate(caCert, caKey, ktls.KeyfileInput{
			AltNames: []string{
				"127.0.0.1",
			},
			Email: nil,
		})

		require.NoError(t, err)

		require.NoError(t, os.WriteFile(server1, []byte(ktls.AsKeyfile(serverCert, serverKey)), 0644))
	})

	t.Run("Arrange CA 2", func(t *testing.T) {
		caCert, caKey, err := ktls.CreateTLSCACertificate("Test Root Certificate")

		require.NoError(t, err)

		require.NoError(t, os.WriteFile(ca2, []byte(ktls.AsKeyfile(caCert, caKey)), 0644))

		require.NoError(t, os.WriteFile(ca2Pem, []byte(caCert), 0644))

		serverCert, serverKey, err := ktls.CreateTLSServerCertificate(caCert, caKey, ktls.KeyfileInput{
			AltNames: []string{
				"127.0.0.1",
			},
			Email: nil,
		})

		require.NoError(t, err)

		require.NoError(t, os.WriteFile(server2, []byte(ktls.AsKeyfile(serverCert, serverKey)), 0644))
	})

	c, health, internal, external := startService(t,
		"--health.tls.keyfile=",
		fmt.Sprintf("--services.tls.keyfile=%s", server1),
		fmt.Sprintf("--services.external.tls.keyfile=%s", server2),
	)
	defer c.Require(t)

	t.Run("Without TLS", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--tls.enabled=false",
				"client",
				"health",
				"v1"))
		})
		t.Run("internal", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--tls.enabled=false",
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"error reading server preface: EOF\"")
		})
		t.Run("external", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--tls.enabled=false",
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"error reading server preface: EOF\"")
		})
	})

	t.Run("With TLS Fallback", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--tls.enabled=true",
				"--tls.fallback=true",
				"--tls.insecure=true",
				"client",
				"health",
				"v1"))
		})
		t.Run("internal", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--tls.enabled=true",
				"--tls.insecure=true",
				"--tls.fallback=true",
				"client",
				"health",
				"v1"))
		})
		t.Run("external", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--tls.enabled=true",
				"--tls.insecure=true",
				"--tls.fallback=true",
				"client",
				"health",
				"v1"))
		})
	})

	t.Run("With TLS - wrong CA", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--tls.enabled=true",
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"transport: authentication handshake failed: tls: first record does not look like a TLS handshake\"")
		})
		t.Run("internal", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--tls.enabled=true",
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"transport: authentication handshake failed: tls: failed to verify certificate: x509: certificate signed by unknown authority\"")
		})
		t.Run("external", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--tls.enabled=true",
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"transport: authentication handshake failed: tls: failed to verify certificate: x509: certificate signed by unknown authority\"")
		})
	})

	t.Run("With TLS - valid CA1", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--tls.enabled=true",
				fmt.Sprintf("--tls.ca=%s", ca1Pem),
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"transport: authentication handshake failed: tls: first record does not look like a TLS handshake\"")
		})
		t.Run("internal", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--tls.enabled=true",
				fmt.Sprintf("--tls.ca=%s", ca1Pem),
				"client",
				"health",
				"v1"))
		})
		t.Run("external", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--tls.enabled=true",
				fmt.Sprintf("--tls.ca=%s", ca1Pem),
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"transport: authentication handshake failed: tls: failed to verify certificate: x509: certificate signed by unknown authority (possibly because of \\\"x509: ECDSA verification failure\\\" while trying to verify candidate authority certificate \\\"Test Root Certificate\\\")\"")
		})
	})

	t.Run("With TLS - insecure", func(t *testing.T) {
		t.Run("health", func(t *testing.T) {
			tgrpc.AsGRPCError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", health),
				"--tls.enabled=true",
				"--tls.insecure=true",
				"client",
				"health",
				"v1")).Code(t, codes.Unavailable).Errorf(t, "connection error: desc = \"transport: authentication handshake failed: tls: first record does not look like a TLS handshake\"")
		})
		t.Run("internal", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", internal),
				"--tls.enabled=true",
				"--tls.insecure=true",
				"client",
				"health",
				"v1"))
		})
		t.Run("external", func(t *testing.T) {
			require.NoError(t, executeSync(t, shutdown.Context(),
				fmt.Sprintf("--address=127.0.0.1:%d", external),
				"--tls.enabled=true",
				"--tls.insecure=true",
				"client",
				"health",
				"v1"))
		})
	})
}
