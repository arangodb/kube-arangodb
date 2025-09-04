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

package grpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"

	pbSharedV1 "github.com/arangodb/kube-arangodb/integrations/shared/v1/definition"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewGRPCError(code codes.Code, msg string, args ...interface{}) GRPCError {
	return &grpcError{
		err: status.Newf(code, msg, args...),
	}
}

type GRPCError interface {
	With(err ...error) GRPCError
	Err() error
}

type grpcError struct {
	err *status.Status
}

func (g grpcError) With(errs ...error) GRPCError {
	if g.err.Code() == codes.OK {
		return g
	}

	e := g.err

	for _, err := range errs {
		if v, ok := err.(errors.Array); ok {
			if len(v) > 0 {
				p := make([]protoadapt.MessageV1, len(v))

				for i, n := range v {
					p[i] = AsGRPCMessage(n)
				}

				if q, err := g.err.WithDetails(p...); err == nil {
					e = q
				}
			}
		} else {
			if q, err := e.WithDetails(AsGRPCMessage(err)); err == nil {
				e = q
			}
		}
	}

	return grpcError{
		err: e,
	}
}

func (g grpcError) Err() error {
	return g.err.Err()
}

type Interface interface {
	AsGRPCError() protoadapt.MessageV1
}

func AsGRPCMessage(err error) protoadapt.MessageV1 {
	if err == nil {
		return &pbSharedV1.Error{Message: "unknown error"}
	}

	if v, ok := err.(Interface); ok {
		return v.AsGRPCError()
	}

	return &pbSharedV1.Error{Message: err.Error()}
}
