//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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

	"k8s.io/client-go/kubernetes"

	storagev1 "k8s.io/api/storage/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"

	"github.com/stretchr/testify/require"

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

func TestPVCChangeStorage(t *testing.T) {
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	arangoClient := kubeArangoClient.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	mode := api.DeploymentModeCluster

	defaultStorageClass := getDefaultStorageClassOrDie(t, kubecli)
	randomString := strings.ToLower(uniuri.NewLen(4))
	newStorageClassName := defaultStorageClass.GetName() + randomString

	newStorage := defaultStorageClass.DeepCopy()
	newStorage.ObjectMeta = metav1.ObjectMeta{
		Name: newStorageClassName,
	}
	newStorage, err := kubecli.StorageV1().StorageClasses().Create(newStorage)
	require.NoError(t, err)
	defer func() {
		err := kubecli.StorageV1().StorageClasses().Delete(newStorage.Name, &metav1.DeleteOptions{})
		assert.NoError(t, err)
	}()

	name := strings.Replace(fmt.Sprintf("tcs-%s-%s", mode[:2], randomString), ".", "", -1)
	depl, err := newDeploymentWithValidation(name, func(deployment *api.ArangoDeployment) {
		var agentsCount, coordinatorCount, DBServersCount = 3, 2, 3

		deployment.Spec.Mode = api.NewMode(mode)
		deployment.Spec.Environment = api.NewEnvironment(api.EnvironmentProduction)

		volumeMode := corev1.PersistentVolumeFilesystem
		deployment.Spec.DBServers.VolumeClaimTemplate = &corev1.PersistentVolumeClaim{
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				VolumeMode: &volumeMode,
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
			},
		}

		deployment.Spec.DBServers.Count = util.NewInt(DBServersCount)
		deployment.Spec.Agents.Count = util.NewInt(agentsCount)
		deployment.Spec.Coordinators.Count = util.NewInt(coordinatorCount)
	})
	require.NoError(t, err)

	// Create deployment
	_, err = arangoClient.DatabaseV1().ArangoDeployments(k8sNameSpace).Create(depl)
	require.NoError(t, err, "failed to create deployment: %s", err)
	defer deferedCleanupDeployment(arangoClient, depl.GetName(), k8sNameSpace)

	depl, err = waitUntilDeployment(arangoClient, depl.GetName(), k8sNameSpace, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))
	require.NotNil(t, depl.Spec.DBServers.VolumeClaimTemplate)

	// Fill collection with documents
	documentGenerator := NewDocumentGenerator(kubecli, depl, "collectionTest", 3, 200)
	documentGenerator.generate(t, func(documentIndex int) interface{} {
		type oneValue struct {
			value int
		}
		return &oneValue{value: documentIndex}
	})

	// Update deployment
	_, err = updateDeployment(arangoClient, depl.GetName(), k8sNameSpace, func(spec *api.DeploymentSpec) {
		spec.DBServers.VolumeClaimTemplate.Spec.StorageClassName = util.NewString(newStorageClassName)
	})
	require.NoError(t, err, "failed to update deployment: %s", err)

	// Check for updated deployment
	isStorageChanged := func(deployment *api.ArangoDeployment) error {
		pvc := deployment.Spec.DBServers.VolumeClaimTemplate
		if pvc == nil {
			return fmt.Errorf("persistant volume claim can not be nil")
		}
		if pvc.Spec.StorageClassName == nil {
			return fmt.Errorf("storage class name can not be nil")
		}
		if *pvc.Spec.StorageClassName != newStorageClassName {
			return fmt.Errorf("storage class name has not been changed")
		}

		for _, server := range deployment.Status.Members.DBServers {
			pvc, err := kubecli.CoreV1().PersistentVolumeClaims(k8sNameSpace).Get(server.PersistentVolumeClaimName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if pvc.Spec.StorageClassName == nil {
				return fmt.Errorf("storage class name can not be nil")
			}
			if *pvc.Spec.StorageClassName != newStorageClassName {
				return fmt.Errorf("storage class name has not been chagned")
			}
		}
		return nil
	}

	depl, err = waitUntilDeployment(arangoClient, depl.GetName(), k8sNameSpace, isStorageChanged, time.Minute*5)
	require.NoError(t, err, "failed to change storage class for db servers: %s", err)

	// Check if documents are the same in the new storage
	documentGenerator.check(t)

	// Cleanup
	removeDeployment(arangoClient, depl.GetName(), k8sNameSpace)
}

