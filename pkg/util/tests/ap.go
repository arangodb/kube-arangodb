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

package tests

import (
	"testing"

	"github.com/stretchr/testify/require"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
)

func MarkArangoProfileAsReady(t *testing.T, obj *schedulerApi.ArangoProfile) {
	if err := obj.Spec.Validate(); err != nil {
		obj.Status.Conditions.Update(schedulerApi.SpecValidCondition, false, "Spec Invalid", "Spec Invalid")
		obj.Status.Conditions.Update(schedulerApi.ReadyCondition, false, "Spec Invalid", "Spec Invalid")
		return
	}
	obj.Status.Conditions.Update(schedulerApi.SpecValidCondition, true, "Spec Valid", "Spec Valid")

	checksum, err := obj.Spec.Template.Checksum()
	require.NoError(t, err)

	obj.Status.Accepted = &schedulerApi.ProfileAcceptedTemplate{
		Checksum: checksum,
		Template: obj.Spec.Template.DeepCopy(),
	}
	obj.Status.Conditions.Update(schedulerApi.ReadyCondition, true, "OK", "OK")
}
