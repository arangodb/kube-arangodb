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

package integrations

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

type Factory func() Integration

type Integration interface {
	Name() string

	Description() string

	Register(cmd *cobra.Command, fs FlagEnvHandler) error

	Handler(ctx context.Context, cmd *cobra.Command) (svc.Handler, error)
}

type IntegrationEnablement interface {
	Integration

	EnabledTypes() (internal, external bool)
}

func GetIntegrationEnablement(in Integration) (internal, external bool) {
	if v, ok := in.(IntegrationEnablement); ok {
		return v.EnabledTypes()
	}

	return true, false
}
