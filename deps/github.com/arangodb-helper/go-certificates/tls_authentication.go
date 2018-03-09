//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package certificates

import (
	"crypto/tls"
)

type TLSAuthentication interface {
	CACertificate() string
	ClientCertificate() string
	ClientKey() string
}

// CreateTLSConfigFromAuthentication creates a tls.Config object from given configuration.
func CreateTLSConfigFromAuthentication(a TLSAuthentication, insecureSkipVerify bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
	}
	var err error
	tlsConfig.RootCAs, err = LoadCertPool(a.CACertificate())
	if err != nil {
		return nil, maskAny(err)
	}
	if a.ClientCertificate() != "" && a.ClientKey() != "" {
		clientCert, err := tls.X509KeyPair([]byte(a.ClientCertificate()), []byte(a.ClientKey()))
		if err != nil {
			return nil, maskAny(err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}
	return tlsConfig, nil
}
