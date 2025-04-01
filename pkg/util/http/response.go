//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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
	goHttp "net/http"
	goStrings "strings"
)

// NewSimpleJSONResponse returns handler which server static json on GET request
func NewSimpleJSONResponse(obj interface{}) (goHttp.Handler, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	return simpleJSONResponse{data: data}, nil
}

type simpleJSONResponse struct {
	data []byte
}

func (s simpleJSONResponse) ServeHTTP(writer goHttp.ResponseWriter, request *goHttp.Request) {
	if goStrings.ToUpper(request.Method) != goHttp.MethodGet {
		writer.WriteHeader(goHttp.StatusMethodNotAllowed)
		return
	}

	writer.WriteHeader(goHttp.StatusOK)
	writer.Write(s.data)
}
