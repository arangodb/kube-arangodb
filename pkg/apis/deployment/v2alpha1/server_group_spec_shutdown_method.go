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

package v2alpha1

// ServerGroupShutdownMethod enum of possible shutdown methods
type ServerGroupShutdownMethod string

// Default return default value for ServerGroupShutdownMethod
func (s *ServerGroupShutdownMethod) Default() ServerGroupShutdownMethod {
	return ServerGroupShutdownMethodAPI
}

// Get return current or default value of ServerGroupShutdownMethod
func (s *ServerGroupShutdownMethod) Get() ServerGroupShutdownMethod {
	if s == nil {
		return s.Default()
	}

	switch t := *s; t {
	case ServerGroupShutdownMethodAPI, ServerGroupShutdownMethodDelete:
		return t
	default:
		return s.Default()
	}
}

const (
	// ServerGroupShutdownMethodAPI API Shutdown method
	ServerGroupShutdownMethodAPI ServerGroupShutdownMethod = "api"
	// ServerGroupShutdownMethodDelete Pod Delete shutdown method
	ServerGroupShutdownMethodDelete ServerGroupShutdownMethod = "delete"
)
