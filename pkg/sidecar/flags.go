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
		Name:        "sidecar.gateway.address",
		Description: "Address of the http gateway server",
		Default:     fmt.Sprintf("0.0.0.0:%d", shared.InternalSidecarContainerPortHTTP),
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
	flagArangodb = cli.Flag[string]{
		Name:        "arangodb.endpoint",
		Description: "ArangoDB Endpoint",
		Default:     "",
	}
)
