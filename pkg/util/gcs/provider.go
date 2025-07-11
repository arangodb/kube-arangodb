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

package gcs

import (
	"google.golang.org/api/option"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type ProviderType string

const (
	ProviderTypeServiceAccount ProviderType = "serviceAccount"
)

type Provider struct {
	Type ProviderType

	ServiceAccount ProviderServiceAccount
}

func (p Provider) Provider() (option.ClientOption, error) {
	switch p.Type {
	case ProviderTypeServiceAccount:
		return p.ServiceAccount.provider()
	default:
		return nil, errors.Errorf("Unknown provider: %s", p.Type)
	}
}

type ProviderServiceAccount struct {
	JSON string
	File string
}

func (p ProviderServiceAccount) provider() (option.ClientOption, error) {
	if p.File != "" {
		return option.WithCredentialsFile(p.File), nil
	}

	if p.JSON != "" {
		return option.WithCredentialsJSON([]byte(p.JSON)), nil
	}

	return nil, errors.Errorf("No service account credentials file")
}
