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

package v1

import "github.com/pkg/errors"

// ServerGroupPortProtocol define supported protocols of listeners
type ServerGroupPortProtocol string

// Get returns current protocol. If is nil then default is returned
func (s *ServerGroupPortProtocol) Get() ServerGroupPortProtocol {
	if s == nil {
		return ServerGroupPortProtocolDefault
	}
	return *s
}

// New returns pointer to copy of protocol value
func (s ServerGroupPortProtocol) New() *ServerGroupPortProtocol {
	return &s
}

// Validate validates if protocol is known and have valid value
func (s *ServerGroupPortProtocol) Validate() error {
	if s == nil {
		return nil
	}

	switch v := *s; v {
	case ServerGroupPortProtocolHTTP, ServerGroupPortProtocolHTTPS:
		return nil
	default:
		return errors.Errorf("Unknown proto %s", v)
	}
}

const (
	// ServerGroupPortProtocolHTTP defines HTTP protocol
	ServerGroupPortProtocolHTTP ServerGroupPortProtocol = "http"
	// ServerGroupPortProtocolHTTPS defines HTTPS protocol
	ServerGroupPortProtocolHTTPS ServerGroupPortProtocol = "https"
	// ServerGroupPortProtocolDefault defines default (HTTP) protocol
	ServerGroupPortProtocolDefault = ServerGroupPortProtocolHTTP
)
