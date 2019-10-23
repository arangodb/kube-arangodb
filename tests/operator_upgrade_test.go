package tests

import (
	"fmt"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	kubeArangoClient "github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/dchest/uniuri"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

const (
	operatorTestDeploymentName string = "arango-deployment-operator"
	oldOperatorTestImage       string = "arangodb/kube-arangodb:0.3.16"
)

func TestOperatorUpgradeFrom038(t *testing.T) {
	ns := getNamespace(t)
	kubecli := mustNewKubeClient(t)
	c := kubeArangoClient.MustNewInCluster()

	if err := waitForArangoDBPodsGone(ns, kubecli); err != nil {
		t.Fatalf("Remaining arangodb pods did not vanish, can not start test: %v", err)
	}

	currentimage, err := updateOperatorImage(t, ns, kubecli, oldOperatorTestImage)
	if err != nil {
		t.Fatalf("Could not replace operator with old image: %v", err)
	}
	defer updateOperatorImage(t, ns, kubecli, currentimage)

	if err := waitForOperatorImage(ns, kubecli, oldOperatorTestImage); err != nil {
		t.Fatalf("Old Operator not ready in time: %v", err)
	}

	depl := newDeployment(fmt.Sprintf("opup-%s", uniuri.NewLen(4)))
	depl.Spec.TLS = api.TLSSpec{}         // should auto-generate cert
	depl.Spec.SetDefaults(depl.GetName()) // this must be last

	// Create deployment
	if _, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl); err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer removeDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	_, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	podsWatcher, err := kubecli.CoreV1().Pods(ns).Watch(metav1.ListOptions{
		LabelSelector: fields.OneTermEqualSelector("app", "arangodb").String(),
	})
	if err != nil {
		t.Fatalf("Failed to watch pods: %v", err)
	}
	defer podsWatcher.Stop()

	errorChannel := make(chan error)
	go func() {
		var addedPods []string
		for {
			select {
			case ev, ok := <-podsWatcher.ResultChan():
				if !ok {
					return // Abort
				}
				if pod, ok := ev.Object.(*v1.Pod); ok {
					if k8sutil.IsArangoDBImageIDAndVersionPod(*pod) {
						continue
					}

					switch ev.Type {
					case watch.Modified:
						if !k8sutil.IsPodReady(pod) {
							errorChannel <- fmt.Errorf("Pod no longer ready: %s", pod.GetName())
						}
						break
					case watch.Deleted:
						errorChannel <- fmt.Errorf("Pod was deleted: %s", pod.GetName())
						break
					case watch.Added:
						if len(addedPods) >= 9 {
							errorChannel <- fmt.Errorf("New pod was created: %s", pod.GetName())
						}
						addedPods = append(addedPods, pod.GetName())
						break
					}
				}
			}
		}
	}()

	if _, err := updateOperatorImage(t, ns, kubecli, currentimage); err != nil {
		t.Fatalf("Failed to replace new ")
	}

	if err := waitForOperatorImage(ns, kubecli, currentimage); err != nil {
		t.Fatalf("New operator not ready in time: %v", err)
	}

	select {
	case <-time.After(1 * time.Minute):
		break // cool
	case err := <-errorChannel:
		// not cool
		t.Errorf("Deployment had error: %v", err)
	}
}

func updateOperatorImage(t *testing.T, ns string, kube kubernetes.Interface, newImage string) (string, error) {
	for {
		depl, err := kube.AppsV1().Deployments(ns).Get(operatorTestDeploymentName, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		old, err := getOperatorImage(depl)
		if err != nil {
			return "", err
		}
		setOperatorImage(depl, newImage)
		if _, err := kube.AppsV1().Deployments(ns).Update(depl); k8sutil.IsConflict(err) {
			continue
		} else if err != nil {
			return "", err
		}
		return old, nil
	}
}

func updateOperatorDeployment(ns string, kube kubernetes.Interface) (*appsv1.Deployment, error) {
	return kube.AppsV1().Deployments(ns).Get(operatorTestDeploymentName, metav1.GetOptions{})
}

func getOperatorImage(depl *appsv1.Deployment) (string, error) {
	for _, c := range depl.Spec.Template.Spec.Containers {
		if c.Name == "operator" {
			return c.Image, nil
		}
	}

	return "", fmt.Errorf("Operator container not found")
}

func setOperatorImage(depl *appsv1.Deployment, image string) {
	for i := range depl.Spec.Template.Spec.Containers {
		c := &depl.Spec.Template.Spec.Containers[i]
		if c.Name == "operator" {
			c.Image = image
		}
	}
}

func waitForArangoDBPodsGone(ns string, kube kubernetes.Interface) error {
	return retry.Retry(func() error {
		_, err := kube.CoreV1().Pods(ns).List(metav1.ListOptions{
			LabelSelector: fields.OneTermEqualSelector("app", "arangodb").String(),
		})
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return err
	}, deploymentReadyTimeout)
}

func waitForOperatorImage(ns string, kube kubernetes.Interface, image string) error {
	return retry.Retry(func() error {
		pods, err := kube.CoreV1().Pods(ns).List(metav1.ListOptions{
			LabelSelector: fields.OneTermEqualSelector("app", operatorTestDeploymentName).String(),
		})
		if err != nil {
			return err
		}
		for _, pod := range pods.Items {
			for _, c := range pod.Spec.Containers {
				if c.Name == "operator" {
					if c.Image != image {
						return fmt.Errorf("in pod %s found image %s, expected %s", pod.Name, c.Image, image)
					}
				}
			}
		}
		return nil
	}, deploymentReadyTimeout)
}
