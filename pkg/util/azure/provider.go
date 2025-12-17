//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package azure

import (
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ProviderType string

const (
	ProviderTypeSecret ProviderType = "secret"
)

type Provider struct {
	Type ProviderType

	TenantID string

	Secret ProviderSecret
}

func (c Provider) GetCredentials() (azcore.TokenCredential, error) {
	switch c.Type {
	case ProviderTypeSecret:
		return c.Secret.getCredentials(c.TenantID)
	}

	return nil, errors.Errorf("unable to get credentials for type '%s'", c.Type)
}

type ProviderSecret struct {
	ClientID     string
	ClientIDFile string

	ClientSecret     string
	ClientSecretFile string
}

func (p ProviderSecret) getCredentials(tenantID string) (*azidentity.ClientSecretCredential, error) {
	id, err := p.GetClientID()
	if err != nil {
		return nil, err
	}

	secret, err := p.GetClientSecret()
	if err != nil {
		return nil, err
	}

	return azidentity.NewClientSecretCredential(tenantID, id, secret, nil)
}

func (p ProviderSecret) GetClientID() (string, error) {
	if f := p.ClientIDFile; f != "" {
		data, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	if f := p.ClientID; f != "" {
		return f, nil
	}

	return "", errors.New("no client id found")
}

func (p ProviderSecret) GetClientSecret() (string, error) {
	if f := p.ClientSecretFile; f != "" {
		data, err := os.ReadFile(f)
		if err != nil {
			return "", err
		}

		return string(data), nil
	}

	if f := p.ClientSecret; f != "" {
		return f, nil
	}

	return "", errors.New("no client secret found")
}
