//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package integrations

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/cobra"

	pbImplPongV1 "github.com/arangodb/kube-arangodb/integrations/pong/v1"
	pbImplShutdownV1 "github.com/arangodb/kube-arangodb/integrations/shutdown/v1"
	"github.com/arangodb/kube-arangodb/pkg/integrations/clients"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

var registerer = util.NewRegisterer[string, Factory]()

func Register(cmd *cobra.Command) error {
	var c configuration

	return c.Register(cmd)
}

type configurationTest struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type configuration struct {
	// Only for testing
	test *configurationTest

	registered []Integration

	health struct {
		serviceConfiguration
		shutdownEnabled bool
	}

	services struct {
		internal, external serviceConfiguration
	}
}

type serviceConfiguration struct {
	enabled bool

	address string

	tls struct {
		keyfile string
	}

	auth struct {
		t string

		token string
	}
}

func (s *serviceConfiguration) Config() (svc.Configuration, error) {
	var cfg svc.Configuration

	cfg.Address = s.address

	switch strings.ToLower(s.auth.t) {
	case "none":
		break
	case "token":
		if s.auth.token == "" {
			return util.Default[svc.Configuration](), errors.Errorf("Token is empty")
		}

		cfg.Options = append(cfg.Options,
			basicTokenAuthUnaryInterceptor(s.auth.token),
			basicTokenAuthStreamInterceptor(s.auth.token),
		)
	}

	if keyfile := s.tls.keyfile; keyfile != "" {
		if tls, err := tlsServerOptions(keyfile); err != nil {
			return svc.Configuration{}, err
		} else {
			cfg.TLSOptions = tls
		}
	}

	return cfg, nil
}

func (c *configuration) Register(cmd *cobra.Command) error {
	c.registered = util.FormatList(registerer.Items(), func(a util.KV[string, Factory]) Integration {
		return a.V()
	})

	sort.Slice(c.registered, func(i, j int) bool {
		return c.registered[i].Name() < c.registered[j].Name()
	})

	cmd.RunE = c.run

	f := cmd.Flags()

	f.StringVar(&c.health.address, "health.address", "0.0.0.0:9091", "Address to expose health service")
	f.BoolVar(&c.health.shutdownEnabled, "health.shutdown.enabled", true, "Determines if shutdown service should be enabled and exposed")
	f.StringVar(&c.health.auth.t, "health.auth.type", "None", "Auth type for health service")
	f.StringVar(&c.health.auth.token, "health.auth.token", "", "Token for health service (when auth service is token)")
	f.StringVar(&c.health.tls.keyfile, "health.tls.keyfile", "", "Path to the keyfile")

	f.BoolVar(&c.services.internal.enabled, "services.enabled", true, "Defines if internal access is enabled")
	f.StringVar(&c.services.internal.address, "services.address", "127.0.0.1:9092", "Address to expose internal services")
	f.StringVar(&c.services.internal.auth.t, "services.auth.type", "None", "Auth type for internal service")
	f.StringVar(&c.services.internal.auth.token, "services.auth.token", "", "Token for internal service (when auth service is token)")
	f.StringVar(&c.services.internal.tls.keyfile, "services.tls.keyfile", "", "Path to the keyfile")

	f.BoolVar(&c.services.external.enabled, "services.external.enabled", false, "Defines if external access is enabled")
	f.StringVar(&c.services.external.address, "services.external.address", "0.0.0.0:9093", "Address to expose external services")
	f.StringVar(&c.services.external.auth.t, "services.external.auth.type", "None", "Auth type for external service")
	f.StringVar(&c.services.external.auth.token, "services.external.auth.token", "", "Token for external service (when auth service is token)")
	f.StringVar(&c.services.external.tls.keyfile, "services.external.tls.keyfile", "", "Path to the keyfile")

	for _, service := range c.registered {
		prefix := fmt.Sprintf("integration.%s", service.Name())

		f.Bool(prefix, false, service.Description())
		internal, external := GetIntegrationEnablement(service)
		f.Bool(fmt.Sprintf("%s.internal", prefix), internal, fmt.Sprintf("Defones if Internal access to service %s is enabled", service.Name()))
		f.Bool(fmt.Sprintf("%s.external", prefix), external, fmt.Sprintf("Defones if External access to service %s is enabled", service.Name()))

		if err := service.Register(cmd, func(name string) string {
			return fmt.Sprintf("%s.%s", prefix, name)
		}); err != nil {
			return errors.Wrapf(err, "Unable to register service %s", service.Name())
		}
	}

	return clients.Register(cmd)
}

