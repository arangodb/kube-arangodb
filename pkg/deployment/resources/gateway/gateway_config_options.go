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

package gateway

type ConfigOptions struct {
	MergeSlashes *bool `json:"mergeSlashes,omitempty"`

	// WebSocketsHTTP2 enables RFC 8441 Extended CONNECT on the downstream listener so WebSocket
	// upgrades can be tunneled over HTTP/2. It is still gated on a destination declaring a websocket
	// upgrade. Defaults to false.
	WebSocketsHTTP2 *bool `json:"webSocketsHTTP2,omitempty"`
}

func (c *ConfigOptions) Validate() error {
	return nil
}

func (c *ConfigOptions) GetMergeSlashes() bool {
	if c == nil || c.MergeSlashes == nil {
		return true
	}

	return *c.MergeSlashes
}

func (c *ConfigOptions) GetWebSocketsHTTP2() bool {
	if c == nil || c.WebSocketsHTTP2 == nil {
		return false
	}

	return *c.WebSocketsHTTP2
}
