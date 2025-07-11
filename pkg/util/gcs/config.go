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

package gcs

import (
	"google.golang.org/api/option"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Config struct {
	ProjectID string

	Provider Provider
}

func (c Config) GetClientOptions() ([]option.ClientOption, error) {
	if c.ProjectID == "" {
		return nil, errors.New("projectID is required")
	}

	var r = make([]option.ClientOption, 0, 2)
	if auth, err := c.Provider.Provider(); err != nil {
		return nil, err
	} else {
		r = append(r, auth)
	}

	return r, nil
}
