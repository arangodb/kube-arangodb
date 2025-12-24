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
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/api/server"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

func (i *implementation) GetVersion(ctx context.Context, empty *definition.Empty) (*server.Version, error) {
	if i.authenticate(ctx) != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated")
	}

	v := version.GetVersionV1()

	return &server.Version{
		Version:   string(v.Version),
		Build:     v.Build,
		Edition:   string(v.Edition),
		GoVersion: v.GoVersion,
		BuildDate: v.BuildDate,
	}, nil
}
