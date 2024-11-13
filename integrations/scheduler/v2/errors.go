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

package v2

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

func asGRPCError(err error) error {
	if err == nil {
		return nil
	}

	if kerrors.IsForbiddenC(err) {
		return status.Errorf(codes.PermissionDenied, "Permission Denied: %s", err.Error())
	}

	if kerrors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, "NotFound: %s", err.Error())
	}

	return status.Errorf(codes.Internal, "Unable to run action: %s", err.Error())
}
