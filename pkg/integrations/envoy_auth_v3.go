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

	pbImplEnvoyAuthV3 "github.com/arangodb/kube-arangodb/integrations/envoy/auth/v3"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func init() {
	registerer.Register(pbImplEnvoyAuthV3.Name, func() Integration {
		return &envoyAuthV3{}
	})
}

type envoyAuthV3 struct {
}

func (a envoyAuthV3) Name() string {
	return pbImplEnvoyAuthV3.Name
}

func (a envoyAuthV3) Description() string {
	return "Enable EnvoyAuthV3 Integration Service"
}

func (a envoyAuthV3) Register(cmd *cobra.Command, arg ArgGen) error {
	return nil
}

func (a envoyAuthV3) Handler(ctx context.Context) (svc.Handler, error) {
	return pbImplEnvoyAuthV3.New(), nil
}
