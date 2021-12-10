//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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
// Author Tomasz Mielech
//

package k8sutil

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type TimeoutRunFunc func(ctxChild context.Context) error

const (
	// LabelKeyArangoDeployment is the key of the label used to store the ArangoDeployment name in
	LabelKeyArangoDeployment = "arango_deployment"
	// LabelKeyArangoLocalStorage is the key of the label used to store the ArangoLocalStorage name in
	LabelKeyArangoLocalStorage = "arango_local_storage"
	// LabelKeyApp is the key of the label used to store the application name in (fixed to AppName)
	LabelKeyApp = "app"
	// LabelKeyRole is the key of the label used to store the role of the resource in
	LabelKeyRole = "role"
	// LabelKeyArangoExporter is the key of the label used to indicate that a exporter is present
	LabelKeyArangoExporter = "arango_exporter"
	// LabelKeyArangoMember is the key of the label used to store the ArangoDeployment member ID in
	LabelKeyArangoMember = "deployment.arangodb.com/member"
	// LabelKeyArangoZone is the key of the label used to store the ArangoDeployment zone ID in
	LabelKeyArangoZone = "deployment.arangodb.com/zone"
	// LabelKeyArangoScheduled is the key of the label used to define that member is already scheduled
	LabelKeyArangoScheduled = "deployment.arangodb.com/scheduled"
	// LabelKeyArangoTopology is the key of the label used to store the ArangoDeployment topology ID in
	LabelKeyArangoTopology = "deployment.arangodb.com/topology"

	// AppName is the fixed value for the "app" label
	AppName = "arangodb"

	// minDefaultRequestTimeout is minimum default request timeout to k8s.
	minDefaultRequestTimeout = time.Second * 3
)

var requestTimeout = minDefaultRequestTimeout

// GetRequestTimeout gets request timeout for one call to kubernetes.
func GetRequestTimeout() time.Duration {
	return requestTimeout
}

// RunWithTimeout runs the function with the provided timeout or with default timeout.
func RunWithTimeout(ctx context.Context, run TimeoutRunFunc, timeout ...time.Duration) error {
	t := GetRequestTimeout()
	if len(timeout) > 0 {
		t = timeout[0]
	}

	ctxChild, cancel := context.WithTimeout(ctx, t)
	defer cancel()

	return run(ctxChild)
}

// SetRequestTimeout sets request timeout for one call to kubernetes.
func SetRequestTimeout(timeout time.Duration) {
	if timeout > minDefaultRequestTimeout {
		requestTimeout = timeout
	}
}

// AddOwnerRefToObject adds given owner reference to given object
func AddOwnerRefToObject(obj metav1.Object, ownerRef *metav1.OwnerReference) {
	if ownerRef != nil {
		obj.SetOwnerReferences(append(obj.GetOwnerReferences(), *ownerRef))
	}
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
func DeploymentListOpt(deploymentName string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForDeployment(deploymentName, "")).String(),
	}
}

// LocalStorageListOpt creates a ListOptions matching all labels for the given local storage name.
func LocalStorageListOpt(localStorageName, role string) metav1.ListOptions {
	return metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(LabelsForLocalStorage(localStorageName, role)).String(),
	}
}
