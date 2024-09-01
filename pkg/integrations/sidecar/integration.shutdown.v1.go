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
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

type IntegrationShutdownV1 struct {
	Core *Core
}

func (i IntegrationShutdownV1) Name() (string, string) {
	return "SHUTDOWN", "V1"
}

func (i IntegrationShutdownV1) Validate() error {
	return nil
}

func (i IntegrationShutdownV1) Args() (k8sutil.OptionPairs, error) {
	options := k8sutil.CreateOptionPairs()

	options.Add("--integration.shutdown.v1", true)

	options.Merge(i.Core.Args(i))

	return options, nil
}
