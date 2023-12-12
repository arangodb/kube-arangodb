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

package v1alpha1

import (
	batch "k8s.io/api/batch/v1"

	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ArangoMLCronJobSpec struct {
	// +doc/type: batch.CronJobSpec
	// +doc/link: Kubernetes Documentation|https://godoc.org/k8s.io/api/batch/v1#CronJobSpec
	*batch.CronJobSpec `json:",inline"`
}

func (a *ArangoMLCronJobSpec) Validate() error {
	if a == nil {
		return errors.Newf("Spec is not defined")
	}

	var err []error
	if a.CronJobSpec == nil {
		err = append(err, shared.PrefixResourceErrors("spec", errors.Newf("CronJobSpec is not defined")))
	}

	return shared.WithErrors(err...)
}
