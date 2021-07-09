//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

package webhook

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

func NewWebServer(log zerolog.Logger, host string, port uint16, key, cert, prefix string) (*http.Server, error) {
	s := &http.Server{}

	s.Addr = fmt.Sprintf("%s:%d", host, port)

	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	s.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{
			tlsCert,
		},
	}

	mux := http.NewServeMux()

	// Validation
	validationLogger := log.With().Str("type", "validation").Logger()
	mux.Handle(fmt.Sprintf("%s/validate", prefix), NewValidationWebhookHandler(validationLogger))

	s.Handler = mux

	return s, nil
}
