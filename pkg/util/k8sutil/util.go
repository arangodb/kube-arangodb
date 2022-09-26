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

package k8sutil

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	// LabelKeyArangoDeployment is the key of the label used to store the ArangoDeployment name in
	LabelKeyArangoDeployment = "arango_deployment"
	// LabelKeyArangoLocalStorage is the key of the label used to store the ArangoLocalStorage name in
	LabelKeyArangoLocalStorage = "arango_local_storage"
	// LabelKeyApp is the key of the label used to store the application name in (fixed to AppName)
	LabelKeyApp = "app"
	// LabelKeyRole is the key of the label used to store the role of the resource in
	LabelKeyRole = "role"
	// LabelKeyArangoExporter is the key of the label used to indicate that an exporter is present
	LabelKeyArangoExporter = "arango_exporter"
	// LabelKeyArangoMember is the key of the label used to store the ArangoDeployment member ID in
	LabelKeyArangoMember = "deployment.arangodb.com/member"
	// LabelKeyArangoZone is the key of the label used to store the ArangoDeployment zone ID in
	LabelKeyArangoZone = "deployment.arangodb.com/zone"
	// LabelKeyArangoScheduled is the key of the label used to define that member is already scheduled
	LabelKeyArangoScheduled = "deployment.arangodb.com/scheduled"
	// LabelKeyArangoTopology is the key of the label used to store the ArangoDeployment topology ID in
	LabelKeyArangoTopology = "deployment.arangodb.com/topology"
	// LabelKeyArangoLeader is the key of the label used to store the current leader of a group instances.
	LabelKeyArangoLeader = "deployment.arangodb.com/leader"
	// LabelKeyArangoActive is the key of the label used to mark members as active.
	LabelKeyArangoActive = "deployment.arangodb.com/active"
	// LabelValueArangoActive is the value of the label used to mark members as active.
	LabelValueArangoActive = "true"
	// AppName is the fixed value for the "app" label
	AppName = "arangodb"
)

// AddOwnerRefToObject adds given owner reference to given object
func AddOwnerRefToObject(obj meta.Object, ownerRef *meta.OwnerReference) {
	if ownerRef != nil {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), *ownerRef))
	}
}

// RemoveOwnerRefToObjectIfNeeded removes given owner reference to given object if it exists
func RemoveOwnerRefToObjectIfNeeded(obj meta.Object, ownerRef *meta.OwnerReference) bool {
	exists := -1
	if ownerRef != nil {
		own := obj.GetOwnerReferences()

		for id, existingOwnerRef := range own {
			if existingOwnerRef.UID == ownerRef.UID {
				exists = id
				break
			}
		}

		if exists == -1 {
			return false
		}

		no := make([]meta.OwnerReference, 0, len(own))

		for id := range own {
			if id == exists {
				continue
			}

			no = append(no, own[id])
		}

		obj.SetOwnerReferences(no)
		return true
	}

	return false
}

// UpdateOwnerRefToObjectIfNeeded add given owner reference to given object if it does not exist yet
func UpdateOwnerRefToObjectIfNeeded(obj meta.Object, ownerRef *meta.OwnerReference) bool {
	if ownerRef != nil {
		for _, existingOwnerRef := range obj.GetOwnerReferences() {
			if existingOwnerRef.UID == ownerRef.UID {
				return false
			}
		}

		AddOwnerRefToObject(obj, ownerRef)
		return true
	}
	return false
}

// LabelsForExporterServiceSelector returns a map of labels, used to select the all arangodb-exporter containers
func LabelsForExporterServiceSelector(deploymentName string) map[string]string {
	return map[string]string{
		LabelKeyArangoDeployment: deploymentName,
		LabelKeyArangoExporter:   "yes",
	}
}

// LabelsForExporterService returns a map of labels, used to select the all arangodb-exporter containers
func LabelsForExporterService(deploymentName string) map[string]string {
	return map[string]string{
		LabelKeyArangoDeployment: deploymentName,
		LabelKeyApp:              AppName,
	}
}

// LabelsForMember returns a map of labels, given to all resources for given deployment name and member id
func LabelsForMember(deploymentName, role, id string) map[string]string {
	l := LabelsForDeployment(deploymentName, role)

	if id != "" {
		l[LabelKeyArangoMember] = id
	}

	return l
}

// LabelsForActiveMember returns a map of labels, given to active members for given deployment name and member id
func LabelsForActiveMember(deploymentName, role, id string) map[string]string {
	l := LabelsForMember(deploymentName, role, id)

	l[LabelKeyArangoActive] = LabelValueArangoActive

	return l
}

// LabelsForLeaderMember returns a map of labels for given deployment name and member id and role and leadership.
func LabelsForLeaderMember(deploymentName, role, id string) map[string]string {
	l := LabelsForMember(deploymentName, role, id)
	l[LabelKeyArangoLeader] = "true"

	return l
}

// LabelsForDeployment returns a map of labels, given to all resources for given deployment name
func LabelsForDeployment(deploymentName, role string) map[string]string {
	l := map[string]string{
		LabelKeyArangoDeployment: deploymentName,
		LabelKeyApp:              AppName,
	}
	if role != "" {
		l[LabelKeyRole] = role
	}
	return l
}

// LabelsForLocalStorage returns a map of labels, given to all resources for given local storage name
func LabelsForLocalStorage(localStorageName, role string) map[string]string {
	l := map[string]string{
		LabelKeyArangoLocalStorage: localStorageName,
		LabelKeyApp:                AppName,
	}
	if role != "" {
		l[LabelKeyRole] = role
	}
	return l
}

// DeploymentListOpt creates a ListOptions matching all labels for the given deployment name.
func DeploymentListOpt(deploymentName string) meta.ListOptions {
	return meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForDeployment(deploymentName, "")).String(),
	}
}

// LocalStorageListOpt creates a ListOptions matching all labels for the given local storage name.
func LocalStorageListOpt(localStorageName, role string) meta.ListOptions {
	return meta.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForLocalStorage(localStorageName, role)).String(),
	}
}
