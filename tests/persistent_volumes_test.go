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
// Author Jan Christoph Uhde <jan@uhdejc.com>
//
package tests

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/arangodb/arangosync-client/pkg/retry"

	"github.com/dchest/uniuri"
	"github.com/stretchr/testify/assert"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"github.com/arangodb/kube-arangodb/pkg/util"
)

// TODO - add description
func TestPVCExists(t *testing.T) {
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	//k8sClient := mustNewKubeClient(t)

	// volumesList, err := k8sClient.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	// assert.NoError(t, err, "error while listing volumes")
	// claimsList, err := k8sClient.CoreV1().PersistentVolumeClaims(k8sNameSpace).List(metav1.ListOptions{})
	// assert.NoError(t, err, "error while listing volume claims")

	// fmt.Printf("----------------------------------------")
	// fmt.Printf("%v %v", volumesList, claimsList)
	// fmt.Printf("----------------------------------------")
	// fmt.Printf("%v %v", len(volumesList.Items), len(claimsList.Items))
	// fmt.Printf("----------------------------------------")

	mode := api.DeploymentModeCluster
	engine := api.StorageEngineRocksDB

	deploymentClient := kubeArangoClient.MustNewInCluster()
	deploymentTemplate := newDeployment(strings.Replace(fmt.Sprintf("tpers-%s-%s-%s", mode[:2], engine[:2], uniuri.NewLen(4)), ".", "", -1))
	deploymentTemplate.Spec.Mode = api.NewMode(mode)
	deploymentTemplate.Spec.StorageEngine = api.NewStorageEngine(engine)
	deploymentTemplate.Spec.TLS = api.TLSSpec{}
	//deploymentTemplate.Spec.Environment = api.NewEnvironment(api.EnvironmentDevelopment)
	//deploymentTemplate.Spec.Image = util.NewString("arangodb/arangodb:3.3.4")
	//deploymentTemplate.Spec.DBServers.Count = util.NewInt(numNodes + 1)
	deploymentTemplate.Spec.SetDefaults(deploymentTemplate.GetName()) // this must be last
	assert.NoError(t, deploymentTemplate.Spec.Validate())

	// Create deployment
	_, err := deploymentClient.DatabaseV1().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate)
	assert.NoError(t, err, "failed to create deplyment: %s", err)

	_, err = waitUntilDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace, deploymentIsReady())
	assert.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err)) // <-- fails here at the moment

	// TODO - add tests that check the number of volumes and claims

	// Cleanup
	removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)
}

func TestPVCResize(t *testing.T) {
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	k8sClient := mustNewKubeClient(t)

	mode := api.DeploymentModeCluster
	engine := api.StorageEngineRocksDB

	size10GB, _ := resource.ParseQuantity("10Gi")
	size08GB, _ := resource.ParseQuantity("8Gi")

	deploymentClient := kubeArangoClient.MustNewInCluster()
	deploymentTemplate := newDeployment(strings.Replace(fmt.Sprintf("trsz-%s-%s-%s", mode[:2], engine[:2], uniuri.NewLen(4)), ".", "", -1))
	deploymentTemplate.Spec.Mode = api.NewMode(mode)
	deploymentTemplate.Spec.StorageEngine = api.NewStorageEngine(engine)
	deploymentTemplate.Spec.TLS = api.TLSSpec{}
	deploymentTemplate.Spec.DBServers.Resources.Requests = corev1.ResourceList{corev1.ResourceStorage: size08GB}
	deploymentTemplate.Spec.SetDefaults(deploymentTemplate.GetName()) // this must be last
	assert.NoError(t, deploymentTemplate.Spec.Validate())

	// Create deployment
	_, err := deploymentClient.DatabaseV1().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate)
	defer removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)
	assert.NoError(t, err, "failed to create deplyment: %s", err)

	depl, err := waitUntilDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace, deploymentIsReady())
	assert.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))

	// Get list of all pvcs for dbservers
	for _, m := range depl.Status.Members.DBServers {
		pvc, err := k8sClient.CoreV1().PersistentVolumeClaims(k8sNameSpace).Get(m.PersistentVolumeClaimName, metav1.GetOptions{})
		assert.NoError(t, err, "failed to get pvc: %s", err)
		volumeSize, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		assert.True(t, ok, "pvc does not have storage resource")
		assert.True(t, volumeSize.Cmp(size08GB) == 0, "wrong volume size: expected: %s, found: %s", size08GB.String(), volumeSize.String())
	}

	// Update the deployment
	// Try to change image version
	depl, err = updateDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace,
		func(depl *api.DeploymentSpec) {
			depl.DBServers.Resources.Requests[corev1.ResourceStorage] = size10GB
		})
	if err != nil {
		t.Fatalf("Failed to update the deployment")
	} else {
		t.Log("Updated deployment")
	}

	if err := retry.Retry(func() error {
		// Get list of all pvcs for dbservers and check for new size
		for _, m := range depl.Status.Members.DBServers {
			pvc, err := k8sClient.CoreV1().PersistentVolumeClaims(k8sNameSpace).Get(m.PersistentVolumeClaimName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			volumeSize, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			if !ok {
				return fmt.Errorf("pvc does not have storage resource")
			}
			if volumeSize.Cmp(size10GB) != 0 {
				return fmt.Errorf("wrong pvc size: expected: %s, found: %s", size10GB.String(), volumeSize.String())
			}
			volume, err := k8sClient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			volumeSize, ok = volume.Spec.Capacity[corev1.ResourceStorage]
			if !ok {
				return fmt.Errorf("pv does not have storage resource")
			}
			if volumeSize.Cmp(size10GB) != 0 {
				return fmt.Errorf("wrong volume size: expected: %s, found: %s", size10GB.String(), volumeSize.String())
			}
			if k8sutil.IsPersistentVolumeClaimFileSystemResizePending(pvc) {
				return fmt.Errorf("persistent volume claim file system resize pending")
			}
		}
		return nil
	}, 5*time.Minute); err != nil {
		t.Fatalf("PVCs not resized: %s", err.Error())
	}

}

