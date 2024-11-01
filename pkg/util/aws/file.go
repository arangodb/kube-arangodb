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
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type fileProvider struct {
	lock sync.Mutex

	accessKeyIDFile     string
	secretAccessKeyFile string
	sessionTokenFile    string

	recent time.Time
}

func (f *fileProvider) Retrieve() (credentials.Value, error) {
	if f == nil {
		return credentials.Value{}, errors.Errorf("Object is nil")
	}

	f.lock.Lock()
	defer f.lock.Unlock()

	var v credentials.Value

	v.ProviderName = "dynamic-file-provider"

	if data, err := os.ReadFile(f.accessKeyIDFile); err != nil {
		return credentials.Value{}, errors.Wrapf(err, "Unable to open AccessKeyID File")
	} else {
		v.AccessKeyID = string(data)
	}

	if data, err := os.ReadFile(f.secretAccessKeyFile); err != nil {
		return credentials.Value{}, errors.Wrapf(err, "Unable to open SecretAccessKey File")
	} else {
		v.SecretAccessKey = string(data)
	}

	if f.sessionTokenFile != "" {
		if data, err := os.ReadFile(f.sessionTokenFile); err != nil {
			return credentials.Value{}, errors.Wrapf(err, "Unable to open SessionToken File")
		} else {
			v.SessionToken = string(data)
		}
	}

	f.recent = util.RecentFileModTime(f.accessKeyIDFile, f.secretAccessKeyFile, f.sessionTokenFile)

	return credentials.Value{}, nil
}

func (f *fileProvider) IsExpired() bool {
	f.lock.Lock()
	defer f.lock.Unlock()

	return util.RecentFileModTime(f.accessKeyIDFile, f.secretAccessKeyFile, f.sessionTokenFile).After(f.recent)
}