func (c *configuration) run(cmd *cobra.Command, args []string) error {
	if t := c.test; t == nil {
		return c.runWithContext(shutdown.Context(), shutdown.Stop, cmd)
	} else {
		return c.runWithContext(t.ctx, t.cancel, cmd)
	}
}

func (c *configuration) runWithContext(ctx context.Context, cancel context.CancelFunc, cmd *cobra.Command) error {
	println(version.GetVersionV1().String())

	healthConfig, err := c.health.Config()
	if err != nil {
		return errors.Wrapf(err, "Unable to parse health config")
	}
	internalConfig, err := c.services.internal.Config()
	if err != nil {
		return errors.Wrapf(err, "Unable to parse internal config")
	}
	externalConfig, err := c.services.external.Config()
	if err != nil {
		return errors.Wrapf(err, "Unable to parse external config")
	}

	var internalHandlers, externalHandlers, healthHandlers []svc.Handler

	pong := pbImplPongV1.New()

	internalHandlers = append(internalHandlers, pong)
	externalHandlers = append(externalHandlers, pong)
	healthHandlers = append(healthHandlers, pong)

	for _, handler := range c.registered {
		if ok, err := cmd.Flags().GetBool(fmt.Sprintf("integration.%s", handler.Name())); err != nil {
			return err
		} else {
			internalEnabled, err := cmd.Flags().GetBool(fmt.Sprintf("integration.%s.internal", handler.Name()))
			if err != nil {
				return err
			}

			externalEnabled, err := cmd.Flags().GetBool(fmt.Sprintf("integration.%s.external", handler.Name()))
			if err != nil {
				return err
			}

			logger.
				Str("service", handler.Name()).
				Bool("enabled", ok).
				Bool("internal", internalEnabled).
				Bool("external", externalEnabled).
				Info("Service discovered")

			if ok && (internalEnabled || externalEnabled) {
				if svc, err := handler.Handler(ctx, cmd); err != nil {
					return err
				} else {
					if internalEnabled {
						internalHandlers = append(internalHandlers, svc)
					}

					if externalEnabled {
						externalHandlers = append(externalHandlers, svc)
					}
				}
			}
		}
	}

	if c.health.shutdownEnabled {
		healthHandlers = append(healthHandlers, pbImplShutdownV1.New(cancel))
	}

	health := svc.NewHealthService(healthConfig, svc.Readiness, healthHandlers...)

	internalHandlers = append(internalHandlers, health)
	externalHandlers = append(externalHandlers, health)

	healthHandler := health.Start(ctx)

	logger.Str("address", healthHandler.Address()).Bool("ssl", healthConfig.TLSOptions != nil).Info("Health handler started")

	var wg sync.WaitGroup

	var internal, external error

	if c.services.internal.enabled {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s := svc.NewService(internalConfig, internalHandlers...).StartWithHealth(ctx, health)

			logger.Str("address", s.Address()).Str("type", "internal").Bool("ssl", internalConfig.TLSOptions != nil).Info("Service handler started")

			internal = s.Wait()

			if internal != nil {
				logger.Err(internal).Str("address", s.Address()).Str("type", "internal").Error("Service handler failed")
			}
		}()
	}

	if c.services.external.enabled {
		wg.Add(1)

		go func() {
			defer wg.Done()
			s := svc.NewService(externalConfig, externalHandlers...).StartWithHealth(ctx, health)

			logger.Str("address", s.Address()).Str("type", "external").Bool("ssl", externalConfig.TLSOptions != nil).Info("Service handler started")

			external = s.Wait()

			if external != nil {
				logger.Err(external).Str("address", s.Address()).Str("type", "external").Error("Service handler failed")
			}
		}()
	}

	wg.Wait()

	return errors.Errors(internal, external)
}
