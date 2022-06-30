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

package cmd

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

	deplv1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	acli "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	pvcspec := core.PersistentVolumeClaim{
		ObjectMeta: meta.ObjectMeta{
			Name: claimname,
			Labels: map[string]string{
				"app":      "arangodb",
				"rebooted": "yes",
			},
		},
		Spec: core.PersistentVolumeClaimSpec{
			AccessModes: []core.PersistentVolumeAccessMode{core.ReadWriteOnce},
			VolumeName:  name,
			Resources: core.ResourceRequirements{
				Requests: core.ResourceList{
					core.ResourceStorage: *resource.NewQuantity(1024*1024*1024, resource.DecimalSI),
				},
			},
			StorageClassName: util.NewString(storageClassName),
		},
	}

	_, err := kube.CoreV1().PersistentVolumeClaims(ns).Create(context.Background(), &pvcspec, meta.CreateOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create pvc")
	}
	defer func() {
		if deletePVC {
			logger.Str("pvc-name", claimname).Debug("deleting pvc")
			kube.CoreV1().PersistentVolumeClaims(ns).Delete(context.Background(), claimname, meta.DeleteOptions{})
		}
	}()

	podname := "arangodb-reboot-pod-" + name
	podspec := core.Pod{
		ObjectMeta: meta.ObjectMeta{
			Name: podname,
		},
		Spec: core.PodSpec{
			RestartPolicy: core.RestartPolicyNever,
			Containers: []core.Container{
				core.Container{
					Name:            "inspector",
					Image:           image,
					ImagePullPolicy: core.PullAlways,
					Command:         []string{"arangodb_operator"},
					Args:            []string{"reboot", "inspect"},
					Env: []core.EnvVar{
						core.EnvVar{
							Name:  constants.EnvOperatorPodNamespace,
							Value: ns,
						},
					},
					VolumeMounts: []core.VolumeMount{
						core.VolumeMount{
							MountPath: "/data",
							Name:      "data",
						},
					},
					Ports: []core.ContainerPort{
						core.ContainerPort{
							ContainerPort: 8080,
						},
					},
					ReadinessProbe: &core.Probe{
						Handler: core.Handler{
							HTTPGet: &core.HTTPGetAction{
								Path: "/info",
								Port: intstr.FromInt(8080),
							},
						},
					},
				},
			},
			Volumes: []core.Volume{
				k8sutil.CreateVolumeWithPersitantVolumeClaim("data", claimname),
			},
		},
	}

	_, err = kube.CoreV1().Pods(ns).Create(context.Background(), &podspec, meta.CreateOptions{})
	if err != nil {
		return "", "", errors.Wrap(err, "failed to create pod")
	}
	defer kube.CoreV1().Pods(ns).Delete(context.Background(), podname, meta.DeleteOptions{})

	podwatch, err := kube.CoreV1().Pods(ns).Watch(context.Background(), meta.ListOptions{FieldSelector: fields.OneTermEqualSelector("metadata.name", podname).String()})
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
			pod, ok := ev.Object.(*core.Pod)
			if !ok {
				return "", "", fmt.Errorf("failed to get pod")
			}

			switch pod.Status.Phase {
			case core.PodFailed:
				return "", "", fmt.Errorf("pod failed: %s", pod.Status.Reason)
			case core.PodRunning:
				podReady := false
				for _, c := range pod.Status.Conditions {
					if c.Type == core.PodReady && c.Status == core.ConditionTrue {
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
	volume, err := kube.CoreV1().PersistentVolumes().Get(context.Background(), vname, meta.GetOptions{})
	if err != nil {
		return VolumeInfo{}, errors.Wrapf(err, "failed to GET volume %s", vname)
	}

	switch volume.Status.Phase {
	case core.VolumeAvailable:
		break
	case core.VolumeReleased:
		// we have to remove the claim reference
		volume.Spec.ClaimRef = nil
		if _, err := kube.CoreV1().PersistentVolumes().Update(context.Background(), volume, meta.UpdateOptions{}); err != nil {
			return VolumeInfo{}, errors.Wrapf(err, "failed to remove claim reference")
		}
	default:
		return VolumeInfo{}, fmt.Errorf("Volume %s phase is %s, expected %s", vname, volume.Status.Phase, core.VolumeAvailable)
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
	pod, err := kube.CoreV1().Pods(ns).Get(context.Background(), name, meta.GetOptions{})
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

	depl := deplv1.ArangoDeployment{
		ObjectMeta: meta.ObjectMeta{
			Name: deplname,
		},
		Spec: deplv1.DeploymentSpec{
			Image: util.NewString(arangoimage),
			Coordinators: deplv1.ServerGroupSpec{
				Count: util.NewInt(rebootOptions.Coordinators),
			},
			Agents: deplv1.ServerGroupSpec{
				Count: util.NewInt(len(agnt)),
			},
			DBServers: deplv1.ServerGroupSpec{
				Count: util.NewInt(len(prmr)),
			},
		},
	}

	if rebootOptions.LicenseSecretName != "" {
		depl.Spec.License.SecretName = util.NewString(rebootOptions.LicenseSecretName)
	}

	for _, info := range agnt {
		depl.Status.Members.Agents = append(depl.Status.Members.Agents, deplv1.MemberStatus{
			ID:                        info.UUID,
			PersistentVolumeClaimName: info.Claim,
			PodName:                   k8sutil.CreatePodName(deplname, deplv1.ServerGroupAgents.AsRole(), info.UUID, "-rbt"),
		})
	}

	for _, info := range prmr {
		depl.Status.Members.DBServers = append(depl.Status.Members.DBServers, deplv1.MemberStatus{
			ID:                        info.UUID,
			PersistentVolumeClaimName: info.Claim,
			PodName:                   k8sutil.CreatePodName(deplname, deplv1.ServerGroupDBServers.AsRole(), info.UUID, "-rbt"),
		})
	}

	if _, err := cli.DatabaseV1().ArangoDeployments(ns).Create(context.Background(), &depl, meta.CreateOptions{}); err != nil {
		return errors.Wrap(err, "failed to create ArangoDeployment")
	}

	return nil
}

func cmdRebootRun(cmd *cobra.Command, args []string) {

	volumes := args
	namespace := os.Getenv(constants.EnvOperatorPodNamespace)
	podname := os.Getenv(constants.EnvOperatorPodName)

	// Create kubernetes client
	client, ok := kclient.GetDefaultFactory().Client()
	if !ok {
		logger.Fatal("Failed to get client")
	}

	kubecli := client.Kubernetes()

	extcli := client.Arango()

	image, err := getMyImage(kubecli, namespace, podname)
	if err != nil {
		logger.Err(err).Fatal("failed to get my image")
	}

	vinfo, err := preflightChecks(kubecli, volumes)
	if err != nil {
		logger.Err(err).Fatal("preflight checks failed")
	}

	var wg sync.WaitGroup
	ctx := context.Background()
	resultChan := make(chan VolumeInspectResult)
	received := 0

	for _, volumeName := range volumes {
		logger.Str("volume", volumeName).Debug("Starting inspection")
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
				logger.Err(res.Error).Error("Inspection failed")
			} else {
				logger.Str("claim", res.Claim).Str("uuid", res.UUID).Info("Inspection completed")
			}
			members[res.UUID] = res
			received++
		case <-ctx.Done():
			panic(ctx.Err())
		}
	}

	logger.Debug("results complete - generating ArangoDeployment resource")

	if err := createArangoDeployment(extcli, namespace, rebootOptions.DeploymentName, rebootOptions.ImageName, members); err != nil {
		logger.Err(err).Error("failed to create deployment")
	}

	logger.Info("ArangoDeployment created.")

	// Wait for everyone to be completed
	wg.Wait()
}

// inspectDatabaseDirectory inspects the given directory and returns the inspection result or an error
func inspectDatabaseDirectory(dirname string) (*inspectResult, error) {
	// Access the database directory and look for the following files
	// 	UUID

	uuidfile := path.Join(dirname, "UUID")
	uuid, err := ioutil.ReadFile(path.Clean(uuidfile))
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
		logger.Err(err).Fatal("Failed to listen and serve")
	}
}
