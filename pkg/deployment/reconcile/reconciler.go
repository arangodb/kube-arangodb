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

package reconcile

import (
	"context"

	"github.com/rs/zerolog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/logging"
)

var reconcileLogger = logging.Global().RegisterAndGetLogger("deployment-reconcile", logging.Info)

// Reconciler is the service that takes care of bring the a deployment
// in line with its (changed) specification.
type Reconciler struct {
	namespace, name string
	log             logging.Logger
	planLogger      logging.Logger
	context         Context

	metrics Metrics
}

// NewReconciler creates a new reconciler with given context.
func NewReconciler(namespace, name string, context Context) *Reconciler {
	r := &Reconciler{
		context:   context,
		namespace: namespace,
		name:      name,
	}
	r.log = reconcileLogger.WrapObj(r)
	r.planLogger = r.log.Str("section", "plan")
	return r
}

func (r *Reconciler) WrapLogger(in *zerolog.Event) *zerolog.Event {
	return in.Str("namespace", r.namespace).Str("name", r.name)
}

// CheckDeployment checks for obviously broken things and fixes them immediately
func (r *Reconciler) CheckDeployment(ctx context.Context) error {
	spec := r.context.GetSpec()
	status := r.context.GetStatus()

	if spec.GetMode().HasCoordinators() {
		// Check if there are coordinators
		if status.Members.Coordinators.AllFailed() {
			r.log.Error("All coordinators failed - reset")
			for _, m := range status.Members.Coordinators {
				cache, ok := r.context.ACS().ClusterCache(m.ClusterID)
				if !ok {
					r.log.Warn("Cluster is not ready")
					continue
				}

				if err := cache.Client().Kubernetes().CoreV1().Secrets(cache.Namespace()).Delete(ctx, m.Pod.GetName(), meta.DeleteOptions{}); err != nil {
					r.log.Err(err).Error("Failed to delete secret")
				}
				m.Phase = api.MemberPhaseNone

				if err := r.context.UpdateMember(ctx, m); err != nil {
					r.log.Err(err).Error("Failed to update member")
				}
			}
		}
	}

	return nil
}
