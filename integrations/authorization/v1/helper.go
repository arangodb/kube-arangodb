//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"time"

	"google.golang.org/grpc"

	pbAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func ServiceClient(svc svc.Service, opts ...grpc.DialOption) cache.Object[pbAuthorizationV1.AuthorizationV1Client] {
	return cache.NewObject[pbAuthorizationV1.AuthorizationV1Client](func(ctx context.Context) (pbAuthorizationV1.AuthorizationV1Client, time.Duration, error) {
		conn, err := svc.Dial(opts...)
		if err != nil {
			return nil, 0, err
		}

		return pbAuthorizationV1.NewAuthorizationV1Client(conn), time.Hour * 24 * 365, nil
	})
}
