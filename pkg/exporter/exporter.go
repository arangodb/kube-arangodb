//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package exporter

import (
	"net/http"
	"time"

	operatorHTTP "github.com/arangodb/kube-arangodb/pkg/util/http"
)

type Authentication func() (string, error)

// CreateArangodJwtAuthorizationHeader calculates a JWT authorization header, for authorization
// of a request to an arangod server, based on the given secret.
// If the secret is empty, nothing is done.
func CreateArangodJwtAuthorizationHeader(jwt string) (string, error) {
	return "bearer " + jwt, nil
}

func NewExporter(endpoint string, url string, handler http.Handler) operatorHTTP.PlainServer {
	s := http.NewServeMux()

	s.Handle(url, handler)

	s.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>ArangoDB Exporter</title></head>
             <body>
             <h1>ArangoDB Exporter</h1>
             <p><a href='/metrics'>Metrics</a></p>
             </body>
             </html>`))
	})

	return operatorHTTP.NewServer(&http.Server{
		Addr:              endpoint,
		ReadTimeout:       time.Second * 30,
		ReadHeaderTimeout: time.Second * 15,
		WriteTimeout:      time.Second * 30,
		Handler:           s,
	})
}
