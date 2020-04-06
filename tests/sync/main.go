package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/arangodb/kube-arangodb/pkg/apis/replication"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"

	sync "github.com/arangodb/arangosync-client/client"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	dapi "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	rapi "github.com/arangodb/kube-arangodb/pkg/apis/replication/v1"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

var (
	arangoImage          string
	arangoSyncTestImage  string
	arangoSyncImage      string
	licenseKeySecretName string
	namespace            string
	additionalTestArgs   string
)

const (
	accessPackageSecretName = "dst-access-package"
	dstDeploymentName       = "dc-dst"
	srcDeploymentName       = "dc-src"
	replicationResourceName = "dc-dst-src-replication"
	arangosyncTestPodName   = "kube-arango-sync-tests"
)

func init() {
	flag.StringVar(&arangoImage, "arango-image", "arangodb/enterprise:latest", "ArangoDB Enterprise image used for test")
	flag.StringVar(&arangoSyncTestImage, "arango-sync-test-image", "", "ArangoSync test image")
	flag.StringVar(&arangoSyncImage, "arango-sync-image", "", "ArangoSync Image used for testing")
	flag.StringVar(&licenseKeySecretName, "license-key-secret-name", "arangodb-license-key", "Secret name of the license key used for the deployments")
	flag.StringVar(&namespace, "namespace", "default", "Testing namespace")
	flag.StringVar(&additionalTestArgs, "test-args", "", "Additional parameters passed to the test executable")
}

func newDeployment(ns, name string) *dapi.ArangoDeployment {
	return &dapi.ArangoDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: dapi.SchemeGroupVersion.String(),
			Kind:       deployment.ArangoDeploymentResourceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
			// OwnerReferences: []metav1.OwnerReference{
			// 	metav1.OwnerReference{
			// 	},
			// },
		},
		Spec: dapi.DeploymentSpec{
			Image: util.NewString(arangoImage),
			License: dapi.LicenseSpec{
				SecretName: util.NewString(licenseKeySecretName),
			},
		},
	}
}

func newSyncDeployment(ns, name string, accessPackage bool) *dapi.ArangoDeployment {
	d := newDeployment(ns, name)
	d.Spec.Sync = dapi.SyncSpec{
		Enabled: util.NewBool(true),
		ExternalAccess: dapi.SyncExternalAccessSpec{
			ExternalAccessSpec: dapi.ExternalAccessSpec{
				Type: dapi.NewExternalAccessType(dapi.ExternalAccessTypeNone),
			},
		},
	}

	d.Spec.SyncMasters.Args = append(d.Spec.SyncMasters.Args, "--log.level=debug")
	d.Spec.SyncWorkers.Args = append(d.Spec.SyncWorkers.Args, "--log.level=debug")

	if accessPackage {
		d.Spec.Sync.ExternalAccess.AccessPackageSecretNames = []string{accessPackageSecretName}
	}

	if arangoSyncImage != "" {
		d.Spec.Sync.Image = util.NewString(arangoSyncImage)
	}
	return d
}

func newReplication(ns, name string) *rapi.ArangoDeploymentReplication {
	return &rapi.ArangoDeploymentReplication{
		TypeMeta: metav1.TypeMeta{
			APIVersion: rapi.SchemeGroupVersion.String(),
			Kind:       replication.ArangoDeploymentReplicationResourceKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: rapi.DeploymentReplicationSpec{
			Source: rapi.EndpointSpec{
				DeploymentName: util.NewString(srcDeploymentName),
				Authentication: rapi.EndpointAuthenticationSpec{
					KeyfileSecretName: util.NewString(accessPackageSecretName),
				},
				TLS: rapi.EndpointTLSSpec{
					CASecretName: util.NewString(accessPackageSecretName),
				},
			},
			Destination: rapi.EndpointSpec{
				DeploymentName: util.NewString(dstDeploymentName),
			},
		},
	}
}

func newArangoSyncTestJob(ns, name string) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
}

