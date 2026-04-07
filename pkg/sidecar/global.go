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
	"context"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

// HandlerBuilder is the function used to build a sidecar handler.
type HandlerBuilder func(ctx context.Context, cmd *cobra.Command) (svc.Handler, bool, error)

// Subtype groups everything an integration contributes to the sidecar:
// its handler builder and the flags it owns. Each integration is
// self-contained and registers its own flags via init().
type Subtype struct {
	// Build constructs the handler for this integration. The bool return
	// indicates whether the handler should be added to the running set.
	Build HandlerBuilder

	// Flags are the CLI flags owned by this integration. They are
	// registered onto the parent sidecar command together with the
	// common flags.
	Flags []cli.FlagRegisterer
}

var global = util.NewRegisterer[string, Subtype]()

// register is a helper used by integration init() functions to register
// themselves with the global sidecar registry.
func register(name string, build HandlerBuilder, flags ...cli.FlagRegisterer) {
	global.MustRegister(name, Subtype{
		Build: build,
		Flags: flags,
	})
}
