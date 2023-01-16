//
// DISCLAIMER
//
// Copyright 2019-2023 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	"context"

	coreosv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	deploymentApi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

func LabelsForExporterServiceMonitor(name string, obj deploymentApi.DeploymentSpec) map[string]string {
	base := LabelsForExporterServiceMonitorSelector(name)

	for k, v := range obj.Metrics.ServiceMonitor.GetLabels(map[string]string{
		"context": "metrics",
		"metrics": "prometheus",
	}) {
		base[k] = v
	}

	return base
}

func LabelsForExporterServiceMonitorSelector(name string) map[string]string {
	return map[string]string{
		k8sutil.LabelKeyArangoDeployment: name,
		k8sutil.LabelKeyApp:              k8sutil.AppName,
	}
}

func (r *Resources) makeEndpoint(isSecure bool) coreosv1.Endpoint {
	if isSecure {
		return coreosv1.Endpoint{
			Port:     "exporter",
			Interval: "10s",
			Scheme:   "https",
			TLSConfig: &coreosv1.TLSConfig{
				SafeTLSConfig: coreosv1.SafeTLSConfig{
					InsecureSkipVerify: true,
				},
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

		version := r.context.GetMembersState().State().Version.Version
		endpoint.Path = getArangoExporterInternalEndpoint(version)

		return coreosv1.ServiceMonitorSpec{
			JobLabel: "k8s-app",
			Endpoints: []coreosv1.Endpoint{
				endpoint,
			},
			Selector: meta.LabelSelector{
				MatchLabels: LabelsForExporterServiceMonitorSelector(r.context.GetName()),
			},
		}, nil
	default:
		return coreosv1.ServiceMonitorSpec{
			JobLabel: "k8s-app",
			Endpoints: []coreosv1.Endpoint{
				r.makeEndpoint(spec.IsSecure()),
			},
			Selector: meta.LabelSelector{
				MatchLabels: LabelsForExporterServiceMonitorSelector(deploymentName),
			},
		}, nil
	}
}

// EnsureServiceMonitor creates or updates a ServiceMonitor.
func (r *Resources) EnsureServiceMonitor(ctx context.Context, enabled bool) error {
	if !enabled {
		return nil
	}

	// Some preparations:
	log := r.log.Str("section", "service-monitor")
	apiObject := r.context.GetAPIObject()
	deploymentName := apiObject.GetName()
	ns := apiObject.GetNamespace()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()

	if !spec.Metrics.ServiceMonitor.IsEnabled() || !spec.Metrics.IsEnabled() {
		return nil
	}

	wantMetrics := spec.Metrics.IsEnabled()
	serviceMonitorName := k8sutil.CreateExporterClientServiceName(deploymentName)

	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		log.Error("Cannot get a monitoring client.")
		return errors.Newf("Client not initialised")
	}

	mClient := client.Monitoring()

	// Check if ServiceMonitor already exists
	serviceMonitors := mClient.MonitoringV1().ServiceMonitors(ns)
	ctxChild, cancel := globals.GetGlobalTimeouts().Kubernetes().WithTimeout(ctx)
	defer cancel()
	servMon, err := serviceMonitors.Get(ctxChild, serviceMonitorName, meta.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			if !wantMetrics {
				return nil
			}

			spec, err := r.serviceMonitorSpec()
			if err != nil {
				return err
			}

			// Need to create one:
			smon := &coreosv1.ServiceMonitor{
				ObjectMeta: meta.ObjectMeta{
					Name:            serviceMonitorName,
					Labels:          LabelsForExporterServiceMonitor(r.context.GetName(), r.context.GetSpec()),
					OwnerReferences: []meta.OwnerReference{owner},
				},
				Spec: spec,
			}

			err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				_, err := serviceMonitors.Create(ctxChild, smon, meta.CreateOptions{})
				return err
			})
			if err != nil {
				log.Err(err).Error("Failed to create ServiceMonitor %s", serviceMonitorName)
				return errors.WithStack(err)
			}
			log.Debug("ServiceMonitor %s successfully created.", serviceMonitorName)
			return nil
		} else {
			log.Err(err).Error("Failed to get ServiceMonitor %s", serviceMonitorName)
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
		log.Debug("Found unneeded ServiceMonitor %s, but not owned by us, will not touch it", serviceMonitorName)
		return nil
	}
	if wantMetrics {
		log.Debug("ServiceMonitor %s already found, ensuring it is fine.",
			serviceMonitorName)

		spec, err := r.serviceMonitorSpec()
		if err != nil {
			return err
		}

		if equality.Semantic.DeepDerivative(spec, servMon.Spec) {
			log.Debug("ServiceMonitor %s already found and up to date.",
				serviceMonitorName)
			return nil
		}

		servMon.Spec = spec

		err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			_, err := serviceMonitors.Update(ctxChild, servMon, meta.UpdateOptions{})
			return err
		})
		if err != nil {
			return err
		}

		return nil
	}
	// Need to get rid of the ServiceMonitor:
	err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
		return serviceMonitors.Delete(ctxChild, serviceMonitorName, meta.DeleteOptions{})
	})
	if err == nil {
		log.Debug("Deleted ServiceMonitor %s", serviceMonitorName)
		return nil
	}
	log.Err(err).Error("Could not delete ServiceMonitor %s.", serviceMonitorName)
	return errors.WithStack(err)
}
