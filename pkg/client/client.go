//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
// Author Ewout Prangsma
//

package client

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"

	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

var (
	maskAny = errors.WithStack
)

// MustNewInCluster creates an in-cluster client, or panics
// when a failure is detected.
func MustNewInCluster() versioned.Interface {
	cli, err := NewInCluster()
	if err != nil {
		panic(err)
	}
	return cli
}

// MustNew creates a client with given config, or panics
// when a failure is detected.
func MustNew(cfg *rest.Config) versioned.Interface {
	cli, err := New(cfg)
	if err != nil {
		panic(err)
	}
	return cli
}

// NewInCluster creates an in-cluster client, or returns an error
// when a failure is detected.
func NewInCluster() (versioned.Interface, error) {
	cfg, err := k8sutil.InClusterConfig()
	if err != nil {
		return nil, maskAny(err)
	}
	cli, err := New(cfg)
	if err != nil {
		return nil, maskAny(err)
	}
	return cli, nil
}

// New creates a client with given config, or returns an error
// when a failure is detected.
func New(cfg *rest.Config) (versioned.Interface, error) {
	cli, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, maskAny(err)
	}
	return cli, nil
}
