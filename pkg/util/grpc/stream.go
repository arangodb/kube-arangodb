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

package grpc

import (
	"io"

	"google.golang.org/grpc"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type RecvInterface[T any] interface {
	Recv() (T, error)
	grpc.ClientStream
}

func Recv[T any](recv RecvInterface[T], parser func(T) error) error {
	for {
		resp, err := recv.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			if cerr := recv.CloseSend(); cerr != nil {
				return errors.Errors(err, cerr)
			}

			return err
		}

		if err := parser(resp); err != nil {
			return err
		}
	}
}

type SendInterface[T, O any] interface {
	Send(T) error
	CloseAndRecv() (O, error)
	grpc.ClientStream
}

func Send[T, O any](send SendInterface[T, O], batch func() (T, error)) (O, error) {
	for {
		v, err := batch()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return send.CloseAndRecv()
			}

			if cerr := send.CloseSend(); cerr != nil {
				return util.Default[O](), errors.Errors(err, cerr)
			}

			return util.Default[O](), err
		}

		if err := send.Send(v); err != nil {
			if cerr := send.CloseSend(); cerr != nil {
				return util.Default[O](), errors.Errors(err, cerr)
			}

			return util.Default[O](), err
		}
	}
}
