//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package openid

import meta "k8s.io/apimachinery/pkg/apis/meta/v1"

type Session struct {
	Key string `json:"_key"`

	IDToken string `json:"idToken"`

	ExpiresAt        meta.Time `json:"expiresAt"`
	ExpiresAtSeconds int64     `json:"expiresAtSeconds"`

	Username string `json:"username"`
}

func (s *Session) SetKey(k string) {
	s.Key = k
}

func (s *Session) GetKey() string {
	if s == nil {
		return ""
	}

	return s.Key
}
