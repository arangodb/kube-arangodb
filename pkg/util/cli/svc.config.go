package cli

type ServiceConfigurationInput struct {
	Address string

	TLS ServiceConfigurationInputTLS

	Gateway ServiceConfigurationInputGateway
}

type ServiceConfigurationInputTLS struct {
	Enabled bool
}

type ServiceConfigurationInputGateway struct {
	Enabled bool
	Address string
}
