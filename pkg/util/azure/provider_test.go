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

package azure

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// certAndKeyPEM returns a fresh self-signed certificate and its PKCS#8 private key as separate PEM
// blocks, matching how a native kubernetes.io/tls Secret exposes tls.crt and tls.key.
func certAndKeyPEM(t *testing.T) (cert, key []byte) {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "arangodb-test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	require.NoError(t, err)

	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err)

	cert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	key = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return cert, key
}

func Test_Provider_Certificate(t *testing.T) {
	cert, key := certAndKeyPEM(t)
	bundle := append(append([]byte{}, cert...), key...)

	t.Run("inline bundle (cert+key together)", func(t *testing.T) {
		cred, err := Provider{
			Type:        ProviderTypeCertificate,
			TenantID:    "00000000-0000-0000-0000-000000000000",
			Certificate: ProviderCertificate{ClientID: "11111111-1111-1111-1111-111111111111", Certificate: string(bundle)},
		}.GetCredentials()
		require.NoError(t, err)
		require.NotNil(t, cred)
	})

	t.Run("inline separate certificate and key (tls-style)", func(t *testing.T) {
		cred, err := Provider{
			Type:     ProviderTypeCertificate,
			TenantID: "00000000-0000-0000-0000-000000000000",
			Certificate: ProviderCertificate{
				ClientID:    "11111111-1111-1111-1111-111111111111",
				Certificate: string(cert),
				Key:         string(key),
			},
		}.GetCredentials()
		require.NoError(t, err)
		require.NotNil(t, cred)
	})

	t.Run("from a native TLS secret's files (tls.crt/tls.key)", func(t *testing.T) {
		dir := t.TempDir()
		crtFile := filepath.Join(dir, "tls.crt")
		keyFile := filepath.Join(dir, "tls.key")
		idFile := filepath.Join(dir, "clientId")
		require.NoError(t, os.WriteFile(crtFile, cert, 0600))
		require.NoError(t, os.WriteFile(keyFile, key, 0600))
		require.NoError(t, os.WriteFile(idFile, []byte("11111111-1111-1111-1111-111111111111"), 0600))

		cred, err := Provider{
			Type:     ProviderTypeCertificate,
			TenantID: "00000000-0000-0000-0000-000000000000",
			Certificate: ProviderCertificate{
				ClientIDFile:    idFile,
				CertificateFile: crtFile,
				KeyFile:         keyFile,
			},
		}.GetCredentials()
		require.NoError(t, err)
		require.NotNil(t, cred)
	})

	t.Run("missing certificate is an error", func(t *testing.T) {
		_, err := Provider{
			Type:        ProviderTypeCertificate,
			TenantID:    "t",
			Certificate: ProviderCertificate{ClientID: "c"},
		}.GetCredentials()
		require.Error(t, err)
	})

	t.Run("invalid certificate is an error", func(t *testing.T) {
		_, err := Provider{
			Type:        ProviderTypeCertificate,
			TenantID:    "t",
			Certificate: ProviderCertificate{ClientID: "c", Certificate: "not a certificate"},
		}.GetCredentials()
		require.Error(t, err)
	})
}

func Test_Provider_Secret(t *testing.T) {
	t.Run("inline", func(t *testing.T) {
		cred, err := Provider{
			Type:     ProviderTypeSecret,
			TenantID: "00000000-0000-0000-0000-000000000000",
			Secret:   ProviderSecret{ClientID: "id", ClientSecret: "secret"},
		}.GetCredentials()
		require.NoError(t, err)
		require.NotNil(t, cred)
	})

	t.Run("missing client id is an error", func(t *testing.T) {
		_, err := Provider{
			Type:   ProviderTypeSecret,
			Secret: ProviderSecret{ClientSecret: "secret"},
		}.GetCredentials()
		require.Error(t, err)
	})
}

func Test_Provider_UnknownType(t *testing.T) {
	_, err := Provider{Type: "managed"}.GetCredentials()
	require.Error(t, err)
}