func waitForSyncDeploymentReady(ctx context.Context, ns, name string, kubecli kubernetes.Interface, c versioned.Interface) error {
	return retry.Retry(func() error {
		deployment, err := c.DatabaseV1().ArangoDeployments(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		sc, err := mustNewArangoDBSyncClient(ctx, kubecli, deployment)
		if err != nil {
			return err
		}

		info, err := sc.Master().Status(ctx)
		if err != nil {
			return err
		}

		if info.Status != sync.SyncStatusRunning {
			return fmt.Errorf("SyncStatus not running: %s", info.Status)
		}

		return nil
	}, 5*time.Minute)
}

func setupArangoDBCluster(ctx context.Context, kube kubernetes.Interface, c versioned.Interface) error {

	dstSpec := newSyncDeployment(namespace, dstDeploymentName, false)
	srcSpec := newSyncDeployment(namespace, srcDeploymentName, true)

	if _, err := c.DatabaseV1().ArangoDeployments(namespace).Create(srcSpec); err != nil {
		return err
	}
	if _, err := c.DatabaseV1().ArangoDeployments(namespace).Create(dstSpec); err != nil {
		return err
	}

	replSpec := newReplication(namespace, replicationResourceName)
	if _, err := c.ReplicationV1().ArangoDeploymentReplications(namespace).Create(replSpec); err != nil {
		return err
	}

	log.Print("Deployments and Replication created")

	//if err := waitForSyncDeploymentReady(ctx, namespace, srcSpec.GetName(), kube, c); err != nil {
	//	return errors.Wrap(err, "Source Cluster not ready")
	//}

	if err := waitForSyncDeploymentReady(ctx, namespace, dstSpec.GetName(), kube, c); err != nil {
		return errors.Wrap(err, "Destination Cluster not ready")
	}

	log.Print("Deployments and Replication ready")

	return nil
}

func waitForReplicationGone(ns, name string, c versioned.Interface) error {
	return retry.Retry(func() error {
		if _, err := c.ReplicationV1().ArangoDeploymentReplications(ns).Get(name, metav1.GetOptions{}); k8sutil.IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
		return fmt.Errorf("Replication resource not gone")
	}, 1*time.Minute)
}

func waitForDeploymentGone(ns, name string, c versioned.Interface) error {
	return retry.Retry(func() error {
		if _, err := c.DatabaseV1().ArangoDeployments(ns).Get(name, metav1.GetOptions{}); k8sutil.IsNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
		return fmt.Errorf("Deployment resource %s not gone", name)
	}, 1*time.Minute)
}

func removeReplicationWaitForCompletion(ns, name string, c versioned.Interface) error {
	if err := c.ReplicationV1().ArangoDeploymentReplications(ns).Delete(name, &metav1.DeleteOptions{}); err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := waitForReplicationGone(ns, name, c); err != nil {
		return err
	}
	return nil
}

func removeDeploymentWaitForCompletion(ns, name string, c versioned.Interface) error {
	if err := c.DatabaseV1().ArangoDeployments(ns).Delete(name, &metav1.DeleteOptions{}); err != nil {
		if k8sutil.IsNotFound(err) {
			return nil
		}
		return err
	}
	if err := waitForDeploymentGone(ns, name, c); err != nil {
		return err
	}
	return nil
}

func cleanupArangoDBCluster(ctx context.Context, kube kubernetes.Interface, c versioned.Interface) error {
	if err := removeReplicationWaitForCompletion(namespace, replicationResourceName, c); err != nil {
		return err
	}
	if err := removeDeploymentWaitForCompletion(namespace, dstDeploymentName, c); err != nil {
		return err
	}
	if err := removeDeploymentWaitForCompletion(namespace, srcDeploymentName, c); err != nil {
		return err
	}
	return nil
}

func waitForPodRunning(ns, name string, kube kubernetes.Interface) error {
	return retry.Retry(func() error {
		pod, err := kube.CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		if !k8sutil.IsPodReady(pod) {
			return fmt.Errorf("pod not ready")
		}
		return nil

	}, 1*time.Minute)
}

func copyPodLogs(ns, name string, kube kubernetes.Interface) error {
	logs, err := kube.CoreV1().Pods(ns).GetLogs(name, &corev1.PodLogOptions{
		Follow: true,
	}).Stream()
	if err != nil {
		return err
	}

	defer logs.Close()
	if _, err := io.Copy(os.Stdout, logs); err != nil {
		return err
	}
	return nil
}

func createArangoSyncTestPod(ns, name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				corev1.Container{
					Name:            "tests",
					Image:           arangoSyncTestImage,
					ImagePullPolicy: corev1.PullAlways,
					Args:            []string{"-test.v", additionalTestArgs},
					Env: []corev1.EnvVar{
						corev1.EnvVar{
							Name:  "MASTERAENDPOINTS",
							Value: fmt.Sprintf("https://%s-sync.%s.svc:8629/", srcDeploymentName, namespace),
						},
						corev1.EnvVar{
							Name:  "MASTERBENDPOINTS",
							Value: fmt.Sprintf("https://%s-sync.%s.svc:8629/", dstDeploymentName, namespace),
						},
						corev1.EnvVar{
							Name:  "CLUSTERAENDPOINTS",
							Value: fmt.Sprintf("https://%s.%s.svc:8529/", srcDeploymentName, namespace),
						},
						corev1.EnvVar{
							Name:  "CLUSTERBENDPOINTS",
							Value: fmt.Sprintf("https://%s.%s.svc:8529/", dstDeploymentName, namespace),
						},
						corev1.EnvVar{
							Name:  "CLUSTERACACERT",
							Value: "/data/access/ca.crt",
						},
						corev1.EnvVar{
							Name:  "CLUSTERACLIENTCERT",
							Value: "/data/access/tls.keyfile",
						},
						corev1.EnvVar{
							Name:  "CLUSTERMANAGED",
							Value: "yes",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							MountPath: "/data/access",
							Name:      "access",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "access",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: accessPackageSecretName,
						},
					},
				},
			},
		},
	}
}

