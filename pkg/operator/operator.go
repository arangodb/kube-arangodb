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

package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	kwatch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	backupdef "github.com/arangodb/kube-arangodb/pkg/apis/backup"
	depldef "github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	deplapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	repldef "github.com/arangodb/kube-arangodb/pkg/apis/replication"
	replapi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	lsapi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/deployment"
	arangoClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	arangoInformer "github.com/arangodb/kube-arangodb/pkg/generated/informers/externalversions"
	"github.com/arangodb/kube-arangodb/pkg/handlers/backup"
	"github.com/arangodb/kube-arangodb/pkg/handlers/job"
	"github.com/arangodb/kube-arangodb/pkg/handlers/policy"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/operator/scope"
	operatorV2 "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/replication"
	"github.com/arangodb/kube-arangodb/pkg/storage"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/probe"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

const (
	initRetryWaitTime = 30 * time.Second
)

var logger = logging.Global().RegisterAndGetLogger("operator", logging.Info)

type operatorV2type string

const (
	backupOperator operatorV2type = "backup"
	appsOperator   operatorV2type = "apps"
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

	log                    logging.Logger
	deployments            map[string]*deployment.Deployment
	deploymentReplications map[string]*replication.DeploymentReplication
	localStorages          map[string]*storage.LocalStorage
}

type Config struct {
	ID                          string
	Namespace                   string
	PodName                     string
	ServiceAccount              string
	OperatorImage               string
	EnableDeployment            bool
	EnableDeploymentReplication bool
	EnableStorage               bool
	EnableBackup                bool
	EnableApps                  bool
	EnableK2KClusterSync        bool
	AllowChaos                  bool
	ScalingIntegrationEnabled   bool
	SingleMode                  bool
	Scope                       scope.Scope
	ReconciliationDelay         time.Duration
	ShutdownDelay               time.Duration
	ShutdownTimeout             time.Duration
}

type Dependencies struct {
	Client                     kclient.Client
	EventRecorder              record.EventRecorder
	LivenessProbe              *probe.LivenessProbe
	DeploymentProbe            *probe.ReadyProbe
	DeploymentReplicationProbe *probe.ReadyProbe
	StorageProbe               *probe.ReadyProbe
	BackupProbe                *probe.ReadyProbe
	AppsProbe                  *probe.ReadyProbe
	K2KClusterSyncProbe        *probe.ReadyProbe
}

// NewOperator instantiates a new operator from given config & dependencies.
func NewOperator(config Config, deps Dependencies) (*Operator, error) {
	o := &Operator{
		Config:                 config,
		Dependencies:           deps,
		deployments:            make(map[string]*deployment.Deployment),
		deploymentReplications: make(map[string]*replication.DeploymentReplication),
		localStorages:          make(map[string]*storage.LocalStorage),
	}
	o.log = logger.WrapObj(o)
	return o, nil
}

// Run the operator
func (o *Operator) Run() {
	if o.Config.EnableDeployment {
		if !o.Config.SingleMode {
			go o.runLeaderElection("arango-deployment-operator", constants.LabelRole, o.onStartDeployment, o.Dependencies.DeploymentProbe)
		} else {
			go o.runWithoutLeaderElection("arango-deployment-operator", constants.LabelRole, o.onStartDeployment, o.Dependencies.DeploymentProbe)
		}
	}
	if o.Config.EnableDeploymentReplication {
		if !o.Config.SingleMode {
			go o.runLeaderElection("arango-deployment-replication-operator", constants.LabelRole, o.onStartDeploymentReplication, o.Dependencies.DeploymentReplicationProbe)
		} else {
			go o.runWithoutLeaderElection("arango-deployment-replication-operator", constants.LabelRole, o.onStartDeploymentReplication, o.Dependencies.DeploymentReplicationProbe)
		}
	}
	if o.Config.EnableStorage {
		if !o.Config.SingleMode {
			go o.runLeaderElection("arango-storage-operator", constants.LabelRole, o.onStartStorage, o.Dependencies.StorageProbe)
		} else {
			go o.runWithoutLeaderElection("arango-storage-operator", constants.LabelRole, o.onStartStorage, o.Dependencies.StorageProbe)
		}
	}
	if o.Config.EnableBackup {
		if !o.Config.SingleMode {
			go o.runLeaderElection("arango-backup-operator", constants.BackupLabelRole, o.onStartBackup, o.Dependencies.BackupProbe)
		} else {
			go o.runWithoutLeaderElection("arango-backup-operator", constants.BackupLabelRole, o.onStartBackup, o.Dependencies.BackupProbe)
		}
	}
	if o.Config.EnableApps {
		if !o.Config.SingleMode {
			go o.runLeaderElection("arango-apps-operator", constants.AppsLabelRole, o.onStartApps, o.Dependencies.AppsProbe)
		} else {
			go o.runWithoutLeaderElection("arango-apps-operator", constants.AppsLabelRole, o.onStartApps, o.Dependencies.AppsProbe)
		}
	}
	if o.Config.EnableK2KClusterSync {
		// Nothing to do
		o.log.Warn("K2K Cluster sync is permanently disabled")
	}

	ctx := util.CreateSignalContext(context.Background())
	<-ctx.Done()
	o.log.Info("Got interrupt signal, running shutdown handler in %s...", o.Config.ShutdownDelay)
	time.Sleep(o.Config.ShutdownDelay)
	o.handleShutdown()
}

