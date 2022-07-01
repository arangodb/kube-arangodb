//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"

	core "k8s.io/api/core/v1"

	"github.com/arangodb-helper/go-certificates"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	inspectorInterface "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
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

func (r *Resources) GetCertsFromData(caPem []byte) Certificates {
	log := r.log.Str("section", "tls")
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
			log.Err(err).Error("Unable to parse certificate")
			continue
		}

		certs = append(certs, cert)
	}

	return certs
}

func (r *Resources) GetCertsFromSecret(secret *core.Secret) Certificates {
	caPem, exists := secret.Data[core.ServiceAccountRootCAKey]
	if !exists {
		return nil
	}

	return r.GetCertsFromData(caPem)
}

func (r *Resources) GetKeyCertFromCache(cachedStatus inspectorInterface.Inspector, spec api.DeploymentSpec, certName, keyName string) (Certificates, interface{}, error) {
	caSecret, exists := cachedStatus.Secret().V1().GetSimple(spec.TLS.GetCASecretName())
	if !exists {
		return nil, nil, errors.Newf("CA Secret does not exists")
	}

	return GetKeyCertFromSecret(caSecret, keyName, certName)
}

func GetKeyCertFromSecret(secret *core.Secret, certName, keyName string) (Certificates, interface{}, error) {
	ca, exists := secret.Data[certName]
	if !exists {
		return nil, nil, errors.Newf("Key %s missing in secret", certName)
	}

	key, exists := secret.Data[keyName]
	if !exists {
		return nil, nil, errors.Newf("Key %s missing in secret", keyName)
	}

	cert, keys, err := certificates.LoadFromPEM(string(ca), string(key))
	if err != nil {
		return nil, nil, err
	}

	return cert, keys, nil
}
