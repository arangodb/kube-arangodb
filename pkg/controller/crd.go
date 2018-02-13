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

package controller

import (
	"fmt"

	"github.com/pkg/errors"

	api "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	"github.com/arangodb/k8s-operator/pkg/util/k8sutil"
)

// initResourceIfNeeded initializes the custom resource definition when
// instructed to do so by the config.
func (c *Controller) initResourceIfNeeded() error {
	if c.Config.CreateCRD {
		if err := c.initCRD(); err != nil {
			return maskAny(fmt.Errorf("Failed to initialize Custom Resource Definition: %v", err))
		}
	}
	return nil
}

// initCRD creates the CustomResourceDefinition and waits for it to be ready.
func (c *Controller) initCRD() error {
	log := c.Dependencies.Log

	log.Info().Msg("Calling CreateCRD")
	if err := k8sutil.CreateCRD(c.KubeExtCli, api.ArangoDeploymentCRDName, api.ArangoDeploymentResourceKind, api.ArangoDeploymentResourcePlural, "arangodb"); err != nil {
		return maskAny(errors.Wrapf(err, "failed to create CRD: %v", err))
	}
	log.Info().Msg("Waiting for CRD ready")
	if err := k8sutil.WaitCRDReady(c.KubeExtCli, api.ArangoDeploymentCRDName); err != nil {
		return maskAny(err)
	}
	log.Info().Msg("CRD is ready")
	return nil
}
