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

package k8s

import (
	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func CoreSecret(in *core.Secret) *core.Secret {
	return FilterP(in, func(in *core.Secret) *core.Secret {
		return &core.Secret{
			ObjectMeta: ObjectMetaFilter(in.ObjectMeta),
			Data:       in.Data,
		}
	})
}

func CoreSecretChecksum(in *core.Secret) (string, error) {
	return util.SHA256FromJSON(CoreSecret(in))
}
