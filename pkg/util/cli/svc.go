package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	ktls "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/tls"
	"github.com/arangodb/kube-arangodb/pkg/util/svc"
)

func NewServiceConfiguration(prefix string, mods ...util.Mod[ServiceConfigurationInput]) ServiceConfiguration {
	var cfg ServiceConfigurationInput

	util.ApplyMods(&cfg, mods...)

	return &serviceConfiguration{
		address: Flag[string]{
			Name:        fmt.Sprintf("%s.address", prefix),
			Description: "Server address to listen on",
			Check: func(s string) error {
				if s == "" {
					return errors.Errorf("Argument cannot be empty")
				}

				return nil
			},

			Default: cfg.Address,
		},
		gateway: serviceConfigurationGateway{
			enabled: Flag[bool]{
				Name:        fmt.Sprintf("%s.gateway", prefix),
				Description: "Enables Gateway",
				Default:     cfg.Gateway.Enabled,
			},
			address: Flag[string]{
				Name:        fmt.Sprintf("%s.gateway.address", prefix),
				Description: "Server address to listen on Gateway",
				Check: func(s string) error {
					if s == "" {
						return errors.Errorf("Argument cannot be empty")
					}

					return nil
				},

				Default: cfg.Gateway.Address,
			},
		},
		tls: serviceConfigurationTLS{
			enabled: Flag[bool]{
				Name:        fmt.Sprintf("%s.tls", prefix),
				Description: "Enables TLS",
				Default:     cfg.TLS.Enabled,
			},
			file: Flag[string]{
				Name:        fmt.Sprintf("%s.keyfile", prefix),
				Description: "Path to the KeyFile",
				Default:     "",
			},
		},
	}
}

type ServiceConfiguration interface {
	FlagRegisterer

	Configuration(cmd *cobra.Command, mods ...util.Mod[svc.Configuration]) (svc.Configuration, error)
}

type serviceConfiguration struct {
	address Flag[string]
	gateway serviceConfigurationGateway
	tls     serviceConfigurationTLS
}

type serviceConfigurationTLS struct {
	enabled Flag[bool]
	file    Flag[string]
}

type serviceConfigurationGateway struct {
	address Flag[string]
	enabled Flag[bool]
}

func (s *serviceConfiguration) GetName() string {
	return "ServiceConfiguration"
}

func (s *serviceConfiguration) Register(cmd *cobra.Command) error {
	return RegisterFlags(
		cmd,
		s.address,
		s.gateway.address,
		s.gateway.enabled,
		s.tls.enabled,
		s.tls.file,
	)
}

func (s *serviceConfiguration) Validate(cmd *cobra.Command) error {
	return ValidateFlags(
		s.address,
		s.gateway.address,
		s.gateway.enabled,
		s.tls.enabled,
		s.tls.file,
	)(cmd, nil)
}

func (s *serviceConfiguration) Configuration(cmd *cobra.Command, mods ...util.Mod[svc.Configuration]) (svc.Configuration, error) {
	var cfg svc.Configuration

	if v, err := s.address.Get(cmd); err != nil {
		return cfg, err
	} else {
		cfg.Address = v
	}

	if enabled, err := s.gateway.enabled.Get(cmd); err != nil {
		return cfg, err
	} else if enabled {
		var gw svc.ConfigurationGateway

		if v, err := s.gateway.address.Get(cmd); err != nil {
			return cfg, err
		} else {
			gw.Address = v
		}

		cfg.Gateway = &gw
	}

	if enabled, err := s.tls.enabled.Get(cmd); err != nil {
		return cfg, err
	} else if enabled {
		if z, err := s.tls.file.Get(cmd); err != nil {
			return cfg, err
		} else if z == "" {
			cfg.TLSOptions = ktls.NewSelfSignedTLSConfig("selfsigned")
		} else {
			cfg.TLSOptions = ktls.NewKeyfileTLSConfig(z)
		}
	}

	util.ApplyMods(&cfg, mods...)

	return cfg, nil
}
