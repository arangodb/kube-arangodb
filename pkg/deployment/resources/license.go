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

package resources

import (
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidateLicenseKeySecret checks if the licens key secret exists and is valid
func (r *Resources) ValidateLicenseKeySecret() error {
	spec := r.context.GetSpec().License

	if spec.HasSecretName() {
		secretName := spec.GetSecretName()

		kubecli := r.context.GetKubeCli()
		ns := r.context.GetNamespace()
		s, err := kubecli.CoreV1().Secrets(ns).Get(secretName, metav1.GetOptions{})

		if err != nil {
			return err
		}

		if _, ok := s.Data[constants.SecretKeyToken]; !ok {
			return fmt.Errorf("Invalid secret format")
		}
	}

	return nil
}
