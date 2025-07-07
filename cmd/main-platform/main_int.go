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

package main

import (
	"errors"
	"os"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/platform"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

func main() {
	if err := mainE(); err != nil {
		var v cli.CommandExitCode
		if errors.As(err, &v) {
			os.Exit(v.ExitCode)
		}

		os.Exit(1)
	}
}

func mainE() error {
	c, err := platform.NewInstaller()
	if err != nil {
		return err
	}

	if err := logging.Init(c); err != nil {
		return err
	}

	if err := c.ExecuteContext(shutdown.Context()); err != nil {
		return err
	}

	return nil
}
