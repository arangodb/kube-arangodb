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
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type Config struct {
	Endpoint string
	Region   string

	DisableSSL bool

	HTTP     HTTP
	Provider Provider
	TLS      TLS
}

func (c Config) GetRegion() string {
	if c.Region == "" {
		return "us-east-1"
	}

	return c.Region
}

func (c Config) GetProvider() (credentials.Provider, error) {
	return c.Provider.Provider()
}

func (c Config) GetHttpClient() (*http.Client, error) {
	tls, err := c.TLS.configuration()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create TLS")
	}

	return &http.Client{
		Transport: c.HTTP.configuration(tls),
	}, nil
}

func (c Config) GetAWSSession() (client.ConfigProvider, error) {
	prov, err := c.GetProvider()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create Provider")
	}

	cl, err := c.GetHttpClient()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to create HTTP Client")
	}

	return session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials:      credentials.NewCredentials(prov),
			Endpoint:         util.NewType(c.Endpoint),
			S3ForcePathStyle: util.NewType(true),
			DisableSSL:       util.NewType(c.DisableSSL),
			HTTPClient:       cl,
		},
	})
}
