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

package resources

import (
	"context"
	"fmt"
	"time"

	policyv1 "k8s.io/api/policy/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/timer"
)

func min(a int, b int) int {
	if a > b {
		return b
	}
	return a
}

// EnsurePDBs ensures Pod Disruption Budgets for different server groups in Cluster mode
func (r *Resources) EnsurePDBs(ctx context.Context) error {

	// Only in Cluster and Production Mode
	spec := r.context.GetSpec()
	if spec.IsProduction() && spec.GetMode().IsCluster() {

		// We want to lose at most one agent and dbserver.
		// Coordinators are not that critical. To keep the service available two should be enough
		minAgents := spec.GetServerGroupSpec(api.ServerGroupAgents).GetCount() - 1
		minDBServers := spec.GetServerGroupSpec(api.ServerGroupDBServers).GetCount() - 1
		minCoordinators := min(spec.GetServerGroupSpec(api.ServerGroupCoordinators).GetCount()-1, 2)

		// Setting those to zero triggers a remove of the PDB
		minSyncMaster := 0
		minSyncWorker := 0
		if spec.Sync.IsEnabled() {
			minSyncMaster = spec.GetServerGroupSpec(api.ServerGroupSyncMasters).GetCount() - 1
			minSyncWorker = spec.GetServerGroupSpec(api.ServerGroupSyncWorkers).GetCount() - 1
		}

		// Ensure all PDBs as calculated
		if err := r.ensurePDBForGroup(ctx, api.ServerGroupAgents, minAgents); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(ctx, api.ServerGroupDBServers, minDBServers); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(ctx, api.ServerGroupCoordinators, minCoordinators); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(ctx, api.ServerGroupSyncMasters, minSyncMaster); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(ctx, api.ServerGroupSyncWorkers, minSyncWorker); err != nil {
			return err
		}
	}

	return nil
}

func PDBNameForGroup(depl string, group api.ServerGroup) string {
	return fmt.Sprintf("%s-%s-pdb", depl, group.AsRole())
}

func newPDBV1Beta1(minAvail int, deplname string, group api.ServerGroup, owner meta.OwnerReference) *policyv1beta1.PodDisruptionBudget {
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: meta.ObjectMeta{
			Name:            PDBNameForGroup(deplname, group),
			OwnerReferences: []meta.OwnerReference{owner},
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: newFromInt(minAvail),
			Selector: &meta.LabelSelector{
				MatchLabels: k8sutil.LabelsForDeployment(deplname, group.AsRole()),
			},
		},
	}
}

func newPDBV1(minAvail int, deplname string, group api.ServerGroup, owner meta.OwnerReference) *policyv1.PodDisruptionBudget {
	return &policyv1.PodDisruptionBudget{
		ObjectMeta: meta.ObjectMeta{
			Name:            PDBNameForGroup(deplname, group),
			OwnerReferences: []meta.OwnerReference{owner},
		},
		Spec: policyv1.PodDisruptionBudgetSpec{
			MinAvailable: newFromInt(minAvail),
			Selector: &meta.LabelSelector{
				MatchLabels: k8sutil.LabelsForDeployment(deplname, group.AsRole()),
			},
		},
	}
}

// ensurePDBForGroup ensure pdb for a specific server group, if wantMinAvail is zero, the PDB is removed and not recreated
func (r *Resources) ensurePDBForGroup(ctx context.Context, group api.ServerGroup, wantedMinAvail int) error {
	deplName := r.context.GetAPIObject().GetName()
	pdbName := PDBNameForGroup(deplName, group)
	log := r.log.Str("section", "pdb").Str("group", group.AsRole())
	cache := r.context.ACS().CurrentClusterCache()
	pdbMod := cache.PodDisruptionBudgetsModInterface()

	for {
		var minAvailable *intstr.IntOrString
		var deletionTimestamp *meta.Time

		err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
			if inspector, err := cache.PodDisruptionBudget().V1(); err == nil {
				if pdb, err := inspector.Read().Get(ctxChild, pdbName, meta.GetOptions{}); err != nil {
					return err
				} else {
					minAvailable = pdb.Spec.MinAvailable
					deletionTimestamp = pdb.GetDeletionTimestamp()
				}
			} else if inspector, err := cache.PodDisruptionBudget().V1Beta1(); err == nil {
				if pdb, err := inspector.Read().Get(ctxChild, pdbName, meta.GetOptions{}); err != nil {
					return err
				} else {
					minAvailable = pdb.Spec.MinAvailable
					deletionTimestamp = pdb.GetDeletionTimestamp()
				}
			} else {
				return errors.WithStack(err)
			}

			return nil
		})

		if kerrors.IsNotFound(err) {
			if wantedMinAvail != 0 {
				// No PDB found - create new.
				log.Debug("Creating new PDB")
				err = globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
					var errInternal error

					if cache.PodDisruptionBudget().Version().IsV1() {
						pdb := newPDBV1(wantedMinAvail, deplName, group, r.context.GetAPIObject().AsOwner())
						_, errInternal = pdbMod.V1().Create(ctxChild, pdb, meta.CreateOptions{})
					} else {
						pdb := newPDBV1Beta1(wantedMinAvail, deplName, group, r.context.GetAPIObject().AsOwner())
						_, errInternal = pdbMod.V1Beta1().Create(ctxChild, pdb, meta.CreateOptions{})
					}

					return errInternal
				})

				if err != nil {
					log.Err(err).Error("failed to create PDB")
					return errors.WithStack(err)
				}
			}

			return nil
		} else if err != nil {
			// Some other error than not found.
			return errors.WithStack(err)
		}

		// PDB v1 or v1beta1 is here.
		if minAvailable.IntValue() == wantedMinAvail && wantedMinAvail != 0 {
			return nil
		}
		// Update for PDBs is forbidden, thus one has to delete it and then create it again
		// Otherwise delete it if wantedMinAvail is zero
		log.Int("wanted-min-avail", wantedMinAvail).
			Int("current-min-avail", minAvailable.IntValue()).
			Debug("Recreating PDB")

		// Trigger deletion only if not already deleted.
		if deletionTimestamp == nil {
			// Update the PDB.
			err := globals.GetGlobalTimeouts().Kubernetes().RunWithTimeout(ctx, func(ctxChild context.Context) error {
				if cache.PodDisruptionBudget().Version().IsV1() {
					return pdbMod.V1().Delete(ctxChild, pdbName, meta.DeleteOptions{})
				}

				return pdbMod.V1Beta1().Delete(ctxChild, pdbName, meta.DeleteOptions{})
			})
			if err != nil && !kerrors.IsNotFound(err) {
				log.Err(err).Error("PDB deletion failed")
				return errors.WithStack(err)
			}
		} else {
			log.Debug("PDB already deleted")
		}
		// Exit here if deletion was intended
		if wantedMinAvail == 0 {
			return nil
		}

		log.Debug("Retry loop for PDB")
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.After(time.Second):
		}
	}
}

func newFromInt(v int) *intstr.IntOrString {
	ret := &intstr.IntOrString{}
	*ret = intstr.FromInt(v)
	return ret
}