func (o *Operator) handleShutdown() {
	o.log.Info("Waiting for deployments termination...")
	shutdownCh := make(chan struct{})
	go func() {
		for {
			if len(o.deployments) == 0 {
				break
			}
			time.Sleep(time.Second)
		}
		shutdownCh <- struct{}{}
	}()
	select {
	case <-shutdownCh:
		o.log.Info("All deployments terminated, exiting.")
		return
	case <-timer.After(o.Config.ShutdownTimeout):
		o.log.Info("Timeout reached before all deployments terminated, exiting.")
		return
	}
}

// onStartDeployment starts the deployment operator and run till given channel is closed.
func (o *Operator) onStartDeployment(stop <-chan struct{}) {
	checkFn := func() error {
		_, err := o.Client.Arango().DatabaseV1().ArangoDeployments(o.Namespace).List(context.Background(), meta.ListOptions{})
		return err
	}
	o.waitForCRD(depldef.ArangoDeploymentCRDName, checkFn)
	o.runDeployments(stop)
}

// onStartDeploymentReplication starts the deployment replication operator and run till given channel is closed.
func (o *Operator) onStartDeploymentReplication(stop <-chan struct{}) {
	checkFn := func() error {
		_, err := o.Client.Arango().DatabaseV1().ArangoDeployments(o.Namespace).List(context.Background(), meta.ListOptions{})
		return err
	}
	o.waitForCRD(repldef.ArangoDeploymentReplicationCRDName, checkFn)
	o.runDeploymentReplications(stop)
}

// onStartStorage starts the storage operator and run till given channel is closed.
func (o *Operator) onStartStorage(stop <-chan struct{}) {
	o.waitForCRD(lsapi.ArangoLocalStorageCRDName, nil)
	o.runLocalStorages(stop)
}

// onStartBackup starts the operator and run till given channel is closed.
func (o *Operator) onStartBackup(stop <-chan struct{}) {
	o.onStartOperatorV2(backupOperator, stop)
}

// onStartApps starts the operator and run till given channel is closed.
func (o *Operator) onStartApps(stop <-chan struct{}) {
	o.onStartOperatorV2(appsOperator, stop)
}

// onStartOperatorV2 run the operatorV2 type
func (o *Operator) onStartOperatorV2(operatorType operatorV2type, stop <-chan struct{}) {
	operatorName := fmt.Sprintf("arangodb-%s-operator", operatorType)
	operator := operatorV2.NewOperator(operatorName, o.Namespace, o.OperatorImage)

	util.Rand().Seed(time.Now().Unix())

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

	arangoInformer := arangoInformer.NewSharedInformerFactoryWithOptions(arangoClientSet, 10*time.Second, arangoInformer.WithNamespace(o.Namespace))

	switch operatorType {
	case appsOperator:
		checkFn := func() error {
			_, err := o.Client.Arango().AppsV1().ArangoJobs(o.Namespace).List(context.Background(), meta.ListOptions{})
			return err
		}
		o.waitForCRD(apps.ArangoJobCRDName, checkFn)

		if err = job.RegisterInformer(operator, eventRecorder, arangoClientSet, kubeClientSet, arangoInformer); err != nil {
			panic(err)
		}
	case backupOperator:
		checkFn := func() error {
			_, err := o.Client.Arango().BackupV1().ArangoBackups(o.Namespace).List(context.Background(), meta.ListOptions{})
			return err
		}
		o.waitForCRD(backupdef.ArangoBackupCRDName, checkFn)

		if err = backup.RegisterInformer(operator, eventRecorder, arangoClientSet, kubeClientSet, arangoInformer); err != nil {
			panic(err)
		}

		checkFn = func() error {
			_, err := o.Client.Arango().BackupV1().ArangoBackupPolicies(o.Namespace).List(context.Background(), meta.ListOptions{})
			return err
		}
		o.waitForCRD(backupdef.ArangoBackupPolicyCRDName, checkFn)

		if err = policy.RegisterInformer(operator, eventRecorder, arangoClientSet, kubeClientSet, arangoInformer); err != nil {
			panic(err)
		}
	}

	if err = operator.RegisterStarter(arangoInformer); err != nil {
		panic(err)
	}

	prometheus.MustRegister(operator)

	operator.Start(8, stop)
	o.Dependencies.BackupProbe.SetReady()

	<-stop
}

func (o *Operator) WrapLogger(in *zerolog.Event) *zerolog.Event {
	return in.Str("namespace", o.Namespace)
}
