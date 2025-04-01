//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package tls

import (
	goStrings "strings"
	"time"

	"github.com/arangodb-helper/go-certificates"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const (
	caTTL         = time.Hour * 24 * 365 * 10 // 10 year
	tlsECDSACurve = "P256"                    // This curve is the default that ArangoDB accepts and plenty strong
)

// CreateTLSCACertificate creates a CA certificate
func CreateTLSCACertificate(commonName string) (string, string, error) {
	options := certificates.CreateCertificateOptions{
		CommonName: commonName,
		ValidFrom:  time.Now(),
		ValidFor:   caTTL,
		IsCA:       true,
		ECDSACurve: tlsECDSACurve,
	}

	cert, priv, err := certificates.CreateCertificate(options, nil)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	return cert, priv, nil
}

// CreateTLSServerCertificate creates Sever Cert in PEM Format
func CreateTLSServerCertificate(caCert, caKey string, names KeyfileInput) (string, string, error) {
	ca, err := certificates.LoadCAFromPEM(caCert, caKey)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	options := certificates.CreateCertificateOptions{
		CommonName:     names.AltNames[0],
		Hosts:          names.AltNames,
		EmailAddresses: names.Email,
		ValidFrom:      time.Now(),
		ValidFor:       util.TypeOrDefault(names.TTL, api.DefaultTLSTTL.AsDuration()),
		IsCA:           false,
		ECDSACurve:     tlsECDSACurve,
	}
	cert, priv, err := certificates.CreateCertificate(options, &ca)
	if err != nil {
		return "", "", errors.WithStack(err)
	}

	return cert, priv, nil
}

// CreateTLSServerKeyfile creates Sever Cert in Keyfile Format
func CreateTLSServerKeyfile(caCert, caKey string, names KeyfileInput) (string, error) {
	cert, priv, err := CreateTLSServerCertificate(caCert, caKey, names)
	if err != nil {
		return "", err
	}

	return AsKeyfile(cert, priv), nil
}

func AsKeyfile(cert, priv string) string {
	return goStrings.TrimSpace(cert) + "\n" +
		goStrings.TrimSpace(priv)
}
