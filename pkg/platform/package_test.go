//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package platform

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/yaml"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	platformChart "github.com/arangodb/kube-arangodb/pkg/handlers/platform/chart"
	platformService "github.com/arangodb/kube-arangodb/pkg/handlers/platform/service"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient/external"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
	operator "github.com/arangodb/kube-arangodb/pkg/util/tests"
	"github.com/arangodb/kube-arangodb/pkg/util/tests/suite"
)

//go:embed suite/sample-1.0.0.tgz
var sample_1_0_0 []byte

func saveFile[T any](t *testing.T, z T) string {
	f := fmt.Sprintf("%s/file", t.TempDir())

	data, err := yaml.Marshal(z)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(f, data, 0644))

	return f
}

func executeDumpPackage(t *testing.T, namespace, name string) helm.Package {
	data := bytes.NewBuffer(nil)

	executeCobraCommandC(t, data, "package", "dump", "--namespace", namespace, "--platform.name", name)

	o, err := util.JsonOrYamlUnmarshal[helm.Package](data.Bytes())
	require.NoError(t, err)
	require.NoError(t, o.Validate())
	return o
}

func executeExportPackage(t *testing.T, p helm.Package) string {
	require.NoError(t, p.Validate())

	out := fmt.Sprintf("%s/file", t.TempDir())

	pFile := saveFile(t, p)

	executeCobraCommandC(t, nil, "package", "export", pFile, out)

	return out
}

func executeInstallPackage(t *testing.T, namespace, name string, ps ...helm.Package) {
	z := make([]string, len(ps))
	for i, p := range ps {
		require.NoError(t, p.Validate())

		z[i] = saveFile(t, p)
	}

	args := append([]string{"package", "install", "--namespace", namespace, "--platform.name", name}, z...)

	executeCobraCommandC(t, nil, args...)
}

func executeImportPackage(t *testing.T, registry, path string) helm.Package {
	out := fmt.Sprintf("%s/file", t.TempDir())

	executeCobraCommandC(t, nil, "package", "import", "--registry.docker.insecure", registry, registry, path, out)

	o, err := util.JsonOrYamlUnmarshalFile[helm.Package](out)
	require.NoError(t, err)
	require.NoError(t, o.Validate())
	return o
}

func executeCobraCommandC(t *testing.T, out io.Writer, args ...string) {
	require.NoError(t, executeCobraCommand(t, out, args...))
}

func executeCobraCommand(t *testing.T, out io.Writer, args ...string) error {
	cmd, err := NewInstaller()
	require.NoError(t, err)

	if out != nil {
		cmd.SetOut(out)
	}

	cmd.SetArgs(append([]string{
		"--kubeconfig",
		external.TEST_KUBECONFIG.Get(),
	}, args...))

	return cmd.Execute()
}

func EnsureRegistry(t *testing.T, client kclient.Client, ns string) string {
	_, err := client.Kubernetes().AppsV1().Deployments(ns).Create(shutdown.Context(), &apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name:      "registry",
			Namespace: ns,
		},
		Spec: apps.DeploymentSpec{
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": "registry",
				},
			},
			Replicas: util.NewType[int32](1),
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"app": "registry",
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "registry",
							Image: "registry:2",
							Ports: []core.ContainerPort{
								{
									Name:          "registry",
									ContainerPort: 5000,
								},
							},
						},
					},
				},
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	_, err = client.Kubernetes().CoreV1().Services(ns).Create(shutdown.Context(), &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:      "registry",
			Namespace: ns,
		},
		Spec: core.ServiceSpec{
			Selector: map[string]string{
				"app": "registry",
			},
			Type: core.ServiceTypeNodePort,
			Ports: []core.ServicePort{
				{
					Name:       "registry",
					Port:       5000,
					TargetPort: intstr.FromInt(5000),
				},
			},
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	_, err = util.NewTimeoutFunc(func() (*apps.Deployment, error) {
		depl, err := client.Kubernetes().AppsV1().Deployments(ns).Get(shutdown.Context(), "registry", meta.GetOptions{})
		if err != nil {
			return nil, err
		}

		if depl.Status.ReadyReplicas == 1 {
			return depl, io.EOF
		}

		return nil, nil
	}).Run(shutdown.Context(), time.Minute, time.Second)
	require.NoError(t, err)

	svc, err := client.Kubernetes().CoreV1().Services(ns).Get(shutdown.Context(), "registry", meta.GetOptions{})
	require.NoError(t, err)
	require.Len(t, svc.Spec.Ports, 1)

	pf := k8sutil.NewPortForwarder(client.Config(), k8sutil.PortForwarderServiceDiscovery(ns, "registry"))
	q, err := pf.Start(shutdown.Context(), fmt.Sprintf("%d:%d", svc.Spec.Ports[0].NodePort, svc.Spec.Ports[0].Port))
	require.NoError(t, err)

	go func() {
		require.NoError(t, q.Wait())
	}()

	return fmt.Sprintf("localhost:%d", svc.Spec.Ports[0].NodePort)
}

