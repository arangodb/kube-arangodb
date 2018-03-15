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

package operator

import (
	"context"
	"time"

	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	deplapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	lsapi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/deployment"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/storage"
	"github.com/rs/zerolog"
)

const (
	initRetryWaitTime = 30 * time.Second
)

type Event struct {
	Type         kwatch.EventType
	Deployment   *deplapi.ArangoDeployment
	LocalStorage *lsapi.ArangoLocalStorage
}

type Operator struct {
	Config
	Dependencies

	log           zerolog.Logger
	deployments   map[string]*deployment.Deployment
	localStorages map[string]*storage.LocalStorage
}

type Config struct {
	Namespace      string
	PodName        string
	ServiceAccount string
	CreateCRD      bool
}

type Dependencies struct {
	LogService logging.Service
	KubeCli    kubernetes.Interface
	KubeExtCli apiextensionsclient.Interface
	CRCli      versioned.Interface
}

// NewOperator instantiates a new operator from given config & dependencies.
func NewOperator(config Config, deps Dependencies) (*Operator, error) {
	o := &Operator{
		Config:        config,
		Dependencies:  deps,
		log:           deps.LogService.MustGetLogger("operator"),
		deployments:   make(map[string]*deployment.Deployment),
		localStorages: make(map[string]*storage.LocalStorage),
	}
	return o, nil
}

// Start the operator
func (o *Operator) Start() error {
	log := o.log

	for {
		if err := o.initResourceIfNeeded(); err == nil {
			break
		} else {
			log.Error().Err(err).Msg("Resource initialization failed")
			log.Info().Msgf("Retrying in %s...", initRetryWaitTime)
			time.Sleep(initRetryWaitTime)
		}
	}

	//probe.SetReady()
	o.run()
	panic("unreachable")
}

// run the operator.
// This registers a listener and waits until the process stops.
func (o *Operator) run() {
	log := o.log

	log.Info().Msgf("Running controller in namespace '%s'", o.Config.Namespace)

	go o.runDeployments()
	go o.runLocalStorages()

	// Wait till done
	ctx := context.TODO()
	<-ctx.Done()
}
