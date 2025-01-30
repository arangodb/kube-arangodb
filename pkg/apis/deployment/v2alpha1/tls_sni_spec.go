//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package v2alpha1

import (
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

// TLSSNISpec holds TLS SNI additional certificates
type TLSSNISpec struct {
	// The mapping of the Server Name Indication options.
	// +doc/type: map[string][]string
	// +doc/example: mapping:
	// +doc/example:   secret:
	// +doc/example:     - domain.example.com
	// +doc/link: Server Name Indication|https://docs.arangodb.com/stable/components/arangodb-server/options/#--sslserver-name-indication
	Mapping map[string][]string `json:"mapping,omitempty"`
}

func (s TLSSNISpec) Validate() error {
	mapped := map[string]interface{}{}

	for key, values := range s.Mapping {
		if err := sharedApi.IsValidName(key); err != nil {
			return err
		}

		for _, value := range values {
			if _, exists := mapped[value]; exists {
				return errors.Errorf("sni for host %s is already defined", value)
			}

			// Mark value as existing
			mapped[value] = nil

			if err := sharedApi.IsValidDomain(value); err != nil {
				return err
			}
		}
	}

	return nil
}

// SetDefaultsFrom fills unspecified fields with a value from given source spec.
func (s *TLSSNISpec) SetDefaultsFrom(source *TLSSNISpec) {
	if source == nil {
		return
	}
}
