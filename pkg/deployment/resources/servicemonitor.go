//
// DISCLAIMER
//
// Copyright 2019 ArangoDB Inc, Cologne, Germany
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
// Author Max Neunhoeffer
//

package resources

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	coreosv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	clientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rs/zerolog"
)

// EnsureServiceMonitor creates or updates a ServiceMonitor.
func (f *Resources) EnsureServiceMonitor() error {
	// Some preparations:
	log := r.log
	start := time.Now()
	kubecli := r.context.GetKubeCli()
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()
	serviceMonitorName := deploymentName + "-exporter"

	// First get a client:
	var restConfig *rest.Config
	restConfig, err := k8sutil.InClusterConfig()
	if err != nil {
		return maskAny(err)
	}
	var mClient *clientv1.MonitoringV1Client
	mClient, err = clientv1.NewForConfig(restConfig)
	if err != nil {
		return maskAny(err)
	}

	// Check if ServiceMonitor already exists
	serviceMonitors := mClient.ServiceMonitors()
	var found *coreosv1.ServiceMonitor
	found, err = serviceMonitors.Get(serviceMonitorName, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(rr) {
			// Need to create one:
      smon := &coreosv1.ServiceMonitor{
			  ObjectMeta: metav1.ObjectMeta{
				  Name: serviceMonitorName,
					Labels: LabelsForExporterServiceMonitor(deploymentName),
				},
				Spec: coreosv1.ServiceMonitorSpec{
				  
		} else {
			log.Error().Err(err).Msgf("Failed to get ServiceMonitor %s", serviceMonitorName)
			return maskAny(err)
		}
	}

	log.Debug().Msgf("ServiceMonitor %s already found, no need to create.",
		serviceMonitorName)
	return nil
}
