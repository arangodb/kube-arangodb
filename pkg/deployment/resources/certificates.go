//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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
	"crypto/x509"
	"encoding/pem"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/crypto"
)

func (r *Resources) GetCertsFromData(caPem []byte) crypto.Certificates {
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

func (r *Resources) GetCertsFromSecret(secret *core.Secret) crypto.Certificates {
	caPem, exists := secret.Data[core.ServiceAccountRootCAKey]
	if !exists {
		return nil
	}

	return r.GetCertsFromData(caPem)
}
