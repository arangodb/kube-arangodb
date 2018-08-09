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

package deployment

import (
	"sort"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/server"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// Name returns the name of the deployment.
func (d *Deployment) Name() string {
	return d.apiObject.Name
}

// Namespace returns the namespace that contains the deployment.
func (d *Deployment) Namespace() string {
	return d.apiObject.Namespace
}

// Mode returns the mode of the deployment.
func (d *Deployment) Mode() api.DeploymentMode {
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
	status, _ := d.GetStatus()
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			switch m.Phase {
			case api.MemberPhaseFailed:
				failed = true
			case api.MemberPhaseCreated:
				// Should be ok now
			default:
				// Something is going on
				allGood = true
			}
		}
		return nil
	})
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
	count := 0
	status, _ := d.GetStatus()
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.PodName != "" {
				count++
			}
		}
		return nil
	})
	return count
}

// ReadyPodCount returns the number of pods for the deployment that are in ready state
func (d *Deployment) ReadyPodCount() int {
	count := 0
	status, _ := d.GetStatus()
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.PodName == "" {
				continue
			}
			if m.Conditions.IsTrue(api.ConditionTypeReady) {
				count++
			}
		}
		return nil
	})
	return count
}

// VolumeCount returns the number of volumes for the deployment
func (d *Deployment) VolumeCount() int {
	count := 0
	status, _ := d.GetStatus()
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.PersistentVolumeClaimName != "" {
				count++
			}
		}
		return nil
	})
	return count
}

// ReadyVolumeCount returns the number of volumes for the deployment that are in ready state
func (d *Deployment) ReadyVolumeCount() int {
	count := 0
	status, _ := d.GetStatus()
	pvcs, _ := d.GetOwnedPVCs() // Ignore errors on purpose
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
		for _, m := range list {
			if m.PersistentVolumeClaimName == "" {
				continue
			}
			// Find status
			for _, pvc := range pvcs {
				if pvc.Name == m.PersistentVolumeClaimName {
					if pvc.Status.Phase == v1.ClaimBound {
						count++
					}
				}
			}
		}
		return nil
	})
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
	ns := d.apiObject.Namespace
	svc, err := d.deps.KubeCli.CoreV1().Services(ns).Get(eaSvcName, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	scheme := "https"
	if !d.GetSpec().IsSecure() {
		scheme = "http"
	}
	nodeFetcher := func() (v1.NodeList, error) {
		result, err := d.deps.KubeCli.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return v1.NodeList{}, maskAny(err)
		}
		return *result, nil
	}
	portPredicate := func(p v1.ServicePort) bool {
		return p.TargetPort.IntValue() == k8sutil.ArangoPort
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
	status, _ := d.GetStatus()
	if current := status.CurrentImage; current != nil {
		license := "community"
		if current.Enterprise {
			license = "enterprise"
		}
		return string(current.ArangoDBVersion), license
	}
	return "", ""
}

// Members returns all members of the deployment by role.
func (d *Deployment) Members() map[api.ServerGroup][]server.Member {
	result := make(map[api.ServerGroup][]server.Member)
	status, _ := d.GetStatus()
	status.Members.ForeachServerGroup(func(group api.ServerGroup, list api.MemberStatusList) error {
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
		return nil
	})
	return result
}
