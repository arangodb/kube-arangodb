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

package http

import (
	"encoding/json"
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func WrapError(code int, err error) error {
	if err == nil {
		return nil
	}

	return NewError(code, "%s", err.Error())
}

func NewError(code int, format string, args ...any) error {
	return Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

func IsError(err error) (Error, bool) {
	if err == nil {
		return Error{}, false
	}

	var v Error
	if errors.As(err, &v) {
		return v, true
	}

	return Error{}, false
}

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("HTTP Error (%d): %s", e.Code, e.Message)
}

func (e Error) JSON() []byte {
	data, err := json.Marshal(map[string]any{
		"Code":    e.Code,
		"Message": e.Message,
	})
	if err != nil {
		return nil
	}

	return data
}
