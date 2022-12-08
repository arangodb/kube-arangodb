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

package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
)

var cmdLifecycleStartup = &cobra.Command{
	Use:    "startup",
	RunE:   cmdLifecycleStartupFunc,
	Hidden: true,
}

func cmdLifecycleStartupFunc(cmd *cobra.Command, args []string) error {
	var close bool

	port := ProbePort.GetOrDefault(fmt.Sprintf("%d", shared.ArangoPort))

	server := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
	}

	handlers := http.NewServeMux()

	handlers.HandleFunc("/stop", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		close = true
	})

	server.Handler = handlers

	go func() {
		for {
			if close {
				break
			}
			time.Sleep(time.Millisecond)
		}
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}

	return nil
}