func runArangoSyncTests(kube kubernetes.Interface) error {

	// Start a new pod with the test image
	defer kube.CoreV1().Pods(namespace).Delete(arangosyncTestPodName, &metav1.DeleteOptions{})
	podspec := createArangoSyncTestPod(namespace, arangosyncTestPodName)
	if _, err := kube.CoreV1().Pods(namespace).Create(podspec); err != nil {
		return err
	}

	log.Printf("Test pod created")

	if err := waitForPodRunning(namespace, arangosyncTestPodName, kube); err != nil {
		return err
	}

	log.Printf("Test pod running, receiving log")

	if err := copyPodLogs(namespace, arangosyncTestPodName, kube); err != nil {
		return err
	}

	pod, err := kube.CoreV1().Pods(namespace).Get(arangosyncTestPodName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	if !k8sutil.IsPodSucceeded(pod) {
		return fmt.Errorf("Pod not succeded")
	}

	return nil
}

func main() {
	flag.Parse()
	ctx := context.Background()
	kube := k8sutil.MustNewKubeClient()
	c := client.MustNewClient()

	defer removeReplicationWaitForCompletion(namespace, replicationResourceName, c)
	defer removeDeploymentWaitForCompletion(namespace, dstDeploymentName, c)
	defer removeDeploymentWaitForCompletion(namespace, srcDeploymentName, c)
	if err := setupArangoDBCluster(ctx, kube, c); err != nil {
		log.Printf("Failed to setup deployment: %s", err.Error())
		return
	}

	exitCode := 0

	if err := runArangoSyncTests(kube); err != nil {
		log.Printf("ArangoSync tests failed: %s", err.Error())
		exitCode = 1
	}

	if err := cleanupArangoDBCluster(ctx, kube, c); err != nil {
		log.Printf("Failed to clean up deployments: %s", err.Error())
	}

	os.Exit(exitCode)
}

func mustNewArangoDBSyncClient(ctx context.Context, kubecli kubernetes.Interface, deployment *dapi.ArangoDeployment) (sync.API, error) {
	ns := deployment.GetNamespace()
	secrets := kubecli.CoreV1().Secrets(ns)
	secretName := deployment.Spec.Sync.Authentication.GetJWTSecretName()
	jwtSecret, err := k8sutil.GetTokenSecret(secrets, secretName)
	if err != nil {
		return nil, err
	}

	// Fetch service DNS name
	dnsName := k8sutil.CreateSyncMasterClientServiceDNSName(deployment)
	ep := sync.Endpoint{"https://" + net.JoinHostPort(dnsName, strconv.Itoa(k8sutil.ArangoSyncMasterPort))}

	api, err := sync.NewArangoSyncClient(ep, sync.AuthenticationConfig{JWTSecret: jwtSecret}, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	return api, nil
}
