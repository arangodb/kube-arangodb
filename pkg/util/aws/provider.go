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

package aws

import (
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ProviderType string

const (
	ProviderTypeConfig ProviderType = "config"
	ProviderTypeStatic ProviderType = "static"
	ProviderTypeFile   ProviderType = "file"
)

type Provider struct {
	Type ProviderType

	Config ProviderConfig
	Static ProviderConfigStatic
	File   ProviderConfigFile

	Impersonate ProviderImpersonate
}

func (p Provider) provider() (credentials.Provider, error) {
	switch p.Type {
	case ProviderTypeConfig:
		return p.Config.provider()
	case ProviderTypeStatic:
		return p.Static.provider()
	case ProviderTypeFile:
		return p.File.provider()
	default:
		return nil, errors.Errorf("Unknown provider: %s", p.Type)
	}
}

func (p Provider) Provider() (credentials.Provider, error) {
	prov, err := p.provider()
	if err != nil {
		return nil, err
	}

	return p.Impersonate.provider(prov)
}

type ProviderImpersonate struct {
	Impersonate bool

	Role string
	Name string
}

func (p ProviderImpersonate) provider(in credentials.Provider) (credentials.Provider, error) {
	if !p.Impersonate {
		return in, nil
	}

	return &impersonate{
		config: p,
		creds:  in,
	}, nil
}

type ProviderConfigStatic struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

func (p ProviderConfigStatic) provider() (credentials.Provider, error) {
	return &credentials.StaticProvider{
		Value: credentials.Value{
			AccessKeyID:     p.AccessKeyID,
			SecretAccessKey: p.SecretAccessKey,
			SessionToken:    p.SessionToken,
		},
	}, nil
}

type ProviderConfig struct {
	Filename string
	Profile  string
}

func (p ProviderConfig) provider() (credentials.Provider, error) {
	return &credentials.SharedCredentialsProvider{
		Filename: p.Filename,
		Profile:  p.Profile,
	}, nil
}

type ProviderConfigFile struct {
	AccessKeyIDFile     string
	SecretAccessKeyFile string
	SessionTokenFile    string
}

func (p *ProviderConfigFile) provider() (credentials.Provider, error) {
	return &fileProvider{
		accessKeyIDFile:     p.AccessKeyIDFile,
		secretAccessKeyFile: p.SecretAccessKeyFile,
		sessionTokenFile:    p.SessionTokenFile,
	}, nil
}
