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

package resources

import (
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// EnsureServices creates all services needed to service the deployment
func (r *Resources) EnsureServices() error {
	log := r.log
	kubecli := r.context.GetKubeCli()
	apiObject := r.context.GetAPIObject()
	owner := apiObject.AsOwner()
	spec := r.context.GetSpec()

	log.Debug().Msg("creating services...")

	if _, err := k8sutil.CreateHeadlessService(kubecli, apiObject, owner); err != nil {
		log.Debug().Err(err).Msg("Failed to create headless service")
		return maskAny(err)
	}
	single := spec.GetMode().HasSingleServers()
	svcName, err := k8sutil.CreateDatabaseClientService(kubecli, apiObject, single, owner)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to create database client service")
		return maskAny(err)
	}
	status := r.context.GetStatus()
	if status.ServiceName != svcName {
		status.ServiceName = svcName
		if err := r.context.UpdateStatus(status); err != nil {
			return maskAny(err)
		}
	}

	if spec.Sync.IsEnabled() {
		svcName, err := k8sutil.CreateSyncMasterClientService(kubecli, apiObject, owner)
		if err != nil {
			log.Debug().Err(err).Msg("Failed to create syncmaster client service")
			return maskAny(err)
		}
		status := r.context.GetStatus()
		if status.SyncServiceName != svcName {
			status.SyncServiceName = svcName
			if err := r.context.UpdateStatus(status); err != nil {
				return maskAny(err)
			}
		}
	}
	return nil
}
