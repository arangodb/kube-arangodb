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
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// LoadKeyFile loads a SSL keyfile formatted for the arangod server.
func LoadKeyFile(keyFile string) (tls.Certificate, error) {
	raw, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return tls.Certificate{}, maskAny(err)
	}

	result := tls.Certificate{}
	for {
		var derBlock *pem.Block
		derBlock, raw = pem.Decode(raw)
		if derBlock == nil {
			break
		}
		if derBlock.Type == "CERTIFICATE" {
			result.Certificate = append(result.Certificate, derBlock.Bytes)
		} else if derBlock.Type == "PRIVATE KEY" || strings.HasSuffix(derBlock.Type, " PRIVATE KEY") {
			if result.PrivateKey == nil {
				result.PrivateKey, err = parsePrivateKey(derBlock.Bytes)
				if err != nil {
					return tls.Certificate{}, maskAny(err)
				}
			}
		}
	}

	if len(result.Certificate) == 0 {
		return tls.Certificate{}, maskAny(fmt.Errorf("No certificates found in %s", keyFile))
	}
	if result.PrivateKey == nil {
		return tls.Certificate{}, maskAny(fmt.Errorf("No private key found in %s", keyFile))
	}

	return result, nil
}

// ExtractCACertificateFromKeyFile loads a SSL keyfile formatted for the arangod server and
// extracts the CA certificate(s) from it (if any).
func ExtractCACertificateFromKeyFile(keyFile string) (string, error) {
	raw, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return "", maskAny(err)
	}

	buf := &bytes.Buffer{}
	certificatesFound := 0
	for {
		var derBlock *pem.Block
		derBlock, raw = pem.Decode(raw)
		if derBlock == nil {
			break
		}
		if derBlock.Type == "CERTIFICATE" {
			certificatesFound++
			c, err := x509.ParseCertificate(derBlock.Bytes)
			if err != nil {
				return "", maskAny(err)
			}
			if c.IsCA {
				pem.Encode(buf, derBlock)
			}
		}
	}

	certPem := buf.String()
	if certificatesFound == 0 {
		return "", maskAny(fmt.Errorf("No certificates found in %s", keyFile))
	}
	return certPem, nil
}

// SaveKeyFile creates a keyfile with given certificate & key data
func SaveKeyFile(cert, key string, filename string) error {
	folder := filepath.Dir(filename)
	if folder != "" {
		os.MkdirAll(folder, 0755)
	}
	content := strings.TrimSpace(cert) + "\n" + strings.TrimSpace(key)
	if err := ioutil.WriteFile(filename, []byte(content), 0600); err != nil {
		return maskAny(err)
	}
	return nil
}

// Attempt to parse the given private key DER block. OpenSSL 0.9.8 generates
// PKCS#1 private keys by default, while OpenSSL 1.0.0 generates PKCS#8 keys.
// OpenSSL ecparam generates SEC1 EC private keys for ECDSA. We try all three.
func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, maskAny(errors.New("tls: found unknown private key type in PKCS#8 wrapping"))
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, maskAny(errors.New("tls: failed to parse private key"))
}

// EncodeToString encodes the given certification information into
// 2 strings. The first containing all certificates (PEM encoded),
// the second containing the private key (PEM encoded).
func EncodeToString(c tls.Certificate) (cert, key string) {
	// Encode certificates
	buf := &bytes.Buffer{}
	for _, x := range c.Certificate {
		pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: x})
	}
	certPem := buf.String()

	// Private key
	buf = &bytes.Buffer{}
	pem.Encode(buf, pemBlockForKey(c.PrivateKey))
	privPem := buf.String()

	return certPem, privPem
}
