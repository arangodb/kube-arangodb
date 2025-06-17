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

package v1

import (
	"fmt"
	"time"

	integrationsShared "github.com/arangodb/kube-arangodb/pkg/integrations/shared"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func NewConfiguration() Configuration {
	return Configuration{}
}

type Configuration struct {
	integrationsShared.Endpoint
	integrationsShared.Database

	Prefix string
	TTL    time.Duration
}

func (c Configuration) With(mods ...util.ModR[Configuration]) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}

func (c Configuration) Key(parts ...string) string {
	key := strings.Join(parts, "_")
	if c.Prefix == "" {
		return key
	}

	return fmt.Sprintf("%s_%s", c.Prefix, key)
}

func (c Configuration) Validate() error {
	return nil
}
