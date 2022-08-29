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

package crd

import (
	"sync"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const Version = "arangodb.com/version"

var (
	crds = map[string]crd{}

	crdsLock sync.Mutex
)

type crd struct {
	version driver.Version
	spec    apiextensions.CustomResourceDefinitionSpec
}

func registerCRDWithPanic(name string, crd crd) {
	if err := registerCRD(name, crd); err != nil {
		panic(err)
	}
}

func registerCRD(name string, crd crd) error {
	crdsLock.Lock()
	defer crdsLock.Unlock()

	if _, ok := crds[name]; ok {
		return errors.Newf("CRD %s already exists", name)
	}

	crds[name] = crd

	return nil
}
