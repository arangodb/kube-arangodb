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

package errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCStatus(err error) (*status.Status, bool) {
	v, ok := ExtractCause[GRPCErrorStatus](err)

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

type GRPCErrorStatus interface {
	error

	GRPCStatus() *status.Status
}

func AsGRPCErrorStatus(err error) (GRPCErrorStatus, bool) {
	var v GRPCErrorStatus
	if As(err, &v) {
		return v, true
	}
	return nil, false
}

func ExtractGRPCCause(err error) error {
	if err == nil {
		return nil
	}

	if v, ok := AsGRPCErrorStatus(err); ok {
		return Errorf("%s", v.GRPCStatus().Message())
	}

	return err
}