func TestPVCTemplateResize(t *testing.T) {
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	k8sClient := mustNewKubeClient(t)

	mode := api.DeploymentModeCluster
	engine := api.StorageEngineRocksDB

	size10GB, _ := resource.ParseQuantity("10Gi")
	size08GB, _ := resource.ParseQuantity("8Gi")

	deploymentClient := kubeArangoClient.MustNewInCluster()
	deploymentTemplate := newDeployment(strings.Replace(fmt.Sprintf("trsz-%s-%s-%s", mode[:2], engine[:2], uniuri.NewLen(4)), ".", "", -1))
	deploymentTemplate.Spec.Mode = api.NewMode(mode)
	deploymentTemplate.Spec.StorageEngine = api.NewStorageEngine(engine)
	deploymentTemplate.Spec.TLS = api.TLSSpec{}
	deploymentTemplate.Spec.SetDefaults(deploymentTemplate.GetName()) // this must be last
	assert.NoError(t, deploymentTemplate.Spec.Validate())
	assert.NotNil(t, deploymentTemplate.Spec.DBServers.VolumeClaimTemplate)
	deploymentTemplate.Spec.DBServers.VolumeClaimTemplate.Spec.Resources.Requests[corev1.ResourceStorage] = size08GB

	// Create deployment
	_, err := deploymentClient.DatabaseV1().ArangoDeployments(k8sNameSpace).Create(deploymentTemplate)
	defer removeDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace)
	assert.NoError(t, err, "failed to create deplyment: %s", err)

	depl, err := waitUntilDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace, deploymentIsReady())
	assert.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))

	// Get list of all pvcs for dbservers
	for _, m := range depl.Status.Members.DBServers {
		pvc, err := k8sClient.CoreV1().PersistentVolumeClaims(k8sNameSpace).Get(m.PersistentVolumeClaimName, metav1.GetOptions{})
		assert.NoError(t, err, "failed to get pvc: %s", err)
		volumeSize, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		assert.True(t, ok, "pvc does not have storage resource")
		assert.True(t, volumeSize.Cmp(size08GB) == 0, "wrong volume size: expected: %s, found: %s", size08GB.String(), volumeSize.String())
	}

	// Update the deployment
	// Try to change image version
	depl, err = updateDeployment(deploymentClient, deploymentTemplate.GetName(), k8sNameSpace,
		func(depl *api.DeploymentSpec) {
			depl.DBServers.VolumeClaimTemplate.Spec.Resources.Requests[corev1.ResourceStorage] = size10GB
		})
	if err != nil {
		t.Fatalf("Failed to update the deployment")
	} else {
		t.Log("Updated deployment")
	}

	if err := retry.Retry(func() error {
		// Get list of all pvcs for dbservers and check for new size
		for _, m := range depl.Status.Members.DBServers {
			pvc, err := k8sClient.CoreV1().PersistentVolumeClaims(k8sNameSpace).Get(m.PersistentVolumeClaimName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			volumeSize, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
			if !ok {
				return fmt.Errorf("pvc does not have storage resource")
			}
			if volumeSize.Cmp(size10GB) != 0 {
				return fmt.Errorf("wrong pvc size: expected: %s, found: %s", size10GB.String(), volumeSize.String())
			}
			volume, err := k8sClient.CoreV1().PersistentVolumes().Get(pvc.Spec.VolumeName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			volumeSize, ok = volume.Spec.Capacity[corev1.ResourceStorage]
			if !ok {
				return fmt.Errorf("pv does not have storage resource")
			}
			if volumeSize.Cmp(size10GB) != 0 {
				return fmt.Errorf("wrong volume size: expected: %s, found: %s", size10GB.String(), volumeSize.String())
			}
			if k8sutil.IsPersistentVolumeClaimFileSystemResizePending(pvc) {
				return fmt.Errorf("persistent volume claim file system resize pending")
			}
		}
		return nil
	}, 5*time.Minute); err != nil {
		t.Fatalf("PVCs not resized: %s", err.Error())
	}

}
