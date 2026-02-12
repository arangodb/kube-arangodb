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

	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/cli"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/util/svc/authenticator"
)

func Register() (*cobra.Command, error) {
	var cmd cobra.Command

	cmd.Use = "sidecar"
	cmd.Short = "Runs the sidecar as a daemon with serving group"

	if err := cli.RegisterFlags(&cmd,
		flagAddress,
		flagGatewayAddress,
		flagKeyfile,
		flagAuth,
		flagHealthAddress,
		flagArangodb,
	); err != nil {
		return nil, err
	}

	cmd.RunE = cli.Runner{
		logging.Runner,
		cli.ValidateFlags(),
	}.With(run).Run

	return &cmd, nil
}

func healthConfiguration(cmd *cobra.Command) (svc.Configuration, error) {
	var cfg svc.Configuration

	if addr, err := flagHealthAddress.Get(cmd); err != nil {
		return svc.Configuration{}, err
	} else {
		cfg.Address = addr
	}

	return cfg, nil
}

func configuration(cmd *cobra.Command) (svc.Configuration, error) {
	var cfg svc.Configuration

	if addr, err := flagAddress.Get(cmd); err != nil {
		return svc.Configuration{}, err
	} else {
		cfg.Address = addr
	}

	if addr, err := flagGatewayAddress.Get(cmd); err != nil {
		return svc.Configuration{}, err
	} else {
		cfg.Gateway = &svc.ConfigurationGateway{Address: addr}
	}

	if keyfile, err := flagKeyfile.Get(cmd); err != nil {
		return svc.Configuration{}, err
	} else if keyfile != "" {
		cfg.TLSOptions = ktls.NewKeyfileTLSConfig(keyfile)
	}

	if auth, err := flagAuth.Get(cmd); err != nil {
		return svc.Configuration{}, err
	} else if auth != "" {
		cfg.Authenticator = authenticator.Required(authenticator.NewJWTAuthentication(auth))
	} else {
		cfg.Authenticator = authenticator.Required(authenticator.NewAlwaysAuthenticator())
	}

	return cfg, nil
}

func run(cmd *cobra.Command, args []string) error {
	return runWithContext(cmd.Context(), cmd)
}

func runWithContext(ctx context.Context, cmd *cobra.Command) error {
	var handlers []svc.Handler

	return runWithHealth(ctx, cmd, handlers...)
}

func runWithHealth(ctx context.Context, cmd *cobra.Command, handlers ...svc.Handler) error {
	cfg, err := healthConfiguration(cmd)
	if err != nil {
		return err
	}

	health, err := svc.NewHealthService(cfg, svc.Readiness, handlers...)
	if err != nil {
		return err
	}

	healthHandler := health.Start(ctx)

	logger.Str("address", healthHandler.Address()).Info("Health handler started")

	return runServer(cmd.Context(), cmd, health, handlers...)
}

func runServer(ctx context.Context, cmd *cobra.Command, health svc.Health, handlers ...svc.Handler) error {
	cfg, err := configuration(cmd)
	if err != nil {
		return err
	}

	svc, err := svc.NewService(cfg, handlers...)
	if err != nil {
		return err
	}

	s := svc.StartWithHealth(ctx, health)

	logger.Str("address", s.Address()).Str("httpAddress", s.HTTPAddress()).Str("type", "internal").Bool("ssl", cfg.TLSOptions != nil).Info("Service handler started")

	return s.Wait()
}
