//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package resources

import (
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/pkg/errors"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
)

// ValidateLicenseKeySecret checks if the licens key secret exists and is valid
func (r *Resources) ValidateLicenseKeySecret(cachedStatus inspector.Inspector) error {
	spec := r.context.GetSpec().License

	if spec.HasSecretName() {
		secretName := spec.GetSecretName()

		s, exists := cachedStatus.Secret(secretName)

		if !exists {
			return errors.Errorf("License secret %s does not exist", s)
		}

		if _, ok := s.Data[constants.SecretKeyToken]; !ok {
			return errors.Errorf("Invalid secret format")
		}
	}

	return nil
}
