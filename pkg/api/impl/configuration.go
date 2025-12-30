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

package impl

import (
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func NewConfiguration() Configuration {
	return Configuration{}
}

type Configuration struct {
	Authenticator authenticator.Authenticator

	LivenessProbe   *probe.LivenessProbe
	ReadinessProbes map[string]*probe.ReadyProbe
}

func (c Configuration) With(mods ...util.ModR[Configuration]) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}

func (c Configuration) WithReadinessProbe(name string, enabled bool, p *probe.ReadyProbe) Configuration {
	if !enabled {
		delete(c.ReadinessProbes, name)
		return c
	}

	if c.ReadinessProbes == nil {
		c.ReadinessProbes = map[string]*probe.ReadyProbe{}
	}

	c.ReadinessProbes[name] = p

	return c
}

func (c Configuration) Validate() error {
	return nil
}
