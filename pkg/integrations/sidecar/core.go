//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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
	"strings"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
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

func (c *Core) Args(int Integration) k8sutil.OptionPairs {
	var options k8sutil.OptionPairs
	cmd := strings.Join(util.FormatList(int.Name(), func(a string) string {
		return strings.ToLower(a)
	}), ".")

	options.Add(fmt.Sprintf("--integration.%s.internal", cmd), c.GetInternal())
	options.Add(fmt.Sprintf("--integration.%s.external", cmd), c.GetExternal())

	return options
}
