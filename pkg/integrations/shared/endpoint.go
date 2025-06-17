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

package shared

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	pbAuthenticationV1 "github.com/arangodb/kube-arangodb/integrations/authentication/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ugrpc "github.com/arangodb/kube-arangodb/pkg/util/grpc"
)

type Endpoint struct {
	Address string
}

func (e *Endpoint) AuthClient() cache.Object[pbAuthenticationV1.AuthenticationV1Client] {
	return cache.NewObject[pbAuthenticationV1.AuthenticationV1Client](func(ctx context.Context) (pbAuthenticationV1.AuthenticationV1Client, time.Duration, error) {
		if e == nil {
			return nil, 0, errors.Errorf("Endpoint Ref is empty")
		}

		client, _, err := ugrpc.NewGRPCClient(ctx, pbAuthenticationV1.NewAuthenticationV1Client, e.Address)
		if err != nil {
			return nil, 0, err
		}

		return client, time.Hour, nil
	})
}

func (e *Endpoint) New(cmd *cobra.Command) error {
	if e == nil {
		return errors.Errorf("Endpoint Ref is empty")
	}

	f := cmd.Flags()

	addr, err := f.GetString("services.address")
	if err != nil {
		return err
	}

	*e = Endpoint{
		Address: addr,
	}

	return nil
}
