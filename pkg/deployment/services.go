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

package deployment

import (
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// createServices creates all services needed to service the given deployment
func (d *Deployment) createServices(apiObject *api.ArangoDeployment) error {
	log := d.deps.Log
	kubecli := d.deps.KubeCli
	owner := apiObject.AsOwner()

	log.Debug().Msg("creating services...")

	if _, err := k8sutil.CreateHeadlessService(kubecli, apiObject, owner); err != nil {
		log.Debug().Err(err).Msg("Failed to create headless service")
		return maskAny(err)
	}
	single := apiObject.Spec.GetMode().HasSingleServers()
	if svcName, err := k8sutil.CreateDatabaseClientService(kubecli, apiObject, single, owner); err != nil {
		log.Debug().Err(err).Msg("Failed to create database client service")
		return maskAny(err)
	} else {
		d.status.ServiceName = svcName
		if err := d.updateCRStatus(); err != nil {
			return maskAny(err)
		}
	}
	if apiObject.Spec.Sync.IsEnabled() {
		if svcName, err := k8sutil.CreateSyncMasterClientService(kubecli, apiObject, owner); err != nil {
			log.Debug().Err(err).Msg("Failed to create syncmaster client service")
			return maskAny(err)
		} else {
			d.status.ServiceName = svcName
			if err := d.updateCRStatus(); err != nil {
				return maskAny(err)
			}
		}
	}
	return nil
}
