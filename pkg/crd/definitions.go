//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

	"github.com/arangodb/kube-arangodb/pkg/crd/crds"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

const Version = "arangodb.com/version"

type Fetcher func(opts *crds.CRDOptions) crds.Definition

var (
	registeredCRDs = map[string]Fetcher{}

	crdsLock sync.Mutex
)

func registerCRDWithPanic(fetcher Fetcher) {
	if err := registerCRD(fetcher); err != nil {
		panic(err)
	}
}

func registerCRD(fetcher Fetcher) error {
	crdsLock.Lock()
	defer crdsLock.Unlock()

	crd := fetcher(nil)

	if _, ok := registeredCRDs[crd.CRD.GetName()]; ok {
		return errors.Newf("CRD %s already exists", crd.CRD.GetName())
	}

	registeredCRDs[crd.CRD.GetName()] = fetcher

	return nil
}
