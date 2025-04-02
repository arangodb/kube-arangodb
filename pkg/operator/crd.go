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

package operator

import (
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/crd"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/access"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

// waitForCRD waits for the CustomResourceDefinition (created externally) to be ready.
func (o *Operator) waitForCRD(crdName string, checkFn func() error) {
	log := o.log.Str("crd", crdName)
	log.Debug("Waiting for CRD to be ready")

	for {
		var err error = nil
		if !access.VerifyAccess(shutdown.Context(), o.Dependencies.Client, access.GVR(constants.CustomResourceDefinitionGRv1(), crdName, access.Get)) {
			log.Debug("Check by the CheckFun")
			if checkFn != nil {
				err = crd.WaitReady(checkFn)
			}
		} else {
			log.Debug("Check by tue Cluster Access")
			err = crd.WaitCRDReady(o.Client.KubernetesExtensions(), crdName)
		}

		if err == nil {
			break
		} else {
			log.Err(err).Error("Resource initialization failed")
			log.Info("Retrying in %s...", initRetryWaitTime)
			time.Sleep(initRetryWaitTime)
		}
	}

	log.Debug("CRD ready")
}
