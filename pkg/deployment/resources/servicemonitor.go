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
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"k8s.io/apimachinery/pkg/api/equality"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	coreosv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	clientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func LabelsForExporterServiceMonitor(deploymentName string) map[string]string {
	return map[string]string{
		k8sutil.LabelKeyArangoDeployment: deploymentName,
		k8sutil.LabelKeyApp:              k8sutil.AppName,
		"context":                        "metrics",
		"metrics":                        "prometheus",
	}
}

func LabelsForExporterServiceMonitorSelector(deploymentName string) map[string]string {
	return map[string]string{
		k8sutil.LabelKeyArangoDeployment: deploymentName,
		k8sutil.LabelKeyApp:              k8sutil.AppName,
	}
}

// EnsureMonitoringClient returns a client for looking at ServiceMonitors
// and keeps it in the Resources.
func (r *Resources) EnsureMonitoringClient() (*clientv1.MonitoringV1Client, error) {
	if r.monitoringClient != nil {
		return r.monitoringClient, nil
	}

	// Make a client:
	var restConfig *rest.Config
	restConfig, err := k8sutil.NewKubeConfig()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	mClient, err := clientv1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r.monitoringClient = mClient
	return mClient, nil
}

func (r *Resources) makeEndpoint(isSecure bool) coreosv1.Endpoint {
	if isSecure {
		return coreosv1.Endpoint{
			Port:     "exporter",
			Interval: "10s",
			Scheme:   "https",
			TLSConfig: &coreosv1.TLSConfig{
				InsecureSkipVerify: true,
			},
		}
	} else {
		return coreosv1.Endpoint{
			Port:     "exporter",
			Interval: "10s",
			Scheme:   "http",
		}
	}
}

func (r *Resources) serviceMonitorSpec() (coreosv1.ServiceMonitorSpec, error) {
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	spec := r.context.GetSpec()

	switch spec.Metrics.Mode.Get() {
	case deploymentApi.MetricsModeInternal:
		if spec.Metrics.Authentication.JWTTokenSecretName == nil {
			return coreosv1.ServiceMonitorSpec{}, apiErrors.NewNotFound(schema.GroupResource{Group: "v1/secret"}, "metrics-secret")
		}

		endpoint := r.makeEndpoint(spec.IsSecure())

		endpoint.BearerTokenSecret.Name = *spec.Metrics.Authentication.JWTTokenSecretName
		endpoint.BearerTokenSecret.Key = constants.SecretKeyToken
		endpoint.Path = k8sutil.ArangoExporterInternalEndpoint

		return coreosv1.ServiceMonitorSpec{
			JobLabel: "k8s-app",
			Endpoints: []coreosv1.Endpoint{
				endpoint,
			},
			Selector: metav1.LabelSelector{
				MatchLabels: LabelsForExporterServiceMonitorSelector(deploymentName),
			},
		}, nil
	default:
		return coreosv1.ServiceMonitorSpec{
			JobLabel: "k8s-app",
			Endpoints: []coreosv1.Endpoint{
				r.makeEndpoint(spec.IsSecure()),
			},
			Selector: metav1.LabelSelector{
				MatchLabels: LabelsForExporterServiceMonitorSelector(deploymentName),
			},
		}, nil
	}
}

// EnsureServiceMonitor creates or updates a ServiceMonitor.
func (r *Resources) EnsureServiceMonitor() error {
	// Some preparations:
	log := r.log
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()
	wantMetrics := spec.Metrics.IsEnabled()
	serviceMonitorName := k8sutil.CreateExporterClientServiceName(deploymentName)

	mClient, err := r.EnsureMonitoringClient()
	if err != nil {
		log.Error().Err(err).Msgf("Cannot get a monitoring client.")
		return errors.WithStack(err)
	}

	// Check if ServiceMonitor already exists
	serviceMonitors := mClient.ServiceMonitors(ns)
	servMon, err := serviceMonitors.Get(serviceMonitorName, metav1.GetOptions{})
	if err != nil {
		if k8sutil.IsNotFound(err) {
			if !wantMetrics {
				return nil
			}

			spec, err := r.serviceMonitorSpec()
			if err != nil {
				return err
			}

			// Need to create one:
			smon := &coreosv1.ServiceMonitor{
				ObjectMeta: metav1.ObjectMeta{
					Name:            serviceMonitorName,
					Labels:          LabelsForExporterServiceMonitor(deploymentName),
					OwnerReferences: []metav1.OwnerReference{owner},
				},
				Spec: spec,
			}
			smon, err = serviceMonitors.Create(smon)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to create ServiceMonitor %s", serviceMonitorName)
				return errors.WithStack(err)
			}
			log.Debug().Msgf("ServiceMonitor %s successfully created.", serviceMonitorName)
			return nil
		} else {
			log.Error().Err(err).Msgf("Failed to get ServiceMonitor %s", serviceMonitorName)
			return errors.WithStack(err)
		}
	}
	// Check if the service monitor is ours, otherwise we do not touch it:
	found := false
	for _, owner := range servMon.ObjectMeta.OwnerReferences {
		if owner.Kind == deployment.ArangoDeploymentResourceKind &&
			owner.Name == deploymentName {
			found = true
			break
		}
	}
	if !found {
		log.Debug().Msgf("Found unneeded ServiceMonitor %s, but not owned by us, will not touch it", serviceMonitorName)
		return nil
	}
	if wantMetrics {
		log.Debug().Msgf("ServiceMonitor %s already found, ensuring it is fine.",
			serviceMonitorName)

		spec, err := r.serviceMonitorSpec()
		if err != nil {
			return err
		}

		if equality.Semantic.DeepDerivative(spec, servMon.Spec) {
			log.Debug().Msgf("ServiceMonitor %s already found and up to date.",
				serviceMonitorName)
			return nil
		}

		servMon.Spec = spec

		_, err = serviceMonitors.Update(servMon)
		if err != nil {
			return err
		}

		return nil
	}
	// Need to get rid of the ServiceMonitor:
	err = serviceMonitors.Delete(serviceMonitorName, &metav1.DeleteOptions{})
	if err == nil {
		log.Debug().Msgf("Deleted ServiceMonitor %s", serviceMonitorName)
		return nil
	}
	log.Error().Err(err).Msgf("Could not delete ServiceMonitor %s.", serviceMonitorName)
	return errors.WithStack(err)
}
