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

package sidecar

import (
	"fmt"
	goStrings "strings"

	core "k8s.io/api/core/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type Core struct {
	Internal *bool
	External *bool
}

func (c *Core) GetInternal() bool {
	if c == nil || c.Internal == nil {
		return true
	}

	return *c.Internal
}

func (c *Core) GetExternal() bool {
	if c == nil || c.External == nil {
		return false
	}

	return *c.External
}

func (c *Core) Envs(int Integration, envs ...core.EnvVar) []core.EnvVar {
	cmd := goStrings.Join(util.FormatList(int.Name(), func(a string) string {
		return goStrings.ToUpper(a)
	}), "_")
	var r = []core.EnvVar{
		{
			Name:  fmt.Sprintf("INTEGRATION_%s_INTERNAL", cmd),
			Value: util.BoolSwitch(c.GetInternal(), "true", "false"),
		},
		{
			Name:  fmt.Sprintf("INTEGRATION_%s_EXTERNAL", cmd),
			Value: util.BoolSwitch(c.GetExternal(), "true", "false"),
		},
	}

	r = append(r, envs...)

	return r
}
