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

package crd

import (
	"github.com/arangodb/kube-arangodb/pkg/crd/crds"
)

func init() {
	defs := []func(...func(options *crds.CRDOptions)) crds.Definition{
		crds.PlatformStorageDefinitionWithOptions,
		crds.PlatformChartDefinitionWithOptions,
		crds.PlatformServiceDefinitionWithOptions,
	}
	for _, getDef := range defs {
		defFn := getDef // bring into scope
		registerCRDWithPanic(func(opts *crds.CRDOptions) crds.Definition {
			return defFn(opts.AsFunc())
		}, &crds.CRDOptions{
			WithSchema:   true,
			WithPreserve: false,
		})
	}
}
