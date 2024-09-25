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

package v1

import (
	pbSchedulerV1 "github.com/arangodb/kube-arangodb/integrations/scheduler/v1/definition"
	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

func ExtractStatusMetadata(mt schedulerApi.ArangoSchedulerStatusMetadata) *pbSchedulerV1.StatusMetadata {
	var r pbSchedulerV1.StatusMetadata

	r.Profiles = mt.Profiles

	if obj := mt.Object; obj == nil {
		r.Created = false
	} else {
		r.Created = true
		r.Checksum = util.NewType(obj.GetChecksum())
		r.Uid = util.NewType(obj.GetChecksum())
	}

	return &r
}
