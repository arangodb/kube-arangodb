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

package storage

import (
	"context"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/storage/provisioner"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
)

// ensureProvisionerService ensures that a service is created for accessing the
// provisioners.
func (ls *LocalStorage) ensureProvisionerService(apiObject *api.ArangoLocalStorage) error {
	labels := k8sutil.LabelsForLocalStorage(apiObject.GetName(), roleProvisioner)
	svc := &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:   apiObject.GetName(),
			Labels: labels,
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				core.ServicePort{
					Name:     "provisioner",
					Protocol: core.ProtocolTCP,
					Port:     provisioner.DefaultPort,
				},
			},
			Selector: labels,
		},
	}
	svc.SetOwnerReferences(append(svc.GetOwnerReferences(), apiObject.AsOwner()))
	ns := ls.config.Namespace
	if _, err := ls.deps.Client.Kubernetes().CoreV1().Services(ns).Create(context.Background(), svc, meta.CreateOptions{}); err != nil && !kerrors.IsAlreadyExists(err) {
		return errors.WithStack(err)
	}
	return nil
}
