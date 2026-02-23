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

package authentication

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sidecarSvcAuthnDefinition "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authentication/definition"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
)

func NewClientCache(conn *grpc.ClientConn) cache.Object[[]string] {
	client := sidecarSvcAuthnDefinition.NewSidecarAuthenticationServiceClient(conn)

	return cache.NewObjectHash(func(ctx context.Context, old *string) ([]string, string, time.Duration, error) {
		resp, err := client.GetOptionalKeys(ctx, &sidecarSvcAuthnDefinition.SidecarAuthenticationKeysRequest{Checksum: old})
		if err != nil {
			switch status.Code(err) {
			case codes.AlreadyExists:
				// No change
				return nil, util.OptionalType(old, ""), 15 * time.Second, nil
			case codes.Unavailable:
				// No auth, cache forever
				return nil, "", time.Hour, nil
			}

			return nil, "", 0, err
		}

		return resp.GetKeys(), resp.GetChecksum(), 15 * time.Second, nil
	})
}
