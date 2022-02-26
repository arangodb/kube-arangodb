//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package inspector

import (
	"context"

	"strings"

	"github.com/arangodb/go-driver"
	"k8s.io/client-go/kubernetes"
)

// GetVersionInfo returns kubernetes server version information.
func (i *inspector) GetVersionInfo() driver.Version {
	i.lock.Lock()
	defer i.lock.Unlock()

	return i.versionInfo
}

func getVersionInfo(_ context.Context, inspector *inspector, k kubernetes.Interface, _ string) func() error {
	return func() error {
		inspector.versionInfo = ""
		if v, err := k.Discovery().ServerVersion(); err != nil {
			return err
		} else {
			inspector.versionInfo = driver.Version(strings.TrimPrefix(v.GitVersion, "v"))
		}

		return nil
	}
}
