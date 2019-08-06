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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/util/intstr"

	deplv1alpha "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	extclient "github.com/arangodb/kube-arangodb/pkg/client"
	acli "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

var (
	cmdReboot = &cobra.Command{
		Use:    "reboot",
		Run:    cmdRebootRun,
		Hidden: false,
	}

	rebootOptions struct {
		DeploymentName    string
		ImageName         string
		LicenseSecretName string
		Coordinators      int
	}

	cmdRebootInspect = &cobra.Command{
		Use:    "inspect",
		Run:    cmdRebootInspectRun,
		Hidden: true,
	}

	rebootInspectOptions struct {
		TargetDir string
	}
)

func init() {
	cmdMain.AddCommand(cmdReboot)
	cmdReboot.AddCommand(cmdRebootInspect)
	cmdReboot.Flags().StringVar(&rebootOptions.DeploymentName, "deployment-name", "rebooted-deployment", "Name of the deployment")
	cmdReboot.Flags().StringVar(&rebootOptions.ImageName, "image-name", "arangodb/arangodb:latest", "Image used for the deployment")
	cmdReboot.Flags().StringVar(&rebootOptions.LicenseSecretName, "license-secret-name", "", "Name of secret for license key")
	cmdReboot.Flags().IntVar(&rebootOptions.Coordinators, "coordinators", 1, "Initial number of coordinators")

	cmdRebootInspect.Flags().StringVar(&rebootInspectOptions.TargetDir, "target-dir", "/data", "Path to mounted database directory")
}

type inspectResult struct {
	UUID string `json:"uuid,omitempty"`
}

type inspectResponse struct {
	Error  *string        `json:"error,omitempty"`
	Result *inspectResult `json:"result,omitempty"`
}

type VolumeInspectResult struct {
	UUID  string
	Claim string
	Error error
}

func runVolumeInspector(ctx context.Context, kube kubernetes.Interface, ns, name, image, storageClassName string) (string, string, error) {

	deletePVC := true
	claimname := "arangodb-reboot-pvc-" + name
	pvcspec := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: claimname,
			Labels: map[string]string{
				"app":      "arangodb",
				"rebooted": "yes",
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			VolumeName:  name,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewQuantity(1024*1024*1024, resource.DecimalSI),
				},
			},
			StorageClassName: util.NewString(storageClassName),
		},
	}

	_, err := kube.CoreV1().PersistentVolumeClaims(ns).Create(&pvcspec)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create pvc")
	}
	defer func() {
		if deletePVC {
			cliLog.Debug().Str("pvc-name", claimname).Msg("deleting pvc")
			kube.CoreV1().PersistentVolumeClaims(ns).Delete(claimname, &metav1.DeleteOptions{})
		}
	}()

	podname := "arangodb-reboot-pod-" + name
	podspec := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: podname,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				corev1.Container{
					Name:            "inspector",
					Image:           image,
					ImagePullPolicy: corev1.PullAlways,
					Command:         []string{"arangodb_operator"},
					Args:            []string{"reboot", "inspect"},
					Env: []corev1.EnvVar{
						corev1.EnvVar{
							Name:  constants.EnvOperatorPodNamespace,
							Value: ns,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							MountPath: "/data",
							Name:      "data",
						},
					},
					Ports: []corev1.ContainerPort{
						corev1.ContainerPort{
							ContainerPort: 8080,
						},
					},
					ReadinessProbe: &corev1.Probe{
						Handler: corev1.Handler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/info",
								Port: intstr.FromInt(8080),
							},
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: claimname,
						},
					},
				},
			},
		},
	}

	_, err = kube.CoreV1().Pods(ns).Create(&podspec)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create pod")
	}
	defer kube.CoreV1().Pods(ns).Delete(podname, &metav1.DeleteOptions{})

	podwatch, err := kube.CoreV1().Pods(ns).Watch(metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("metadata.name", podname).String()})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to watch for pod")
	}
	defer podwatch.Stop()

	// wait until pod is terminated
	for {
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case ev, ok := <-podwatch.ResultChan():
			if !ok {
				return "", "", fmt.Errorf("result channel bad")
			}

			// get the pod
			pod, ok := ev.Object.(*corev1.Pod)
			if !ok {
				return "", "", fmt.Errorf("failed to get pod")
			}

			switch pod.Status.Phase {
			case corev1.PodFailed:
				return "", "", fmt.Errorf("pod failed: %s", pod.Status.Reason)
			case corev1.PodRunning:
				podReady := false
				for _, c := range pod.Status.Conditions {
					if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
						podReady = true
					}
				}

				if !podReady {
					continue
				}

				resp, err := http.Get("http://" + pod.Status.PodIP + ":8080/info")
				if err != nil {
					return "", "", errors.Wrap(err, "Failed to get info")
				}

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", "", errors.Wrap(err, "failed to read body")
				}

				var info inspectResponse
				if err := json.Unmarshal(body, &info); err != nil {
					return "", "", errors.Wrap(err, "failed to unmarshal response")
				}

				if info.Error != nil {
					return "", "", fmt.Errorf("pod returned error: %s", *info.Error)
				}
				deletePVC = false

				return info.Result.UUID, claimname, nil
			}
		}
	}
}

