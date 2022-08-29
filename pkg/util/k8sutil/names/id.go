//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package names

import (
	"fmt"
	"strings"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
)

func GetArangodID(group api.ServerGroup) string {
	return GetArangodIDPredefined(group, strings.ToLower(uniuri.NewLen(8)))
}

func GetArangodIDInt(group api.ServerGroup, id int) string {
	return fmt.Sprintf("%s%s", GetArangodIDPrefix(group), fmt.Sprintf("%08d", id))
}

func GetArangodIDPredefined(group api.ServerGroup, id string) string {
	return fmt.Sprintf("%s%s", GetArangodIDPrefix(group), id)
}

// GetArangodIDPrefix returns the prefix required ID's of arangod servers
// in the given group.
func GetArangodIDPrefix(group api.ServerGroup) string {
	switch group {
	case api.ServerGroupSingle:
		return "SNGL-"
	case api.ServerGroupCoordinators:
		return "CRDN-"
	case api.ServerGroupDBServers:
		return "PRMR-"
	case api.ServerGroupAgents:
		return "AGNT-"
	default:
		return ""
	}
}
