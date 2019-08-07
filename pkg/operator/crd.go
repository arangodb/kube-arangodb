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
	deplapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	backapi "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1alpha"
	replapi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1alpha"
	lsapi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/crd"
)

// waitForCRD waits for the CustomResourceDefinition (created externally)
// to be ready.
func (o *Operator) waitForCRD(enableDeployment, enableDeploymentReplication, enableStorage, enableBackup bool) error {
	log := o.log

	if enableDeployment {
		log.Debug().Msg("Waiting for ArangoDeployment CRD to be ready")
		if err := crd.WaitCRDReady(o.KubeExtCli, deplapi.ArangoDeploymentCRDName); err != nil {
			return maskAny(err)
		}
	}

	if enableDeploymentReplication {
		log.Debug().Msg("Waiting for ArangoDeploymentReplication CRD to be ready")
		if err := crd.WaitCRDReady(o.KubeExtCli, replapi.ArangoDeploymentReplicationCRDName); err != nil {
			return maskAny(err)
		}
	}

	if enableStorage {
		log.Debug().Msg("Waiting for ArangoLocalStorage CRD to be ready")
		if err := crd.WaitCRDReady(o.KubeExtCli, lsapi.ArangoLocalStorageCRDName); err != nil {
			return maskAny(err)
		}
	}

	if enableBackup {
		log.Debug().Msg("Wait for ArangoBackup CRD to be ready")
		if err := crd.WaitCRDReady(o.KubeExtCli, backapi.ArangoBackupCRDName); err != nil {
			return maskAny(err)
		}
	}

	log.Debug().Msg("CRDs ready")

	return nil
}
