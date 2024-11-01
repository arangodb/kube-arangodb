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

package aws

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type TLS struct {
	Insecure bool
	CAFiles  []string
}

func (s TLS) configuration() (*tls.Config, error) {
	var r tls.Config

	if s.Insecure {
		r.InsecureSkipVerify = true
	}

	if len(s.CAFiles) > 0 {
		caCertPool := x509.NewCertPool()

		for _, file := range s.CAFiles {
			caCert, err := os.ReadFile(file)
			if err != nil {
				return nil, errors.Wrapf(err, "Unable to load CA from %s", file)
			}
			caCertPool.AppendCertsFromPEM(caCert)
		}

		r.RootCAs = caCertPool
	}

	return &r, nil
}
