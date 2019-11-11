package tests

import (
	"fmt"
	"testing"
	"time"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/dchest/uniuri"
	v1beta1 "k8s.io/api/scheduling/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
)

func waitForPriorityOfServerGroup(kube kubernetes.Interface, c versioned.Interface, depl, ns string, group api.ServerGroup, priority int32) error {
	return retry.Retry(func() error {

		apiObject, err := c.DatabaseV1().ArangoDeployments(ns).Get(depl, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for _, m := range apiObject.Status.Members.MembersOfGroup(group) {
			pod, err := kube.CoreV1().Pods(apiObject.Namespace).Get(m.PodName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			if pod.Spec.Priority == nil {
				return fmt.Errorf("No pod priority set")
			}

			if *pod.Spec.Priority != priority {
				return fmt.Errorf("Wrong pod priority, expected %d, found %d", priority, *pod.Spec.Priority)
			}
		}

		return nil
	}, 5*time.Minute)
}

// TestPriorityClasses creates a PriorityClass and associates coordinators with that class.
// Then check if the pods have the desired priority. Then change the class and check that the pods are rotated.
func TestPriorityClasses(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	lowClassName := "test-low-class"
	lowClassValue := int32(1000)

	highClassName := "test-high-class"
	highClassValue := int32(2000)

	// Create two priority classes
	if _, err := kubecli.SchedulingV1beta1().PriorityClasses().Create(&v1beta1.PriorityClass{
		Value:         lowClassValue,
		GlobalDefault: false,
		Description:   "Low priority test class",
		ObjectMeta: metav1.ObjectMeta{
			Name: lowClassName,
		},
	}); err != nil {
		t.Fatalf("Could not create PC: %v", err)
	}
	defer kubecli.SchedulingV1beta1().PriorityClasses().Delete(lowClassName, &metav1.DeleteOptions{})

	if _, err := kubecli.SchedulingV1beta1().PriorityClasses().Create(&v1beta1.PriorityClass{
		Value:         highClassValue,
		GlobalDefault: false,
		Description:   "Low priority test class",
		ObjectMeta: metav1.ObjectMeta{
			Name: highClassName,
		},
	}); err != nil {
		t.Fatalf("Could not create PC: %v", err)
	}
	defer kubecli.SchedulingV1beta1().PriorityClasses().Delete(highClassName, &metav1.DeleteOptions{})

	// Prepare deployment config
	depl := newDeployment("test-pc-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeCluster)
	depl.Spec.TLS = api.TLSSpec{CASecretName: util.NewString("None")}
	depl.Spec.DBServers.Count = util.NewInt(2)
	depl.Spec.Coordinators.Count = util.NewInt(2)
	depl.Spec.Coordinators.PriorityClassName = lowClassName
	depl.Spec.SetDefaults(depl.GetName()) // this must be last
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Create deployment
	_, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	_, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	if err := waitForPriorityOfServerGroup(kubecli, c, depl.GetName(), ns, api.ServerGroupCoordinators, lowClassValue); err != nil {
		t.Errorf("PDBs not as expected: %v", err)
	}

	_, err = updateDeployment(c, depl.GetName(), ns, func(spec *api.DeploymentSpec) {
		spec.Coordinators.PriorityClassName = highClassName
	})
	if err != nil {
		t.Fatalf("Failed to update deployment: %v", err)
	}

	// Check if priority class is updated
	if err := waitForPriorityOfServerGroup(kubecli, c, depl.GetName(), ns, api.ServerGroupCoordinators, highClassValue); err != nil {
		t.Errorf("Priority not as expected: %v", err)
	}

	removeDeployment(c, depl.GetName(), ns)
}
