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

package deployment

import (
	"context"
	"sort"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/apis/shared"
	memberState "github.com/arangodb/kube-arangodb/pkg/deployment/member"
	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// Name returns the name of the deployment.
func (d *Deployment) Name() string {
	return d.currentObject.Name
}

// Namespace returns the namespace that contains the deployment.
func (d *Deployment) Namespace() string {
	return d.currentObject.Namespace
}

// GetMode returns the mode of the deployment.
func (d *Deployment) GetMode() api.DeploymentMode {
	return d.GetSpec().GetMode()
}

// Environment returns the environment used in the deployment.
func (d *Deployment) Environment() api.Environment {
	return d.GetSpec().GetEnvironment()
}

// StateColor determinates the state of the deployment in color codes.
func (d *Deployment) StateColor() server.StateColor {
	allGood := true
	deploymentAvailable := true
	failed := false
	if d.PodCount() != d.ReadyPodCount() {
		allGood = false
	}
	if d.VolumeCount() != d.ReadyVolumeCount() {
		allGood = false
	}
	status := d.GetStatus()
	for _, m := range status.Members.AsList() {
		switch m.Member.Phase {
		case api.MemberPhaseFailed:
			failed = true
		case api.MemberPhaseCreated:
			// Should be ok now
		default:
			// Something is going on
			allGood = true
		}
	}
	if failed {
		return server.StateRed
	}
	if !deploymentAvailable {
		return server.StateOrange
	}
	if !allGood {
		return server.StateYellow
	}
	return server.StateGreen
}

// PodCount returns the number of pods for the deployment
func (d *Deployment) PodCount() int {
	status := d.GetStatus()
	return len(status.Members.PodNames())
}

// ReadyPodCount returns the number of pods for the deployment that are in ready state
func (d *Deployment) ReadyPodCount() int {
	count := 0
	status := d.GetStatus()
	for _, e := range status.Members.AsList() {
		if e.Member.Pod.GetName() == "" {
			continue
		}
		if e.Member.Conditions.IsTrue(api.ConditionTypeReady) {
			count++
		}
	}
	return count
}

// VolumeCount returns the number of volumes for the deployment
func (d *Deployment) VolumeCount() int {
	count := 0
	status := d.GetStatus()
	for _, e := range status.Members.AsList() {
		if e.Member.PersistentVolumeClaim != nil {
			count++
		}
	}
	return count
}

// ReadyVolumeCount returns the number of volumes for the deployment that are in ready state
func (d *Deployment) ReadyVolumeCount() int {
	count := 0
	status := d.GetStatus()
	pvcs, _ := d.GetOwnedPVCs() // Ignore errors on purpose
	for _, e := range status.Members.AsList() {
		if e.Member.PersistentVolumeClaim.GetName() == "" {
			continue
		}
		// Find status
		for _, pvc := range pvcs {
			if pvc.Name == e.Member.PersistentVolumeClaim.GetName() {
				if pvc.Status.Phase == core.ClaimBound {
					count++
				}
			}
		}
	}
	return count
}

// StorageClasses returns the names of the StorageClasses used by this deployment.
func (d *Deployment) StorageClasses() []string {
	scNames := make(map[string]struct{})
	spec := d.GetSpec()
	mode := spec.GetMode()
	if mode.HasAgents() {
		scNames[spec.Agents.GetStorageClassName()] = struct{}{}
	}
	if mode.HasDBServers() {
		scNames[spec.DBServers.GetStorageClassName()] = struct{}{}
	}
	if mode.HasSingleServers() {
		scNames[spec.Single.GetStorageClassName()] = struct{}{}
	}
	result := make([]string, 0, len(scNames))
	for k := range scNames {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// DatabaseURL returns an URL to reach the database from outside the Kubernetes cluster
// Empty string means that the database is not reachable outside the Kubernetes cluster.
func (d *Deployment) DatabaseURL() string {
	eaSvcName := k8sutil.CreateDatabaseExternalAccessServiceName(d.Name())
	svc, err := d.acs.CurrentClusterCache().Service().V1().Read().Get(context.Background(), eaSvcName, meta.GetOptions{})
	if err != nil {
		return ""
	}
	scheme := "https"
	if !d.GetSpec().IsSecure() {
		scheme = "http"
	}
	nodeFetcher := func() ([]*core.Node, error) {
		if n, err := d.acs.CurrentClusterCache().Node().V1(); err != nil {
			return nil, nil
		} else {
			return n.ListSimple(), nil
		}
	}
	portPredicate := func(p core.ServicePort) bool {
		return p.TargetPort.IntValue() == shared.ArangoPort
	}
	url, err := k8sutil.CreateServiceURL(*svc, scheme, portPredicate, nodeFetcher)
	if err != nil {
		return ""
	}
	return url
}

// DatabaseVersion returns the version used by the deployment
// Returns versionNumber, licenseType
func (d *Deployment) DatabaseVersion() (string, string) {
	status := d.GetStatus()
	if current := status.CurrentImage; current != nil {
		return string(current.ArangoDBVersion), memberState.GetImageLicense(status.CurrentImage)
	}
	return "", ""
}

// Members returns all members of the deployment by role.
func (d *Deployment) Members() map[api.ServerGroup][]server.Member {
	result := make(map[api.ServerGroup][]server.Member)
	status := d.GetStatus()

	for _, group := range api.AllServerGroups {
		list := status.Members.MembersOfGroup(group)
		members := make([]server.Member, len(list))
		for i, m := range list {
			members[i] = member{
				d:     d,
				id:    m.ID,
				group: group,
			}
		}
		if len(members) > 0 {
			result[group] = members
		}
	}

	return result
}
