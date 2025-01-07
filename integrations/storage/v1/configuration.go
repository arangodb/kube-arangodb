package v1

import (
	pbImplStorageV1SharedS3 "github.com/arangodb/kube-arangodb/integrations/storage/v1/shared/s3"
)

type Mod func(c Configuration) Configuration

type ConfigurationType string

const (
	ConfigurationTypeS3 ConfigurationType = "s3"
)

func NewConfiguration(mods ...Mod) Configuration {
	var cfg Configuration

	return cfg.With(mods...)
}

type Configuration struct {
	Type ConfigurationType

	S3 pbImplStorageV1SharedS3.Configuration
}

func (c Configuration) Validate() error {
	return nil
}

func (c Configuration) With(mods ...Mod) Configuration {
	n := c

	for _, mod := range mods {
		n = mod(n)
	}

	return n
}
