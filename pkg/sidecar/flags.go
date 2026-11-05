//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package sidecar

import (
	"fmt"
	"time"

	pbImplAuthorizationV1 "github.com/arangodb/kube-arangodb/integrations/authorization/v1"
	shared "github.com/arangodb/kube-arangodb/pkg/apis/shared"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
)

var (
	flagAddress = cli.Flag[string]{
		Name:        "sidecar.address",
		Description: "Address of the server",
		Default:     fmt.Sprintf("0.0.0.0:%d", shared.InternalSidecarContainerPortGRPC),
	}
	flagGatewayAddress = cli.Flag[string]{
		Name: "sidecar.gateway.address",
		Description: "Address of the internal http gateway server. Defaults to a loopback address: the " +
			"endpoint is consumed only in-Pod (e.g. by arangod), so it is served plain HTTP; bind a routable " +
			"address to have it served over TLS.",
		Default: fmt.Sprintf("127.0.0.1:%d", shared.InternalSidecarContainerPortHTTP),
	}
	flagGatewayExternalAddress = cli.Flag[string]{
		Name: "sidecar.gateway.external.address",
		Description: "Address of the external http gateway server, serving the same routes as the internal " +
			"one on a routable address so an out-of-Pod client (the platform gateway) can reach the management " +
			"API. It follows the deployment TLS setting - served over TLS when a keyfile is configured, plain " +
			"otherwise. Empty disables it.",
		Default: "",
	}
	flagHealthAddress = cli.Flag[string]{
		Name:        "sidecar.health.address",
		Description: "Address of the health server",
		Default:     fmt.Sprintf("0.0.0.0:%d", shared.InternalSidecarContainerPortHealth),
	}
	flagKeyfile = cli.Flag[string]{
		Name:        "sidecar.keyfile",
		Description: "Path to the keyfile",
		Default:     "",
	}
	flagAuth = cli.Flag[string]{
		Name:        "sidecar.auth",
		Description: "Path to the JWT Folder",
		Default:     "",
	}
	flagAuthMode = cli.Flag[string]{
		Name:        "sidecar.auth.mode",
		Description: "Auth Mode",
		Default:     string(pbImplAuthorizationV1.ConfigurationTypeCentral),
	}
	flagArangodb = cli.Flag[string]{
		Name:        "arangodb.endpoint",
		Description: "ArangoDB Endpoint",
		Default:     "",
	}
	flagCentralServicesEnabled = cli.Flag[bool]{
		Name:        "central",
		Description: "Defines if central services are enabled",
		Default:     false,
	}
	flagUnixEnabled = cli.Flag[bool]{
		Name: "sidecar.unix.enabled",
		Description: "Defines if the internal UNIX socket (used for service-to-service calls that bypass " +
			"the network authenticator) is served. Disabling it stops the sidecar creating the socket directory.",
		Default: true,
	}
	flagAuthDeletedTTL = cli.Flag[time.Duration]{
		Name:        "sidecar.auth.deleted-ttl",
		Description: "TTL for soft-deleted RBAC documents before permanent removal from ArangoDB collections",
		Default:     30 * 24 * time.Hour,
	}
)
