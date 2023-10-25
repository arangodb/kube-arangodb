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

package reconcile

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
)

func WithArangoMemberStatusUpdate(ctx context.Context, client inspector.ArangoMemberUpdateInterface, name string, f inspector.ArangoMemberUpdateFunc) error {
	return inspector.WithArangoMemberStatusUpdate(ctx, client, name, func(in *api.ArangoMember) (bool, error) {
		if changed, err := f(in); err != nil {
			return false, err
		} else if !changed {
			return false, nil
		}

		// Update Status

		postArangoMemberStatusUpdate(in)

		return true, nil
	})
}

func postArangoMemberStatusUpdate(in *api.ArangoMember) {
	in.Status.LastUpdateTime = meta.Now()

	if in.Status.Conditions.IsTrue(api.ConditionTypeReady) {
		in.Status.Message = "Member is Ready"
	} else {
		in.Status.Message = "Member is Not Ready"
	}
}
