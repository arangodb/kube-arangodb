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
	"sync"
	"time"

	"github.com/arangodb/go-driver"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangoclustersynchronization"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangotask"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/endpoints"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/node"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/persistentvolumeclaim"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/pod"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/service"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/serviceaccount"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/servicemonitor"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/throttle"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	inspectorLoadersList inspectorLoaders
	inspectorLoadersLock sync.Mutex
)

func requireRegisterInspectorLoader(i inspectorLoader) {
	if !registerInspectorLoader(i) {
		panic("Unable to register inspector loader")
	}
}

func registerInspectorLoader(i inspectorLoader) bool {
	inspectorLoadersLock.Lock()
	defer inspectorLoadersLock.Unlock()

	n := i.Name()

	if inspectorLoadersList.Get(n) != -1 {
		return false
	}

	inspectorLoadersList = append(inspectorLoadersList, i)

	return true
}

type inspectorLoaders []inspectorLoader

func (i inspectorLoaders) Get(name string) int {
	for id, k := range i {
		if k.Name() == name {
			return id
		}
	}

	return -1
}

type inspectorLoader interface {
	Name() string

	Component() throttle.Component

	Load(context context.Context, i *inspectorState)

	Verify(i *inspectorState) error

	Copy(from, to *inspectorState, override bool)
}

var _ inspector.Inspector = &inspectorState{}

func NewInspector(throttles throttle.Components, client kclient.Client, namespace, deploymentName string) inspector.Inspector {
	if throttles == nil {
		throttles = throttle.NewAlwaysThrottleComponents()
	}

	i := &inspectorState{
		namespace:      namespace,
		deploymentName: deploymentName,
		client:         client,
		throttles:      throttles,
		logger:         logging.GlobalLogger().MustGetLogger(logging.LoggerNameInspector),
	}

	return i
}

type inspectorStateDeploymentResult struct {
	depl *api.ArangoDeployment
	err  error
}

type inspectorState struct {
	lock sync.Mutex

	namespace      string
	deploymentName string

	deploymentResult *inspectorStateDeploymentResult

	client kclient.Client

	last time.Time

	logger zerolog.Logger

	pods                          *podsInspector
	secrets                       *secretsInspector
	persistentVolumeClaims        *persistentVolumeClaimsInspector
	services                      *servicesInspector
	serviceAccounts               *serviceAccountsInspector
	nodes                         *nodesInspector
	podDisruptionBudgets          *podDisruptionBudgetsInspector
	serviceMonitors               *serviceMonitorsInspector
	arangoMembers                 *arangoMembersInspector
	arangoTasks                   *arangoTasksInspector
	arangoClusterSynchronizations *arangoClusterSynchronizationsInspector
	endpoints                     *endpointsInspector

	throttles throttle.Components

	versionInfo driver.Version

	initialised bool
}

func (i *inspectorState) GetCurrentArangoDeployment() (*api.ArangoDeployment, error) {
	if i.deploymentResult == nil {
		return nil, errors.Newf("Deployment not initialised")
	}

	return i.deploymentResult.depl, i.deploymentResult.err
}

func (i *inspectorState) Endpoints() endpoints.Definition {
	return i.endpoints
}

func (i *inspectorState) Initialised() bool {
	if i == nil {
		return false
	}

	return i.initialised
}

func (i *inspectorState) Client() kclient.Client {
	return i.client
}

func (i *inspectorState) Namespace() string {
	return i.namespace
}

func (i *inspectorState) LastRefresh() time.Time {
	return i.last
}

func (i *inspectorState) Secret() secret.Definition {
	return i.secrets
}

func (i *inspectorState) PersistentVolumeClaim() persistentvolumeclaim.Definition {
	return i.persistentVolumeClaims
}

func (i *inspectorState) Service() service.Definition {
	return i.services
}

func (i *inspectorState) PodDisruptionBudget() poddisruptionbudget.Definition {
	return i.podDisruptionBudgets
}

func (i *inspectorState) ServiceMonitor() servicemonitor.Definition {
	return i.serviceMonitors
}

func (i *inspectorState) ServiceAccount() serviceaccount.Definition {
	return i.serviceAccounts
}

func (i *inspectorState) ArangoMember() arangomember.Definition {
	return i.arangoMembers
}

func (i *inspectorState) GetVersionInfo() driver.Version {
	return i.versionInfo
}

func (i *inspectorState) Node() node.Definition {
	return i.nodes
}

func (i *inspectorState) ArangoClusterSynchronization() arangoclustersynchronization.Definition {
	return i.arangoClusterSynchronizations
}

