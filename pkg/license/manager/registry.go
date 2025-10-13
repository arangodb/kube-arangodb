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

package manager

import (
	"encoding/base64"
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Registry struct {
	Auths map[string]RegistryAuth `json:"auths,omitempty"`
}

type RegistryAuth struct {
	Client string `json:"client"`
	Auth   string `json:"auth,omitempty"`
}

func NewRegistryAuth(endpoint, username, password string, stages ...Stage) (*Registry, error) {
	if len(stages) == 0 {
		return nil, errors.Errorf("Enable Auth for at least one stage")
	}

	var r Registry

	r.Auths = map[string]RegistryAuth{}

	ra := RegistryAuth{
		Client: username,
		Auth:   base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password))),
	}

	for _, s := range stages {
		domain, err := s.RegistryDomain(endpoint)
		if err != nil {
			return nil, err
		}

		r.Auths[domain] = ra
	}

	return &r, nil
}
