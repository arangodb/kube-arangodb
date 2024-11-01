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
	"context"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

type impersonate struct {
	lock sync.Mutex

	credentials credentials.Value
	expires     time.Time

	config ProviderImpersonate

	creds credentials.Provider
}

func (i *impersonate) Retrieve() (credentials.Value, error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	awsSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials:      credentials.NewCredentials(i.creds),
			S3ForcePathStyle: util.NewType(true),
		},
	})
	if err != nil {
		return credentials.Value{}, err
	}

	s := sts.New(awsSession)

	resp, err := s.AssumeRoleWithContext(context.Background(), &sts.AssumeRoleInput{
		RoleArn:         util.NewType(i.config.Role),
		RoleSessionName: util.NewType(i.config.Name),
	})
	if err != nil {
		return credentials.Value{}, err
	}

	if e := resp.Credentials.Expiration; e != nil {
		i.expires = *e
	}

	i.credentials = credentials.Value{
		AccessKeyID:     util.WithDefault(resp.Credentials.AccessKeyId),
		SecretAccessKey: util.WithDefault(resp.Credentials.SecretAccessKey),
		SessionToken:    util.WithDefault(resp.Credentials.SessionToken),
	}

	return i.credentials, nil
}

func (i *impersonate) IsExpired() bool {
	i.lock.Lock()
	defer i.lock.Unlock()

	return time.Now().After(i.expires)
}
