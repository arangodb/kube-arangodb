//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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

package features

func init() {
	registerFeature(gatewayWebSockets)
}

var gatewayWebSockets = &feature{
	name:               "gateway-websockets",
	description:        "Defines if the gateway enables WebSocket upgrades over HTTP/2 (RFC 8441 Extended CONNECT) on its downstream listener",
	enterpriseRequired: false,
	enabledByDefault:   true,
	hidden:             true,
}

// GatewayWebSockets returns the feature gating WebSocket-over-HTTP/2 support on the gateway.
func GatewayWebSockets() Feature {
	return gatewayWebSockets
}
