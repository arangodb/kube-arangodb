//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ProviderType string

const (
	ProviderTypeSecret      ProviderType = "secret"
	ProviderTypeCertificate ProviderType = "certificate"
)

type Provider struct {
	Type ProviderType

	TenantID string

	Secret ProviderSecret

	Certificate ProviderCertificate
}

func (c Provider) GetCredentials() (azcore.TokenCredential, error) {
	switch c.Type {
	case ProviderTypeSecret:
		return c.Secret.getCredentials(c.TenantID)
	case ProviderTypeCertificate:
		return c.Certificate.getCredentials(c.TenantID)
	}

	return nil, errors.Errorf("unable to get credentials for type '%s'", c.Type)
}

// valueOrFile returns the contents of file when a path is set, otherwise the inline value. name is
// used only for the "not found" error message.
func valueOrFile(value, file, name string) (string, error) {
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	if value != "" {
		return value, nil
	}

	return "", errors.Errorf("no %s found", name)
}

type ProviderSecret struct {
	ClientID     string
	ClientIDFile string

	ClientSecret     string
	ClientSecretFile string
}

func (p ProviderSecret) getCredentials(tenantID string) (*azidentity.ClientSecretCredential, error) {
	id, err := p.GetClientID()
	if err != nil {
		return nil, err
	}

	secret, err := p.GetClientSecret()
	if err != nil {
		return nil, err
	}

	return azidentity.NewClientSecretCredential(tenantID, id, secret, nil)
}

func (p ProviderSecret) GetClientID() (string, error) {
	return valueOrFile(p.ClientID, p.ClientIDFile, "client id")
}

func (p ProviderSecret) GetClientSecret() (string, error) {
	return valueOrFile(p.ClientSecret, p.ClientSecretFile, "client secret")
}

// ProviderCertificate authenticates a service principal with a client certificate. The certificate
// and private key may be a single bundle (PEM or PKCS#12) in Certificate, or supplied separately -
// Certificate holding the certificate chain and Key the private key, as a native kubernetes.io/tls
// Secret exposes them (tls.crt / tls.key). Each value may be given inline or from a file.
type ProviderCertificate struct {
	ClientID     string
	ClientIDFile string

	Certificate     string
	CertificateFile string

	Key     string
	KeyFile string

	Password     string
	PasswordFile string
}

func (p ProviderCertificate) getCredentials(tenantID string) (*azidentity.ClientCertificateCredential, error) {
	id, err := p.GetClientID()
	if err != nil {
		return nil, err
	}

	certData, err := p.GetCertificate()
	if err != nil {
		return nil, err
	}

	// When the private key is supplied separately (tls.key), append it so ParseCertificates sees a
	// single PEM bundle with both the certificate and the key.
	keyData, err := p.GetKey()
	if err != nil {
		return nil, err
	}
	if len(keyData) > 0 {
		certData = append(append(certData, '\n'), keyData...)
	}

	password, err := p.GetPassword()
	if err != nil {
		return nil, err
	}

	certs, key, err := azidentity.ParseCertificates(certData, []byte(password))
	if err != nil {
		return nil, err
	}

	return azidentity.NewClientCertificateCredential(tenantID, id, certs, key, nil)
}

func (p ProviderCertificate) GetClientID() (string, error) {
	return valueOrFile(p.ClientID, p.ClientIDFile, "client id")
}

// GetCertificate returns the raw certificate bytes (from a file when a path is set, otherwise
// inline). Kept as bytes so PKCS#12 binary bundles read from a mounted Secret survive.
func (p ProviderCertificate) GetCertificate() ([]byte, error) {
	if f := p.CertificateFile; f != "" {
		return os.ReadFile(f)
	}

	if v := p.Certificate; v != "" {
		return []byte(v), nil
	}

	return nil, errors.New("no certificate found")
}

// GetKey returns the optional separately-provided private key bytes; an empty result is not an
// error, since Certificate may already be a bundle that includes the key.
func (p ProviderCertificate) GetKey() ([]byte, error) {
	if f := p.KeyFile; f != "" {
		return os.ReadFile(f)
	}

	if v := p.Key; v != "" {
		return []byte(v), nil
	}

	return nil, nil
}

// GetPassword returns the optional certificate password; an empty result is not an error, so an
// unencrypted PEM/PKCS#12 bundle needs no password configured.
func (p ProviderCertificate) GetPassword() (string, error) {
	if f := p.PasswordFile; f != "" {
		data, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return p.Password, nil
}
