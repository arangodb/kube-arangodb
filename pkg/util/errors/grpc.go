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

package errors

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcError interface {
	error
	GRPCStatus() *status.Status
}

func GRPCStatus(err error) (*status.Status, bool) {
	v, ok := ExtractCauseHelper[grpcError](err, func(err error) (grpcError, bool) {
		var gs grpcError
		if errors.As(err, &gs) {
			return gs, true
		}

		return nil, false
	})

	if !ok {
		return status.New(codes.Unknown, err.Error()), false
	}

	return v.GRPCStatus(), true
}

func GRPCCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	if v, ok := GRPCStatus(err); ok {
		return v.Code()
	}

	return codes.Unknown
}

func IsGRPCCode(err error, codes ...codes.Code) bool {
	vc := GRPCCode(err)

	for _, code := range codes {
		if vc == code {
			return true
		}
	}

	return false
}
