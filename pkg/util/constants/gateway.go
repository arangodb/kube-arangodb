//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package constants

const (
	ConfigMapChecksumKey = "CHECKSUM"

	ArangoGatewayExecutor        = "/usr/local/bin/envoy"
	GatewayVolumeMountDir        = "/etc/gateway/"
	GatewayVolumeName            = "gateway"
	GatewayConfigFileName        = "gateway.yaml"
	GatewayDynamicConfigFileName = "gateway.dynamic.yaml"
	GatewayCDSConfigFileName     = "gateway.dynamic.cds.yaml"
	GatewayLDSConfigFileName     = "gateway.dynamic.lds.yaml"
	GatewayConfigChecksumENV     = "GATEWAY_CONFIG_CHECKSUM"

	MemberConfigVolumeMountDir = "/etc/member/"
	MemberConfigVolumeName     = "member-config"
)
