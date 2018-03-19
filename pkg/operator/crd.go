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
	"fmt"

	"github.com/pkg/errors"

	deplapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	lsapi "github.com/arangodb/kube-arangodb/pkg/apis/storage/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/crd"
)

// initResourceIfNeeded initializes the custom resource definition when
// instructed to do so by the config.
func (o *Operator) initResourceIfNeeded(enableDeployment, enableStorage bool) error {
	if o.Config.CreateCRD {
		if err := o.initCRD(enableDeployment, enableStorage); err != nil {
			return maskAny(fmt.Errorf("Failed to initialize Custom Resource Definition: %v", err))
		}
	}
	return nil
}

// initCRD creates the CustomResourceDefinition and waits for it to be ready.
func (o *Operator) initCRD(enableDeployment, enableStorage bool) error {
	log := o.Dependencies.Log

	if enableDeployment {
		log.Debug().Msg("Creating ArangoDeployment CRD")
		if err := crd.CreateCRD(o.KubeExtCli, deplapi.SchemeGroupVersion, deplapi.ArangoDeploymentCRDName, deplapi.ArangoDeploymentResourceKind, deplapi.ArangoDeploymentResourcePlural, deplapi.ArangoDeploymentShortNames...); err != nil {
			return maskAny(errors.Wrapf(err, "failed to create CRD: %v", err))
		}
		log.Debug().Msg("Waiting for ArangoDeployment CRD to be ready")
		if err := crd.WaitCRDReady(o.KubeExtCli, deplapi.ArangoDeploymentCRDName); err != nil {
			return maskAny(err)
		}
	}

	if enableStorage {
		log.Debug().Msg("Creating ArangoLocalStorage CRD")
		if err := crd.CreateCRD(o.KubeExtCli, lsapi.SchemeGroupVersion, lsapi.ArangoLocalStorageCRDName, lsapi.ArangoLocalStorageResourceKind, lsapi.ArangoLocalStorageResourcePlural, lsapi.ArangoLocalStorageShortNames...); err != nil {
			return maskAny(errors.Wrapf(err, "failed to create CRD: %v", err))
		}
		log.Debug().Msg("Waiting for ArangoLocalStorage CRD to be ready")
		if err := crd.WaitCRDReady(o.KubeExtCli, lsapi.ArangoLocalStorageCRDName); err != nil {
			return maskAny(err)
		}
	}

	return nil
}
