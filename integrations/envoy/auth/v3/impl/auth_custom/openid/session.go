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

import (
	"time"

	"golang.org/x/oauth2"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbImplEnvoyAuthV3Shared "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3/shared"
)

type Session struct {
	Token     oauth2.Token `json:"token"`
	ExpiresAt meta.Time    `json:"expiresAt"`

	Username string `json:"username"`
}

func (s *Session) Expires() time.Time {
	if s == nil {
		return time.Time{}
	}

	return s.ExpiresAt.Time
}

func (s *Session) AsResponse() *pbImplEnvoyAuthV3Shared.ResponseAuth {
	if s == nil {
		return nil
	}

	return &pbImplEnvoyAuthV3Shared.ResponseAuth{
		User: s.Username,
	}
}
