//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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
	"sort"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	appKey                           = "app"
	roleKey                          = "role"
	appDeploymentOperator            = "arango-deployment-operator"
	appDeploymentReplicationOperator = "arango-deployment-replication-operator"
	appStorageOperator               = "arango-storage-operator"
	roleLeader                       = "leader"
)

// FindOtherOperators looks up references to other operators in the same Kubernetes cluster.
func (o *Operator) FindOtherOperators() []server.OperatorReference {
	if o.Scope.IsNamespaced() {
		// In namespaced scope nothing to do
		return []server.OperatorReference{}
	}

	log := o.log
	var result []server.OperatorReference
	namespaces, err := o.Dependencies.KubeCli.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list namespaces")
	} else {
		for _, ns := range namespaces.Items {
			if ns.Name != o.Config.Namespace {
				log.Debug().Str("namespace", ns.Name).Msg("inspecting namespace for operators")
				refs := o.findOtherOperatorsInNamespace(log, ns.Name, func(server.OperatorType) bool { return true })
				result = append(result, refs...)
			} else {
				log.Debug().Str("namespace", ns.Name).Msg("skip inspecting my own namespace for operators")
			}
		}
	}
	refs := o.findOtherOperatorsInNamespace(log, o.Config.Namespace, func(oType server.OperatorType) bool {
		// Exclude those operators that I provide myself.
		switch oType {
		case server.OperatorTypeDeployment:
			return !o.Dependencies.DeploymentProbe.IsReady()
		case server.OperatorTypeDeploymentReplication:
			return !o.Dependencies.DeploymentReplicationProbe.IsReady()
		case server.OperatorTypeStorage:
			return !o.Dependencies.StorageProbe.IsReady()
		default:
			return true
		}
	})
	result = append(result, refs...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Namespace == result[j].Namespace {
			return result[i].Type < result[j].Type
		}
		return result[i].Namespace < result[j].Namespace
	})

	return result
}

// findOtherOperatorsInNamespace looks up references to other operators in the given namespace.
func (o *Operator) findOtherOperatorsInNamespace(log zerolog.Logger, namespace string, typePred func(server.OperatorType) bool) []server.OperatorReference {
	log = log.With().Str("namespace", namespace).Logger()
	var result []server.OperatorReference
	services, err := o.Dependencies.KubeCli.CoreV1().Services(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Debug().Err(err).Msg("Failed to list services")
		return nil
	}
	nodeFetcher := func() (v1.NodeList, error) {
		if o.Scope.IsNamespaced() {
			return v1.NodeList{}, nil
		}
		result, err := o.Dependencies.KubeCli.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return v1.NodeList{}, errors.WithStack(err)
		}
		return *result, nil
	}
	for _, svc := range services.Items {
		// Filter out unwanted services
		selector := svc.Spec.Selector
		if selector[roleKey] != roleLeader {
			log.Debug().Str("service", svc.Name).Msg("Service has no leader role selector")
			continue
		}
		var oType server.OperatorType
		switch selector[appKey] {
		case appDeploymentOperator:
			oType = server.OperatorTypeDeployment
		case appDeploymentReplicationOperator:
			oType = server.OperatorTypeDeploymentReplication
		case appStorageOperator:
			oType = server.OperatorTypeStorage
		default:
			log.Debug().Str("service", svc.Name).Msg("Service has no or invalid app selector")
			continue
		}
		if !typePred(oType) {
			continue
		}
		var url string
		switch svc.Spec.Type {
		case v1.ServiceTypeNodePort, v1.ServiceTypeLoadBalancer:
			if x, err := k8sutil.CreateServiceURL(svc, "https", nil, nodeFetcher); err == nil {
				url = x
			} else {
				log.Warn().Err(err).Str("service", svc.Name).Msg("Failed to create URL for service")
			}
		default:
			// No suitable service type
			continue
		}
		result = append(result, server.OperatorReference{
			Namespace: svc.GetNamespace(),
			URL:       url,
			Type:      oType,
		})
	}

	log.Debug().Msgf("Found %d operator services", len(result))
	return result
}
