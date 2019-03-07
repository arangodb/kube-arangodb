package resources

import (
	"fmt"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func min(a int, b int) int {
	if a > b {
		return b
	}
	return a
}

// EnsurePDBs ensures Pod Distruption Budgets for different server groups in Cluster mode
func (r *Resources) EnsurePDBs() error {

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
		if err := r.ensurePDBForGroup(api.ServerGroupAgents, minAgents); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(api.ServerGroupDBServers, minDBServers); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(api.ServerGroupCoordinators, minCoordinators); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(api.ServerGroupSyncMasters, minSyncMaster); err != nil {
			return err
		}
		if err := r.ensurePDBForGroup(api.ServerGroupSyncWorkers, minSyncWorker); err != nil {
			return err
		}
	}

	return nil
}

func pdbNameForGroup(depl string, group api.ServerGroup) string {
	return fmt.Sprintf("%s-%s-pdb", depl, group.AsRole())
}

func newPDB(minAvail int, deplname string, group api.ServerGroup, owner metav1.OwnerReference) *policyv1beta1.PodDisruptionBudget {
	return &policyv1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:            pdbNameForGroup(deplname, group),
			OwnerReferences: []metav1.OwnerReference{owner},
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MinAvailable: newFromInt(minAvail),
			Selector: &metav1.LabelSelector{
				MatchLabels: k8sutil.LabelsForDeployment(deplname, group.AsRole()),
			},
		},
	}
}

// ensurePDBForGroup ensure pdb for a specific server group, if wantMinAvail is zero, the PDB is removed and not recreated
func (r *Resources) ensurePDBForGroup(group api.ServerGroup, wantedMinAvail int) error {
	deplname := r.context.GetAPIObject().GetName()
	pdbname := pdbNameForGroup(deplname, group)
	pdbcli := r.context.GetKubeCli().Policy().PodDisruptionBudgets(r.context.GetNamespace())
	log := r.log.With().Str("group", group.AsRole()).Logger()

	pdb, err := pdbcli.Get(pdbname, metav1.GetOptions{})
	if k8sutil.IsNotFound(err) {
		if wantedMinAvail != 0 {
			// No PDB found - create new
			pdb := newPDB(wantedMinAvail, deplname, group, r.context.GetAPIObject().AsOwner())
			log.Debug().Msg("Creating new PDB")
			if _, err := pdbcli.Create(pdb); err != nil {
				log.Error().Err(err).Msg("failed to create PDB")
				return maskAny(err)
			}
		}
		return nil
	} else if err == nil {
		// PDB is there
		// Update for PDBs is forbidden, thus one has to delete it and then create it again
		// Otherwise delete it if wantedMinAvail is zero
		if pdb.Spec.MinAvailable.IntValue() != wantedMinAvail || wantedMinAvail == 0 {
			log.Debug().Int("wanted-min-avail", wantedMinAvail).
				Int("current-min-avail", pdb.Spec.MinAvailable.IntValue()).
				Msg("Recreating PDB")
			pdb.Spec.MinAvailable = newFromInt(wantedMinAvail)

			// Update the PDB
			if err := pdbcli.Delete(pdbname, &metav1.DeleteOptions{}); err != nil {
				log.Error().Err(err).Msg("PDB deletion failed")
				maskAny(err)
			}
			// Create next reconcile loop, this should not be critical
			// 	but only if wantedMinAvail is nonzero
			return nil
		}
	}
	return maskAny(err)
}

func newFromInt(v int) *intstr.IntOrString {
	return &intstr.IntOrString{Type: intstr.Int, IntVal: int32(v)}
}
