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

package token

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

func extractTokenDetails(cache Secret, t string) (*string, []string, time.Duration, error) {
	p, err := cache.Validate(t)
	if err != nil {
		return nil, nil, 0, err
	}

	var user *string
	if v, ok := p.Claims()[ClaimPreferredUsername]; ok {
		if s, ok := v.(string); ok {
			user = util.NewType(s)
		}
	}

	var duration time.Duration = -1

	claims := p.Claims()

	if v, ok := claims[ClaimEXP]; ok {
		switch o := v.(type) {
		case int64:
			duration = time.Until(time.Unix(o, 0))
		case float64:
			duration = time.Until(time.Unix(int64(o), 0))
		}
	}

	var roles []string

	if v, ok := claims[ClaimRoles]; ok {
		switch o := v.(type) {
		case []string:
			roles = o
		case []interface{}:
			for _, v := range o {
				if z, ok := v.(string); ok {
					roles = append(roles, z)
				}
			}
		}
	}

	return user, roles, duration, nil
}
