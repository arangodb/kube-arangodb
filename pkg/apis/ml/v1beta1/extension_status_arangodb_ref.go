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

package v1beta1

import sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"

type ArangoMLExtensionStatusArangoDBRef struct {
	// Secret keeps the information about ArangoDB deployment
	Secret *sharedApi.Object `json:"secret,omitempty"`
	// TLS keeps information about TLS Secret rendered from ArangoDB deployment
	TLS *sharedApi.Object `json:"tls,omitempty"`
}
