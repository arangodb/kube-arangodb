//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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

package v2

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

type Mod func(c Configuration) Configuration

func NewConfiguration() Configuration {
	return Configuration{
		Namespace:  "default",
		MaxHistory: 10,
	}
}

type Configuration struct {
	Namespace string

	Deployment string

	MaxHistory int
}

func (c Configuration) Validate() error {
	if c.Deployment == "" {
		return errors.Errorf("Invalid empty name of deployment")
	}

	if c.Namespace == "" {
		return errors.Errorf("Invalid empty name of namespace")
	}

	return nil
}

func (c Configuration) With(mods ...Mod) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}
