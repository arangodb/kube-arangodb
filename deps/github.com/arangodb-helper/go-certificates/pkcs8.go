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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"

	"github.com/pkg/errors"
)

// Unecrypted PKCS8
var (
	oidPKCS5PBKDF2 = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 12}
	oidPBES2       = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 5, 13}
	oidAES256CBC   = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 1, 42}
)

// Copy from crypto/x509
var (
	oidPublicKeyRSA   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	oidPublicKeyECDSA = asn1.ObjectIdentifier{1, 2, 840, 10045, 2, 1}
)

// Copy from crypto/x509
var (
	oidNamedCurveP224 = asn1.ObjectIdentifier{1, 3, 132, 0, 33}
	oidNamedCurveP256 = asn1.ObjectIdentifier{1, 2, 840, 10045, 3, 1, 7}
	oidNamedCurveP384 = asn1.ObjectIdentifier{1, 3, 132, 0, 34}
	oidNamedCurveP521 = asn1.ObjectIdentifier{1, 3, 132, 0, 35}
)

// Copy from crypto/x509
func oidFromNamedCurve(curve elliptic.Curve) (asn1.ObjectIdentifier, bool) {
	switch curve {
	case elliptic.P224():
		return oidNamedCurveP224, true
	case elliptic.P256():
		return oidNamedCurveP256, true
	case elliptic.P384():
		return oidNamedCurveP384, true
	case elliptic.P521():
		return oidNamedCurveP521, true
	}

	return nil, false
}

type privateKeyInfo struct {
	Version             int
	PrivateKeyAlgorithm []asn1.ObjectIdentifier
	PrivateKey          []byte
}

// Encrypted PKCS8
/*type pbkdf2Params struct {
	Salt           []byte
	IterationCount int
}

type pbkdf2Algorithms struct {
	IDPBKDF2     asn1.ObjectIdentifier
	PBKDF2Params pbkdf2Params
}

type pbkdf2Encs struct {
	EncryAlgo asn1.ObjectIdentifier
	IV        []byte
}

type pbes2Params struct {
	KeyDerivationFunc pbkdf2Algorithms
	EncryptionScheme  pbkdf2Encs
}

type pbes2Algorithms struct {
	IDPBES2     asn1.ObjectIdentifier
	PBES2Params pbes2Params
}

type encryptedPrivateKeyInfo struct {
	EncryptionAlgorithm pbes2Algorithms
	EncryptedData       []byte
}*/

func convertPrivateKeyToPKCS8(priv interface{}) (der []byte, err error) {
	var rb []byte
	var pkey privateKeyInfo

	switch priv := priv.(type) {
	case *ecdsa.PrivateKey:
		eckey, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			return nil, maskAny(err)
		}

		oidNamedCurve, ok := oidFromNamedCurve(priv.Curve)
		if !ok {
			return nil, maskAny(errors.New("pkcs8: unknown elliptic curve"))
		}

		// Per RFC5958, if publicKey is present, then version is set to v2(1) else version is set to v1(0).
		// But openssl set to v1 even publicKey is present
		pkey.Version = 0
		pkey.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 2)
		pkey.PrivateKeyAlgorithm[0] = oidPublicKeyECDSA
		pkey.PrivateKeyAlgorithm[1] = oidNamedCurve
		pkey.PrivateKey = eckey
	case *rsa.PrivateKey:

		// Per RFC5958, if publicKey is present, then version is set to v2(1) else version is set to v1(0).
		// But openssl set to v1 even publicKey is present
		pkey.Version = 0
		pkey.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 1)
		pkey.PrivateKeyAlgorithm[0] = oidPublicKeyRSA
		pkey.PrivateKey = x509.MarshalPKCS1PrivateKey(priv)
	}

	rb, err = asn1.Marshal(pkey)
	if err != nil {
		return nil, maskAny(err)
	}

	return rb, nil
}