func (i *inspectorState) ArangoTask() arangotask.Definition {
	return i.arangoTasks
}

func (i *inspectorState) Refresh(ctx context.Context) error {
	return i.refresh(ctx, inspectorLoadersList...)
}

func (i *inspectorState) GetThrottles() throttle.Components {
	return i.throttles
}

func (i *inspectorState) Pod() pod.Definition {
	return i.pods
}

func (i *inspectorState) refresh(ctx context.Context, loaders ...inspectorLoader) error {
	return i.refreshInThreads(ctx, 15, loaders...)
}

func (i *inspectorState) refreshInThreads(ctx context.Context, threads int, loaders ...inspectorLoader) error {
	i.lock.Lock()
	defer i.lock.Unlock()

	var m sync.WaitGroup

	p, close := util.ParallelThread(threads)
	defer close()

	m.Add(len(loaders))

	n := i.copyCore()

	if v, err := n.client.Kubernetes().Discovery().ServerVersion(); err != nil {
		n.versionInfo = ""
	} else {
		n.versionInfo = driver.Version(strings.TrimPrefix(v.GitVersion, "v"))
	}

	start := time.Now()
	i.logger.Debug().Msg("Pre-inspector refresh start")
	d, err := i.client.Arango().DatabaseV1().ArangoDeployments(i.namespace).Get(context.Background(), i.deploymentName, meta.GetOptions{})
	n.deploymentResult = &inspectorStateDeploymentResult{
		depl: d,
		err:  err,
	}

	i.logger.Debug().Msg("Inspector refresh start")

	for id := range loaders {
		go func(id int) {
			defer m.Done()

			c := loaders[id].Component()

			t := n.throttles.Get(c)

			if !t.Throttle() {
				i.logger.Debug().Str("component", string(c)).Msg("Inspector refresh skipped")
				return
			}

			i.logger.Debug().Str("component", string(c)).Msg("Inspector refresh")

			defer func() {
				i.logger.Debug().Str("component", string(c)).Str("duration", time.Since(start).String()).Msg("Inspector done")
				t.Delay()
			}()

			<-p
			defer func() {
				p <- struct{}{}
			}()

			loaders[id].Load(ctx, n)
		}(id)
	}

	m.Wait()

	i.logger.Debug().Str("duration", time.Since(start).String()).Msg("Inspector refresh done")

	for id := range loaders {
		if err := loaders[id].Verify(n); err != nil {
			return err
		}
	}

	if err := n.validate(); err != nil {
		return err
	}

	for id := range loaders {
		loaders[id].Copy(n, i, true)
	}

	i.deploymentResult = n.deploymentResult

	i.throttles = n.throttles

	i.last = time.Now()
	i.initialised = true

	return nil
}

func (i *inspectorState) validate() error {
	if err := i.pods.validate(); err != nil {
		return err
	}

	if err := i.secrets.validate(); err != nil {
		return err
	}

	if err := i.serviceAccounts.validate(); err != nil {
		return err
	}

	if err := i.persistentVolumeClaims.validate(); err != nil {
		return err
	}

	if err := i.services.validate(); err != nil {
		return err
	}

	if err := i.nodes.validate(); err != nil {
		return err
	}

	if err := i.podDisruptionBudgets.validate(); err != nil {
		return err
	}

	if err := i.serviceMonitors.validate(); err != nil {
		return err
	}

	if err := i.arangoMembers.validate(); err != nil {
		return err
	}

	if err := i.arangoTasks.validate(); err != nil {
		return err
	}

	if err := i.arangoClusterSynchronizations.validate(); err != nil {
		return err
	}

	if err := i.endpoints.validate(); err != nil {
		return err
	}

	return nil
}

func (i *inspectorState) copyCore() *inspectorState {
	return &inspectorState{
		namespace:                     i.namespace,
		deploymentName:                i.deploymentName,
		client:                        i.client,
		pods:                          i.pods,
		secrets:                       i.secrets,
		persistentVolumeClaims:        i.persistentVolumeClaims,
		services:                      i.services,
		serviceAccounts:               i.serviceAccounts,
		nodes:                         i.nodes,
		podDisruptionBudgets:          i.podDisruptionBudgets,
		serviceMonitors:               i.serviceMonitors,
		arangoMembers:                 i.arangoMembers,
		arangoTasks:                   i.arangoTasks,
		arangoClusterSynchronizations: i.arangoClusterSynchronizations,
		throttles:                     i.throttles.Copy(),
		versionInfo:                   i.versionInfo,
		endpoints:                     i.endpoints,
		deploymentResult:              i.deploymentResult,
		logger:                        i.logger,
	}
}
