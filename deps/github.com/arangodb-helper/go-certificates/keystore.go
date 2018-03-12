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
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	keystore "github.com/pavel-v-chernykh/keystore-go"

	"github.com/pkg/errors"
)

// CreateKeystore creates a java keystore containing the given certificate,
// private key & ca certificate(s).
func CreateKeystore(cert, key, caCert string, alias string, keystorePassword []byte) ([]byte, error) {
	ks := make(keystore.KeyStore)

	// Decode CA cert
	ksCACerts, err := decodeCACertificates(caCert)
	if err != nil {
		return nil, maskAny(errors.Wrap(err, "Failed to decode CA certificates"))
	}
	for alias, ksCACert := range ksCACerts {
		ks[alias] = &keystore.TrustedCertificateEntry{
			Entry:       keystore.Entry{CreationDate: time.Now()},
			Certificate: ksCACert,
		}
	}

	// Decode certificate
	ksCerts, err := decodeCertificates(cert)
	if err != nil {
		return nil, maskAny(errors.Wrap(err, "Failed to decode certificate"))
	}

	// Decode private key
	pk, err := decodePrivateKey(key)
	if err != nil {
		return nil, maskAny(errors.Wrap(err, "Failed to decode private key"))
	}
	encPK, err := convertPrivateKeyToPKCS8(pk)
	if err != nil {
		return nil, maskAny(errors.Wrap(err, "Failed to encode private key"))
	}
	ks[alias] = &keystore.PrivateKeyEntry{
		Entry:     keystore.Entry{CreationDate: time.Now()},
		PrivKey:   encPK,
		CertChain: ksCerts,
	}

	// Encode keystore
	buf := &bytes.Buffer{}
	if err := keystore.Encode(buf, ks, keystorePassword); err != nil {
		return nil, maskAny(errors.Wrap(err, "Failed to encode keystore"))
	}

	return buf.Bytes(), nil
}

// decodeCACertificates takes a PEM encoded string and decodes all certificates
// in it into a map of alias+certificate pairs.
func decodeCACertificates(pemContent string) (map[string]keystore.Certificate, error) {
	ksCerts, err := decodeCertificates(pemContent)
	if err != nil {
		return nil, maskAny(err)
	}

	result := map[string]keystore.Certificate{}
	for _, ksCert := range ksCerts {
		caCerts, err := x509.ParseCertificates(ksCert.Content)
		if err != nil {
			return nil, maskAny(err)
		}
		if len(caCerts) == 0 {
			return nil, maskAny(errors.New("Failed to parse CA certificate"))
		}

		for _, caCert := range caCerts {
			commonName := caCert.Subject.CommonName
			if commonName == "" {
				return nil, maskAny(fmt.Errorf("Missing common name in CA certificate '%s'", caCert.Subject))
			}
			alias := strings.Replace(strings.ToLower(commonName), " ", "", -1)
			result[alias] = ksCert
		}
	}
	return result, nil
}

// decodeCertificates takes a PEM encoded string and decodes it a list of
// keystore certificates.
func decodeCertificates(pemContent string) ([]keystore.Certificate, error) {
	if pemContent == "" {
		return nil, nil
	}
	blocks, err := decodePEMString(pemContent)
	if err != nil {
		return nil, maskAny(errors.Wrap(err, "Failed to decode certificates"))
	}
	var result []keystore.Certificate
	for _, b := range blocks {
		if b.Type == "CERTIFICATE" {
			result = append(result, keystore.Certificate{
				Type:    "X509",
				Content: b.Bytes,
			})
		} else {
			return nil, maskAny(fmt.Errorf("Unexpected block of type '%s' in CA certificates", b.Type))
		}
	}
	return result, nil
}

// decodePrivateKey takes a PEM encoded string and decodes its private key entry.
func decodePrivateKey(pemContent string) (interface{}, error) {
	blocks, err := decodePEMString(pemContent)
	if err != nil {
		return nil, maskAny(errors.Wrap(err, "Failed to decode private key"))
	}
	var result interface{}
	for _, b := range blocks {
		if b.Type == "PRIVATE KEY" || strings.HasSuffix(b.Type, " PRIVATE KEY") {
			if result != nil {
				return nil, maskAny(errors.New("Found multiple private keys"))
			}
			privKey, err := parsePrivateKey(b.Bytes)
			if err != nil {
				return nil, maskAny(err)
			}
			result = privKey
		} else {
			return nil, maskAny(fmt.Errorf("Unexpected block of type '%s' in CA certificates", b.Type))
		}
	}
	return result, nil
}

// decodePEMString takes a PEM encoded string and decodes it into pem blocks.
func decodePEMString(pemContent string) ([]*pem.Block, error) {
	var blocks []*pem.Block
	content := []byte(pemContent)
	for {
		b, remaining := pem.Decode(content)
		if b == nil {
			if len(blocks) > 0 {
				return blocks, nil
			}
			return nil, maskAny(errors.New("failed to decode PEM blocks"))
		}
		blocks = append(blocks, b)
		if len(remaining) == 0 {
			return blocks, nil
		}
		content = remaining
	}
}
