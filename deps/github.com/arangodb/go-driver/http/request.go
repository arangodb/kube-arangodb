//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package http

import (
	"net/http"
	"net/http/httptrace"
	"net/url"
)

// httpRequest implements driver.Request using standard golang http requests.
type httpRequest interface {
	// createHTTPRequest creates a golang http.Request based on the configured arguments.
	createHTTPRequest(endpoint url.URL) (*http.Request, error)
	// WroteRequest implements the WroteRequest function of an httptrace.
	// It sets written to true.
	WroteRequest(httptrace.WroteRequestInfo)
}
