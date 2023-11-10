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

type crdDefinitionGetter func(opts *crds.CRDOptions) crds.Definition

type crdRegistration struct {
	getter      crdDefinitionGetter
	defaultOpts crds.CRDOptions
}

var (
	registeredCRDs = map[string]crdRegistration{}

	crdsLock sync.Mutex
)

func registerCRDWithPanic(getter crdDefinitionGetter, defaultOpts *crds.CRDOptions) {
	if defaultOpts == nil {
		defaultOpts = &crds.CRDOptions{}
	}
	if err := registerCRD(getter, *defaultOpts); err != nil {
		panic(err)
	}
}

func registerCRD(getter crdDefinitionGetter, defaultOpts crds.CRDOptions) error {
	crdsLock.Lock()
	defer crdsLock.Unlock()

	def := getter(nil)
	if _, ok := registeredCRDs[def.CRD.GetName()]; ok {
		return errors.Newf("CRD %s already exists", def.CRD.GetName())
	}
	registeredCRDs[def.CRD.GetName()] = crdRegistration{
		getter:      getter,
		defaultOpts: defaultOpts,
	}

	return nil
}

func GetDefaultCRDOptions() map[string]crds.CRDOptions {
	ret := make(map[string]crds.CRDOptions)
	for n, s := range registeredCRDs {
		ret[n] = s.defaultOpts
	}
	return ret
}
