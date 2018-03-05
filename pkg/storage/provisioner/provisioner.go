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

package provisioner

import (
	"context"

	"github.com/rs/zerolog"
	"k8s.io/client-go/kubernetes"
)

// Config for the storage provisioner
type Config struct {
	NodeName       string
	Namespace      string
	ServiceAccount string
	LocalPath      []string
}

// Dependencies for the storage provisioner
type Dependencies struct {
	Log     zerolog.Logger
	KubeCli kubernetes.Interface
}

// Provisioner implements a Local storage provisioner
type Provisioner struct {
}

// New creates a new local storage provisioner
func New(config Config, deps Dependencies) (*Provisioner, error) {
	return &Provisioner{}, nil
}

// Run the provisioner until the given context is canceled.
func (p *Provisioner) Run(ctx context.Context) {
	// TODO
	<-ctx.Done()
}
