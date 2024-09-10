//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package profile

import (
	"context"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func (h *handler) HandleTemplate(ctx context.Context, item operation.Item, extension *schedulerApi.ArangoProfile, status *schedulerApi.ProfileStatus) (bool, error) {
	templ := extension.Spec.Template
	c, err := templ.Checksum()
	if err != nil {
		return false, err
	}

	if t := status.Accepted; t != nil {
		if c == "" || t.Checksum != c {
			status.Accepted = nil
			return true, operator.Reconcile("Template changed")
		}

		return false, nil
	} else {
		if c != "" {
			status.Accepted = &schedulerApi.ProfileAcceptedTemplate{
				Checksum: c,
				Template: templ.DeepCopy(),
			}

			return true, operator.Reconcile("Template changed")
		}

		return false, nil
	}
}