func doVolumeInspection(ctx context.Context, kube kubernetes.Interface, ns, name, storageClassName string, resultChan chan<- VolumeInspectResult, image string) {
	// Create Volume Claim
	// Create Pod mounting this volume
	// Wait for pod to be completed
	// Read logs - parse json
	// Delete pod
	uuid, claim, err := runVolumeInspector(ctx, kube, ns, name, image, storageClassName)
	if err != nil {
		resultChan <- VolumeInspectResult{Error: err}
	}
	resultChan <- VolumeInspectResult{UUID: uuid, Claim: claim}
}

func checkVolumeAvailable(kube kubernetes.Interface, vname string) (VolumeInfo, error) {
	volume, err := kube.CoreV1().PersistentVolumes().Get(vname, metav1.GetOptions{})
	if err != nil {
		return VolumeInfo{}, errors.Wrapf(err, "failed to GET volume %s", vname)
	}

	switch volume.Status.Phase {
	case corev1.VolumeAvailable:
		break
	case corev1.VolumeReleased:
		// we have to remove the claim reference
		volume.Spec.ClaimRef = nil
		if _, err := kube.CoreV1().PersistentVolumes().Update(volume); err != nil {
			return VolumeInfo{}, errors.Wrapf(err, "failed to remove claim reference")
		}
		break
	default:
		return VolumeInfo{}, fmt.Errorf("Volume %s phase is %s, expected %s", vname, volume.Status.Phase, corev1.VolumeAvailable)
	}

	return VolumeInfo{StorageClassName: volume.Spec.StorageClassName}, nil
}

type VolumeInfo struct {
	StorageClassName string
}

type VolumeListInfo map[string]VolumeInfo

func preflightChecks(kube kubernetes.Interface, volumes []string) (VolumeListInfo, error) {
	info := make(VolumeListInfo)
	// Check if all values are released
	for _, vname := range volumes {
		vi, err := checkVolumeAvailable(kube, vname)
		if err != nil {
			return nil, errors.Wrap(err, "preflight checks failed")
		}
		info[vname] = vi
	}

	return info, nil
}

func getMyImage(kube kubernetes.Interface, ns, name string) (string, error) {
	pod, err := kube.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return pod.Spec.Containers[0].Image, nil
}

