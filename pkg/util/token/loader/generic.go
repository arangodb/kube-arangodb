//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package loader

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	utilToken "github.com/arangodb/kube-arangodb/pkg/util/token"
)

func LoadSecretsFromData(data map[string][]byte) (utilToken.Secret, utilToken.Secrets) {
	if len(data) == 0 {
		return utilToken.EmptySecret(), nil
	}

	var active = utilToken.EmptySecret()

	if r, found := data[utilConstants.ActiveJWTKey]; found {
		active = utilToken.NewSecret(r)
	}

	passive := utilToken.Secrets(util.FormatList(util.FilterList(util.Extract(data), func(k util.KV[string, []byte]) bool {
		return k.K != utilConstants.ActiveJWTKey
	}), func(a util.KV[string, []byte]) utilToken.Secret {
		return utilToken.NewSecret(a.V)
	}))

	return active, passive
}
