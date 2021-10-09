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

	"github.com/arangodb/kube-arangodb/pkg/apis/backup"
	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	"github.com/arangodb/kube-arangodb/pkg/apis/replication"
	lsapi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/crd"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// waitForCRD waits for the CustomResourceDefinition (created externally)
// to be ready.
func (o *Operator) waitForCRD(enableDeployment, enableDeploymentReplication, enableStorage, enableBackup bool) error {
	log := o.log

	if o.Scope.IsNamespaced() {
		if enableDeployment {
			log.Debug().Msg("Waiting for ArangoDeployment CRD to be ready")
			if err := crd.WaitReady(func() error {
				_, err := o.CRCli.DatabaseV1().ArangoDeployments(o.Namespace).List(context.Background(), meta.ListOptions{})
				return err
			}); err != nil {
				return errors.WithStack(err)
			}
		}

		if enableDeploymentReplication {
			log.Debug().Msg("Waiting for ArangoDeploymentReplication CRD to be ready")
			if err := crd.WaitReady(func() error {
				_, err := o.CRCli.ReplicationV1().ArangoDeploymentReplications(o.Namespace).List(context.Background(), meta.ListOptions{})
				return err
			}); err != nil {
				return errors.WithStack(err)
			}
		}

		if enableBackup {
			log.Debug().Msg("Wait for ArangoBackup CRD to be ready")
			if err := crd.WaitReady(func() error {
				_, err := o.CRCli.BackupV1().ArangoBackups(o.Namespace).List(context.Background(), meta.ListOptions{})
				return err
			}); err != nil {
				return errors.WithStack(err)
			}
		}
	} else {
		if enableDeployment {
			log.Debug().Msg("Waiting for ArangoDeployment CRD to be ready")
			if err := crd.WaitCRDReady(o.KubeExtCli, deployment.ArangoDeploymentCRDName); err != nil {
				return errors.WithStack(err)
			}
		}

		if enableDeploymentReplication {
			log.Debug().Msg("Waiting for ArangoDeploymentReplication CRD to be ready")
			if err := crd.WaitCRDReady(o.KubeExtCli, replication.ArangoDeploymentReplicationCRDName); err != nil {
				return errors.WithStack(err)
			}
		}

		if enableStorage {
			log.Debug().Msg("Waiting for ArangoLocalStorage CRD to be ready")
			if err := crd.WaitCRDReady(o.KubeExtCli, lsapi.ArangoLocalStorageCRDName); err != nil {
				return errors.WithStack(err)
			}
		}

		if enableBackup {
			log.Debug().Msg("Wait for ArangoBackup CRD to be ready")
			if err := crd.WaitCRDReady(o.KubeExtCli, backup.ArangoBackupCRDName); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	log.Debug().Msg("CRDs ready")

	return nil
}