func createArangoDeployment(cli acli.Interface, ns, deplname, arangoimage string, results map[string]VolumeInspectResult) error {

	prmr := make(map[string]VolumeInspectResult)
	agnt := make(map[string]VolumeInspectResult)

	for vname, info := range results {
		if strings.HasPrefix(info.UUID, "PRMR") {
			prmr[vname] = info
		} else if strings.HasPrefix(info.UUID, "AGNT") {
			agnt[vname] = info
		} else {
			return fmt.Errorf("unknown server type by uuid: %s", info.UUID)
		}
	}

	depl := deplv1alpha.ArangoDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: deplname,
		},
		Spec: deplv1alpha.DeploymentSpec{
			Image: util.NewString(arangoimage),
			Coordinators: deplv1alpha.ServerGroupSpec{
				Count: util.NewInt(rebootOptions.Coordinators),
			},
			Agents: deplv1alpha.ServerGroupSpec{
				Count: util.NewInt(len(agnt)),
			},
			DBServers: deplv1alpha.ServerGroupSpec{
				Count: util.NewInt(len(prmr)),
			},
		},
	}

	if rebootOptions.LicenseSecretName != "" {
		depl.Spec.License.SecretName = util.NewString(rebootOptions.LicenseSecretName)
	}

	for _, info := range agnt {
		depl.Status.Members.Agents = append(depl.Status.Members.Agents, deplv1alpha.MemberStatus{
			ID:                        info.UUID,
			PersistentVolumeClaimName: info.Claim,
			PodName:                   k8sutil.CreatePodName(deplname, deplv1alpha.ServerGroupAgents.AsRole(), info.UUID, "-rbt"),
		})
	}

	for _, info := range prmr {
		depl.Status.Members.DBServers = append(depl.Status.Members.DBServers, deplv1alpha.MemberStatus{
			ID:                        info.UUID,
			PersistentVolumeClaimName: info.Claim,
			PodName:                   k8sutil.CreatePodName(deplname, deplv1alpha.ServerGroupDBServers.AsRole(), info.UUID, "-rbt"),
		})
	}

	if _, err := cli.DatabaseV1alpha().ArangoDeployments(ns).Create(&depl); err != nil {
		return errors.Wrap(err, "failed to create ArangoDeployment")
	}

	return nil
}

func cmdRebootRun(cmd *cobra.Command, args []string) {

	volumes := args
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	podname := os.Getenv(constants.EnvOperatorPodName)

	// Create kubernetes client
	kubecli, err := k8sutil.NewKubeClient()
	if err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	extcli, err := extclient.NewInCluster()
	if err != nil {
		cliLog.Fatal().Err(err).Msg("failed to create arango extension client")
	}

	image, err := getMyImage(kubecli, namespace, podname)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("failed to get my image")
	}

	vinfo, err := preflightChecks(kubecli, volumes)
	if err != nil {
		cliLog.Fatal().Err(err).Msg("preflight checks failed")
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	resultChan := make(chan VolumeInspectResult)
	received := 0

	for _, volumeName := range volumes {
		cliLog.Debug().Str("volume", volumeName).Msg("Starting inspection")
		wg.Add(1)
		go func(vn string) {
			defer wg.Done()
			doVolumeInspection(ctx, kubecli, namespace, vn, vinfo[vn].StorageClassName, resultChan, image)
		}(volumeName)
	}

	members := make(map[string]VolumeInspectResult)

	for {
		if received == len(volumes) {
			break
		}

		select {
		case res := <-resultChan:
			if res.Error != nil {
				cliLog.Error().Err(res.Error).Msg("Inspection failed")
			} else {
				cliLog.Info().Str("claim", res.Claim).Str("uuid", res.UUID).Msg("Inspection completed")
			}
			members[res.UUID] = res
			received++
		case <-ctx.Done():
			panic(ctx.Err())
		}
	}

	cliLog.Debug().Msg("results complete - generating ArangoDeployment resource")

	if err := createArangoDeployment(extcli, namespace, rebootOptions.DeploymentName, rebootOptions.ImageName, members); err != nil {
		cliLog.Error().Err(err).Msg("failed to create deployment")
	}

	cliLog.Info().Msg("ArangoDeployment created.")

	// Wait for everyone to be completed
	wg.Wait()
}

// inspectDatabaseDirectory inspects the given directory and returns the inspection result or an error
func inspectDatabaseDirectory(dirname string) (*inspectResult, error) {
	// Access the database directory and look for the following files
	// 	UUID

	uuidfile := path.Join(dirname, "UUID")
	uuid, err := ioutil.ReadFile(uuidfile)
	if err != nil {
		return nil, err
	}

	return &inspectResult{UUID: strings.TrimSpace(string(uuid))}, nil
}

func cmdRebootInspectRun(cmd *cobra.Command, args []string) {

	var response inspectResponse
	result, err := inspectDatabaseDirectory(rebootInspectOptions.TargetDir)
	if err != nil {
		response.Error = util.NewString(err.Error())
	}

	response.Result = result

	json, err := json.Marshal(&response)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/info", func(w http.ResponseWriter, req *http.Request) {
		w.Write(json)
	})

	if http.ListenAndServe(":8080", nil); err != nil {
		cliLog.Fatal().Err(err).Msg("Failed to listen and server")
	}
}
