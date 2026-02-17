//
// DISCLAIMER
//
// Copyright 2025-2026 ArangoDB GmbH, Cologne, Germany
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

import "github.com/arangodb/kube-arangodb/pkg/util"

const (
	CONTAINER_CPU_REQUESTS    util.EnvironmentVariable = "CONTAINER_CPU_REQUESTS"
	CONTAINER_MEMORY_REQUESTS util.EnvironmentVariable = "CONTAINER_MEMORY_REQUESTS"
	CONTAINER_CPU_LIMITS      util.EnvironmentVariable = "CONTAINER_CPU_LIMITS"
	CONTAINER_MEMORY_LIMITS   util.EnvironmentVariable = "CONTAINER_MEMORY_LIMITS"

	KUBERNETES_NAMESPACE       util.EnvironmentVariable = "KUBERNETES_NAMESPACE"
	KUBERNETES_POD_NAME        util.EnvironmentVariable = "KUBERNETES_POD_NAME"
	KUBERNETES_POD_IP          util.EnvironmentVariable = "KUBERNETES_POD_IP"
	KUBERNETES_SERVICE_ACCOUNT util.EnvironmentVariable = "KUBERNETES_SERVICE_ACCOUNT"

	ARANGO_DEPLOYMENT_NAME     util.EnvironmentVariable = "ARANGO_DEPLOYMENT_NAME"
	ARANGO_DEPLOYMENT_ENDPOINT util.EnvironmentVariable = "ARANGO_DEPLOYMENT_ENDPOINT"
	ARANGODB_ENDPOINT          util.EnvironmentVariable = "ARANGODB_ENDPOINT"

	AUTHENTICATION_ENABLED util.EnvironmentVariable = "AUTHENTICATION_ENABLED"

	ARANGO_DEPLOYMENT_CA util.EnvironmentVariable = "ARANGO_DEPLOYMENT_CA"

	INTEGRATION_SERVICE_ADDRESS util.EnvironmentVariable = "INTEGRATION_SERVICE_ADDRESS"

	INTEGRATION_API_ADDRESS       util.EnvironmentVariable = "INTEGRATION_API_ADDRESS"
	INTEGRATION_HTTP_ADDRESS      util.EnvironmentVariable = "INTEGRATION_HTTP_ADDRESS"
	INTEGRATION_HTTP_ADDRESS_FULL util.EnvironmentVariable = "INTEGRATION_HTTP_ADDRESS_FULL"

	INTEGRATION_ARANGO_TOKEN util.EnvironmentVariable = TokenEnvName

	CENTRAL_INTEGRATION_SERVICE_ADDRESS   util.EnvironmentVariable = "CENTRAL_INTEGRATION_SERVICE_ADDRESS"
	CENTRAL_INTEGRATION_SECURED           util.EnvironmentVariable = "CENTRAL_INTEGRATION_SECURED"
	CENTRAL_INTEGRATION_HTTP_ADDRESS      util.EnvironmentVariable = "CENTRAL_INTEGRATION_HTTP_ADDRESS"
	CENTRAL_INTEGRATION_HTTP_ADDRESS_FULL util.EnvironmentVariable = "CENTRAL_INTEGRATION_HTTP_ADDRESS_FULL"
)
