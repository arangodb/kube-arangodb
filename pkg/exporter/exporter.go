//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Adam Janikowski
//

package exporter

import (
	"fmt"
	"net/http"
)

type Authentication func() (string, error)

// CreateArangodJwtAuthorizationHeader calculates a JWT authorization header, for authorization
// of a request to an arangod server, based on the given secret.
// If the secret is empty, nothing is done.
func CreateArangodJwtAuthorizationHeader(jwt string) (string, error) {
	return "bearer " + jwt, nil
}

func NewExporter(endpoint string, port int, url string, handler http.Handler) http.Server {
	s := http.NewServeMux()

	s.Handle(url, handler)

	return http.Server{
		Addr:    fmt.Sprintf("%s:%d", endpoint, port),
		Handler: s,
	}
}
