//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

//go:build go1.20

package reconcile

import (
	"crypto/tls"
	"crypto/x509"
)

func isCertificateVerificationError(err error) bool {
	switch err.(type) {
	case x509.UnknownAuthorityError, x509.CertificateInvalidError, *tls.CertificateVerificationError:
		return true
	}

	return false
}