func Test_Package(t *testing.T) {
	client, ns := external.ExternalClient(t)

	defer operator.NewTestingOperator(shutdown.Context(), t, ns, util.Image{Image: "operator:latest"}, client, platformChart.RegisterInformer, platformService.RegisterInformer)()

	deployment := tests.NewMetaObject[*api.ArangoDeployment](t, ns, "example",
		func(t *testing.T, obj *api.ArangoDeployment) {})

	tests.CreateObjects(t, client.Kubernetes(), client.Arango(), &deployment)

	// Run
	registry := EnsureRegistry(t, client, ns)

	t.Run("Define Chart", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), helm.Package{
			Packages: map[string]helm.PackageSpec{
				"sample": {
					Version: "1.0.0",
					Chart:   util.NewType(base64.StdEncoding.EncodeToString(sample_1_0_0)),
				},
			},
		})
	})

	t.Run("Install Chart", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), helm.Package{
			Packages: map[string]helm.PackageSpec{
				"sample": {
					Version: "1.0.0",
					Chart:   util.NewType(base64.StdEncoding.EncodeToString(sample_1_0_0)),
				},
			},
			Releases: map[string]helm.PackageRelease{
				"sample": {
					Package: "sample",
				},
			},
		})

		cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample")
		require.Equal(t, "PLACEHOLDER", cm.Data)

		cmi := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cmi.Data)
	})

	t.Run("Re-Install From Dump", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), executeDumpPackage(t, ns, deployment.GetName()))

		cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample")
		require.Equal(t, "PLACEHOLDER", cm.Data)

		cmi := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cmi.Data)
	})

	t.Run("Second Install Chart", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), helm.Package{
			Packages: map[string]helm.PackageSpec{
				"sample": {
					Version: "1.0.0",
					Chart:   util.NewType(base64.StdEncoding.EncodeToString(sample_1_0_0)),
				},
			},
			Releases: map[string]helm.PackageRelease{
				"sample": {
					Package: "sample",
				},
				"sample-second": {
					Package: "sample",
					Overrides: helm.Values(sharedApi.NewAnyT(t, map[string]string{
						"data": "Ov1",
					})),
				},
			},
		})

		cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample")
		require.Equal(t, "PLACEHOLDER", cm.Data)

		cmi := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cmi.Data)

		cm2 := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second")
		require.Equal(t, "Ov1", cm2.Data)

		cm2i := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cm2i.Data)
	})

	t.Run("Re-Install Second From Dump", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), executeDumpPackage(t, ns, deployment.GetName()))

		cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample")
		require.Equal(t, "PLACEHOLDER", cm.Data)

		cmi := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cmi.Data)

		cm2 := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second")
		require.Equal(t, "Ov1", cm2.Data)

		cm2i := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cm2i.Data)
	})

	t.Run("Default Change", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), helm.Package{
			Packages: map[string]helm.PackageSpec{
				"sample": {
					Version: "1.0.0",
					Chart:   util.NewType(base64.StdEncoding.EncodeToString(sample_1_0_0)),
					Overrides: helm.Values(sharedApi.NewAnyT(t, map[string]string{
						"data": "Ov2",
					})),
				},
			},
			Releases: map[string]helm.PackageRelease{
				"sample": {
					Package: "sample",
				},
				"sample-second": {
					Package: "sample",
					Overrides: helm.Values(sharedApi.NewAnyT(t, map[string]string{
						"data": "Ov1",
					})),
				},
			},
		})

		cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample")
		require.Equal(t, "Ov2", cm.Data)

		cmi := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cmi.Data)

		cm2 := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second")
		require.Equal(t, "Ov1", cm2.Data)

		cm2i := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cm2i.Data)
	})

	t.Run("Re-Install Default From Dump", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), executeDumpPackage(t, ns, deployment.GetName()))

		cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample")
		require.Equal(t, "Ov2", cm.Data)

		cmi := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cmi.Data)

		cm2 := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second")
		require.Equal(t, "Ov1", cm2.Data)

		cm2i := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second", "image")
		require.Equal(t, "gcr.io/gcr-for-testing/cicd/pause:3.5", cm2i.Data)
	})

	var importPack helm.Package

	t.Run("Import", func(t *testing.T) {
		importPack = executeImportPackage(t, registry, executeExportPackage(t, executeDumpPackage(t, ns, deployment.GetName())))

		require.Len(t, importPack.Packages, 1)
		require.Len(t, importPack.Releases, 0)
	})

	t.Run("Re-Install with imported", func(t *testing.T) {
		executeInstallPackage(t, ns, deployment.GetName(), executeDumpPackage(t, ns, deployment.GetName()), importPack)

		cm := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample")
		require.Equal(t, "Ov2", cm.Data)

		cmi := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample", "image")
		require.Equal(t, fmt.Sprintf("%s/pause:3.5", registry), cmi.Data)

		cm2 := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second")
		require.Equal(t, "Ov1", cm2.Data)

		cm2i := suite.GetConfigMap(t, client.Kubernetes(), ns, "sample", "sample-second", "image")
		require.Equal(t, fmt.Sprintf("%s/pause:3.5", registry), cm2i.Data)
	})
}
