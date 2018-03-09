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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

const (
	defaultValidFor = time.Hour * 24 * 365
	defaultRSABits  = 2048
)

type CreateCertificateOptions struct {
	Hosts          []string      // Comma-separated hostnames and IPs to generate a certificate for
	EmailAddresses []string      // List of email address to include in the certificate as alternative name
	ValidFrom      time.Time     // Creation data of the certificate
	ValidFor       time.Duration // Duration that certificate is valid for
	IsCA           bool          // Whether this cert should be its own Certificate Authority
	IsClientAuth   bool          // Whether this cert can be used for client authentication
	RSABits        int           // Size of RSA key to generate. Ignored if ECDSACurve is set
	ECDSACurve     string        // ECDSA curve to use to generate a key. Valid values are P224, P256, P384, P521
}

// CreateCertificate creates a certificate according to the given configuration.
// If ca is nil, the certificate will be self-signed, otherwise the certificate
// will be signed by the given CA certificate+key.
// The resulting certificate + private key will be PEM encoded and returned as string (cert, priv, error).
func CreateCertificate(options CreateCertificateOptions, ca *CA) (string, string, error) {
	// Create private key
	var priv interface{}
	var err error
	switch options.ECDSACurve {
	case "":
		if options.RSABits == 0 {
			options.RSABits = defaultRSABits
		}
		priv, err = rsa.GenerateKey(rand.Reader, options.RSABits)
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return "", "", maskAny(fmt.Errorf("Unknown curve '%s'", options.ECDSACurve))
	}

	notBefore := time.Now()
	if options.ValidFor == 0 {
		options.ValidFor = defaultValidFor
	}
	notAfter := notBefore.Add(options.ValidFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return "", "", maskAny(fmt.Errorf("failed to generate serial number: %v", err))
	}

	commonName := "arangosync"
	if len(options.EmailAddresses) > 0 {
		commonName = options.EmailAddresses[0]
	} else if len(options.Hosts) > 0 {
		commonName = options.Hosts[0]
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: []string{"ArangoDB"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		BasicConstraintsValid: true,
	}

	for _, h := range options.Hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}
	template.EmailAddresses = append(template.EmailAddresses, options.EmailAddresses...)

	if options.IsCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}
	if options.IsClientAuth {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageClientAuth)
	} else {
		template.ExtKeyUsage = append(template.ExtKeyUsage, x509.ExtKeyUsageServerAuth)
	}

	// Create the certificate
	var derBytes []byte
	if ca != nil {
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, ca.Certificate[0], publicKey(priv), ca.PrivateKey)
		if err != nil {
			return "", "", maskAny(fmt.Errorf("Failed to create signed certificate: %v", err))
		}
	} else {
		derBytes, err = x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
		if err != nil {
			return "", "", maskAny(fmt.Errorf("Failed to create self-signed certificate: %v", err))
		}
	}

	// Encode certificate
	// Public key
	buf := &bytes.Buffer{}
	pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if ca != nil {
		for _, c := range ca.Certificate {
			pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
		}
	}
	certPem := buf.String()

	// Private key
	buf = &bytes.Buffer{}
	pem.Encode(buf, pemBlockForKey(priv))
	privPem := buf.String()

	return certPem, privPem, nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
