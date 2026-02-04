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
	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func Register(cmd *cobra.Command) error {
	if err := cli.RegisterFlags(cmd, allFlags...); err != nil {
		return err
	}

	cmd.RunE = cli.Runner{
		logging.Runner,
		cli.ValidateFlags(allFlags...),
	}.With(run).Run

	return nil
}

func run(cmd *cobra.Command, args []string) error {
	handlers, err := services(cmd)
	if err != nil {
		return err
	}

	return runHealth(cmd, handlers...)
}

func runHealth(cmd *cobra.Command, handlers ...svc.Handler) error {
	cfg, err := flagHealth.Configuration(cmd)
	if err != nil {
		return err
	}

	health, err := svc.NewHealthService(cfg, svc.Readiness, handlers...)
	if err != nil {
		return err
	}

	start := health.Start(cmd.Context())

	q := logger.Str("address", start.Address()).Bool("secured", cfg.TLSOptions != nil)

	if z := start.HTTPAddress(); z != "" {
		q = q.Str("http-address", z)
	}

	q.Info("Service health started")

	if err := runService(cmd, health, handlers...); err != nil {
		return err
	}

	return start.Wait()
}

func runService(cmd *cobra.Command, health svc.HealthService, handlers ...svc.Handler) error {
	cfg, err := flagServer.Configuration(cmd)
	if err != nil {
		return err
	}

	svc, err := svc.NewService(cfg, handlers...)
	if err != nil {
		return err
	}

	start := svc.StartWithHealth(cmd.Context(), health)

	q := logger.Str("address", start.Address()).Bool("secured", cfg.TLSOptions != nil)

	if z := start.HTTPAddress(); z != "" {
		q = q.Str("http-address", z)
	}

	q.Info("Service started")

	return start.Wait()
}
