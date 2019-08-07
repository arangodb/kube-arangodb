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
	"github.com/arangodb/kube-arangodb/pkg/backup/operator/event"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"

	deplapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	replapi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1alpha"
	lsapi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/backup/handlers/arango/backup"
	"github.com/arangodb/kube-arangodb/pkg/backup/handlers/arango/policy"
	backupOper "github.com/arangodb/kube-arangodb/pkg/backup/operator"
	"github.com/arangodb/kube-arangodb/pkg/deployment"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/replication"
	"github.com/arangodb/kube-arangodb/pkg/storage"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"k8s.io/client-go/rest"

	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
)

const (
	initRetryWaitTime = 30 * time.Second
)

type Event struct {
	Type                  kwatch.EventType
	Deployment            *deplapi.ArangoDeployment
	DeploymentReplication *replapi.ArangoDeploymentReplication
	LocalStorage          *lsapi.ArangoLocalStorage
}

type Operator struct {
	Config
	Dependencies

	log                    zerolog.Logger
	deployments            map[string]*deployment.Deployment
	deploymentReplications map[string]*replication.DeploymentReplication
	localStorages          map[string]*storage.LocalStorage
}

type Config struct {
	ID                          string
	Namespace                   string
	PodName                     string
	ServiceAccount              string
	LifecycleImage              string
	AlpineImage                 string
	EnableDeployment            bool
	EnableDeploymentReplication bool
	EnableStorage               bool
	EnableBackup                bool
	AllowChaos                  bool
}

type Dependencies struct {
	LogService                 logging.Service
	KubeCli                    kubernetes.Interface
	KubeExtCli                 apiextensionsclient.Interface
	CRCli                      versioned.Interface
	EventRecorder              record.EventRecorder
	LivenessProbe              *probe.LivenessProbe
	DeploymentProbe            *probe.ReadyProbe
	DeploymentReplicationProbe *probe.ReadyProbe
	StorageProbe               *probe.ReadyProbe
	BackupProbe                *probe.ReadyProbe
}

// NewOperator instantiates a new operator from given config & dependencies.
func NewOperator(config Config, deps Dependencies) (*Operator, error) {
	o := &Operator{
		Config:                 config,
		Dependencies:           deps,
		log:                    deps.LogService.MustGetLogger("operator"),
		deployments:            make(map[string]*deployment.Deployment),
		deploymentReplications: make(map[string]*replication.DeploymentReplication),
		localStorages:          make(map[string]*storage.LocalStorage),
	}
	return o, nil
}

// Run the operator
func (o *Operator) Run() {
	if o.Config.EnableDeployment {
		go o.runLeaderElection("arango-deployment-operator", o.onStartDeployment, o.Dependencies.DeploymentProbe)
	}
	if o.Config.EnableDeploymentReplication {
		go o.runLeaderElection("arango-deployment-replication-operator", o.onStartDeploymentReplication, o.Dependencies.DeploymentReplicationProbe)
	}
	if o.Config.EnableStorage {
		go o.runLeaderElection("arango-storage-operator", o.onStartStorage, o.Dependencies.StorageProbe)
	}
	if o.Config.EnableBackup {
		go o.runLeaderElection("arango-backup-operator", o.onStartBackup, o.Dependencies.BackupProbe)
	}
	// Wait until process terminates
	<-context.TODO().Done()
}

// onStartDeployment starts the deployment operator and run till given channel is closed.
func (o *Operator) onStartDeployment(stop <-chan struct{}) {
	for {
		if err := o.waitForCRD(true, false, false, false); err == nil {
			break
		} else {
			log.Error().Err(err).Msg("Resource initialization failed")
			log.Info().Msgf("Retrying in %s...", initRetryWaitTime)
			time.Sleep(initRetryWaitTime)
		}
	}
	o.runDeployments(stop)
}

// onStartDeploymentReplication starts the deployment replication operator and run till given channel is closed.
func (o *Operator) onStartDeploymentReplication(stop <-chan struct{}) {
	for {
		if err := o.waitForCRD(false, true, false, false); err == nil {
			break
		} else {
			log.Error().Err(err).Msg("Resource initialization failed")
			log.Info().Msgf("Retrying in %s...", initRetryWaitTime)
			time.Sleep(initRetryWaitTime)
		}
	}
	o.runDeploymentReplications(stop)
}

// onStartStorage starts the storage operator and run till given channel is closed.
func (o *Operator) onStartStorage(stop <-chan struct{}) {
	for {
		if err := o.waitForCRD(false, false, true, false); err == nil {
			break
		} else {
			log.Error().Err(err).Msg("Resource initialization failed")
			log.Info().Msgf("Retrying in %s...", initRetryWaitTime)
			time.Sleep(initRetryWaitTime)
		}
	}
	o.runLocalStorages(stop)
}

// onStartBackup starts the backup operator and run till given channel is closed.
func (o *Operator) onStartBackup(stop <-chan struct{}) {
	for {
		if err := o.waitForCRD(false, false, false, true); err == nil {
			break
		} else {
			log.Error().Err(err).Msg("Resource initialization failed")
			log.Info().Msgf("Retrying in %s...", initRetryWaitTime)
			time.Sleep(initRetryWaitTime)
		}
	}
	operatorName := "arangodb-backup-operator"
	operator := backupOper.NewOperator(operatorName, o.Namespace)

	rand.Seed(time.Now().Unix())

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	restClient, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	arangoClientSet, err := arangoClientSet.NewForConfig(restClient)
	if err != nil {
		panic(err)
	}

	kubeClientSet, err := kubernetes.NewForConfig(restClient)
	if err != nil {
		panic(err)
	}

	eventRecorder := event.NewEventRecorder(operatorName, kubeClientSet)

	arangoInformer := arangoInformer.NewSharedInformerFactoryWithOptions(arangoClientSet, 10 * time.Second, arangoInformer.WithNamespace(o.Namespace))

	if err = backup.RegisterInformer(operator, eventRecorder, arangoClientSet, kubeClientSet, arangoInformer); err != nil {
		panic(err)
	}

	if err = policy.RegisterInformer(operator, eventRecorder, arangoClientSet, kubeClientSet, arangoInformer); err != nil {
		panic(err)
	}

	if err = operator.RegisterStarter(arangoInformer); err != nil {
		panic(err)
	}

	prometheus.MustRegister(operator)

	operator.Start(4, stop)
	o.Dependencies.BackupProbe.SetReady()

	<-stop
}
