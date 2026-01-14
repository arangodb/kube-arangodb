//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

package v1

import (
	pbImplAuthorizationV1Shared "github.com/arangodb/kube-arangodb/integrations/authorization/v1/shared"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ConfigurationType string

func (c ConfigurationType) Validate() error {
	switch c {
	case ConfigurationTypeAlways:
		return nil
	default:
		return errors.Errorf("Configuration type '%s' is not supported", string(c))
	}
}

const (
	ConfigurationTypeAlways ConfigurationType = "always"
)

func NewConfiguration() Configuration {
	return Configuration{}
}

type Configuration struct {
	Type ConfigurationType `json:"type,omitempty"`
}

func (c Configuration) Plugin() (pbImplAuthorizationV1Shared.Plugin, error) {
	switch c.Type {
	case ConfigurationTypeAlways:
		return pbImplAuthorizationV1Shared.NewAlwaysPlugin(), nil
	default:
		return nil, errors.Errorf("Configuration type '%s' is not supported", string(c.Type))
	}
}

func (c Configuration) Validate() error {
	return errors.Errors(
		shared.PrefixResourceError("type", c.Type.Validate()),
	)
}

func (c Configuration) With(mods ...util.ModR[Configuration]) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}
