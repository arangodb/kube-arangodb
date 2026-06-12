//
// DISCLAIMER
//
// Copyright 2016-2026 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/util/crd"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/access"
	inspectorConstants "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/constants"
)

// waitForCRD waits for the CustomResourceDefinition (created externally) to be ready.
func (o *Operator) waitForCRD(ctx context.Context, crdName string, checkFn func() error) {
	log := o.log.Str("crd", crdName)
	log.Debug("Waiting for CRD to be ready")

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	for {
		var err error = nil
		if !access.VerifyAccess(ctx, o.Dependencies.Client, access.GVR(inspectorConstants.CustomResourceDefinitionGRv1(), crdName, access.Get)) {
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
			select {
			case <-ctx.Done():
				log.Error("CRD %s not ready within timeout, proceeding", crdName)
				return
			case <-time.After(initRetryWaitTime):
				log.Info("Retrying...")
			}
		}
	}

	log.Debug("CRD ready")
}
