//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package resources

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"

	"github.com/arangodb-helper/go-certificates"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	core "k8s.io/api/core/v1"
)

type Certificates []*x509.Certificate

func (c Certificates) Contains(cert *x509.Certificate) bool {
	for _, localCert := range c {
		if !localCert.Equal(cert) {
			return false
		}
	}

	return true
}

func (c Certificates) ContainsAll(certs Certificates) bool {
	if len(certs) == 0 {
		return true
	}

	for _, cert := range certs {
		if !c.Contains(cert) {
			return false
		}
	}

	return true
}

func (c Certificates) ToPem() ([]byte, error) {
	bytes := bytes.NewBuffer([]byte{})

	for _, cert := range c {
		if err := pem.Encode(bytes, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}); err != nil {
			return nil, err
		}
	}

	return bytes.Bytes(), nil
}

func (c Certificates) AsCertPool() *x509.CertPool {
	cp := x509.NewCertPool()

	for _, cert := range c {
		cp.AddCert(cert)
	}

	return cp
}

func GetCertsFromData(log zerolog.Logger, caPem []byte) Certificates {
	certs := make([]*x509.Certificate, 0, 2)

	for {
		pem, rest := pem.Decode(caPem)
		if pem == nil {
			break
		}

		caPem = rest

		cert, err := x509.ParseCertificate(pem.Bytes)
		if err != nil {
			// This error should be ignored
			log.Error().Err(err).Msg("Unable to parse certificate")
			continue
		}

		certs = append(certs, cert)
	}

	return certs
}

func GetCertsFromSecret(log zerolog.Logger, secret *core.Secret) Certificates {
	caPem, exists := secret.Data[core.ServiceAccountRootCAKey]
	if !exists {
		return nil
	}

	return GetCertsFromData(log, caPem)
}

func GetKeyCertFromCache(log zerolog.Logger, cachedStatus inspector.Inspector, spec api.DeploymentSpec, certName, keyName string) (Certificates, interface{}, error) {
	caSecret, exists := cachedStatus.Secret(spec.TLS.GetCASecretName())
	if !exists {
		return nil, nil, errors.Errorf("CA Secret does not exists")
	}

	return GetKeyCertFromSecret(log, caSecret, keyName, certName)
}

func GetKeyCertFromSecret(log zerolog.Logger, secret *core.Secret, certName, keyName string) (Certificates, interface{}, error) {
	ca, exists := secret.Data[certName]
	if !exists {
		return nil, nil, errors.Errorf("Key %s missing in secret", certName)
	}

	key, exists := secret.Data[keyName]
	if !exists {
		return nil, nil, errors.Errorf("Key %s missing in secret", keyName)
	}

	cert, keys, err := certificates.LoadFromPEM(string(ca), string(key))
	if err != nil {
		return nil, nil, err
	}

	return cert, keys, nil
}