// Test deprecated functionality for changing storage class
func TestPVCChangeStorageDeprecated(t *testing.T) {
	longOrSkip(t)

	k8sNameSpace := getNamespace(t)
	arangoClient := kubeArangoClient.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	mode := api.DeploymentModeCluster

	defaultStorageClass := getDefaultStorageClassOrDie(t, kubecli)
	randomString := strings.ToLower(uniuri.NewLen(4))
	newStorageClassName := defaultStorageClass.GetName() + randomString

	newStorage := defaultStorageClass.DeepCopy()
	newStorage.ObjectMeta = metav1.ObjectMeta{
		Name: newStorageClassName,
	}
	newStorage, err := kubecli.StorageV1().StorageClasses().Create(newStorage)
	require.NoError(t, err)
	defer func() {
		err := kubecli.StorageV1().StorageClasses().Delete(newStorage.Name, &metav1.DeleteOptions{})
		assert.NoError(t, err)
	}()

	name := strings.Replace(fmt.Sprintf("tcs-%s-%s", mode[:2], randomString), ".", "", -1)
	depl, err := newDeploymentWithValidation(name, func(deployment *api.ArangoDeployment) {
		var agentsCount, coordinatorCount, DBServersCount = 3, 2, 3

		deployment.Spec.Mode = api.NewMode(mode)
		deployment.Spec.Environment = api.NewEnvironment(api.EnvironmentProduction)

		deployment.Spec.DBServers.Resources.Requests = map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceStorage: resource.MustParse("2Gi"),
		}
		deployment.Spec.DBServers.StorageClassName = util.NewString(defaultStorageClass.Name)
		deployment.Spec.DBServers.Count = util.NewInt(DBServersCount)
		deployment.Spec.Agents.Count = util.NewInt(agentsCount)
		deployment.Spec.Coordinators.Count = util.NewInt(coordinatorCount)
	})
	require.NoError(t, err)

	// Create deployment
	_, err = arangoClient.DatabaseV1().ArangoDeployments(k8sNameSpace).Create(depl)
	require.NoError(t, err, "failed to create deployment: %s", err)
	defer deferedCleanupDeployment(arangoClient, depl.GetName(), k8sNameSpace)

	depl, err = waitUntilDeployment(arangoClient, depl.GetName(), k8sNameSpace, deploymentIsReady())
	require.NoError(t, err, fmt.Sprintf("Deployment not running in time: %s", err))

	// Fill collection with documents
	documentGenerator := NewDocumentGenerator(kubecli, depl, "collectionTest", 3, 200)
	documentGenerator.generate(t, func(documentIndex int) interface{} {
		type oneValue struct {
			value int
		}
		return &oneValue{value: documentIndex}
	})

	// Update deployment
	_, err = updateDeployment(arangoClient, depl.GetName(), k8sNameSpace, func(spec *api.DeploymentSpec) {
		spec.DBServers.StorageClassName = util.NewString(newStorageClassName)
	})
	require.NoError(t, err, "failed to update deployment: %s", err)

	// Check for updated deployment
	isDeprecatedStorageChanged := func(deployment *api.ArangoDeployment) error {
		for _, server := range deployment.Status.Members.DBServers {
			pvc, err := kubecli.CoreV1().PersistentVolumeClaims(k8sNameSpace).Get(server.PersistentVolumeClaimName, metav1.GetOptions{})
			if err != nil {
				return err
			}
			if pvc.Spec.StorageClassName == nil {
				return fmt.Errorf("storage class name can not be nil")
			}
			if *pvc.Spec.StorageClassName != newStorageClassName {
				return fmt.Errorf("storage class name has not been chagned")
			}
		}
		return nil
	}

	depl, err = waitUntilDeployment(arangoClient, depl.GetName(), k8sNameSpace, isDeprecatedStorageChanged, time.Minute*5)
	require.NoError(t, err, "failed to change storage class for db servers: %s", err)

	// Check if documents are the same in the new storage
	documentGenerator.check(t)

	// Cleanup
	removeDeployment(arangoClient, depl.GetName(), k8sNameSpace)
}

func getDefaultStorageClassOrDie(t *testing.T, kubecli kubernetes.Interface) *storagev1.StorageClass {
	var defaultStorageClass *storagev1.StorageClass
	storageClasses, err := kubecli.StorageV1().StorageClasses().List(metav1.ListOptions{})
	require.NoError(t, err)

	for _, sc := range storageClasses.Items {
		if k8sutil.StorageClassIsDefault(&sc) {
			defaultStorageClass = &sc
			break
		}
	}
	require.NotNilf(t, defaultStorageClass, "test needs default storage class")
	return defaultStorageClass
}
