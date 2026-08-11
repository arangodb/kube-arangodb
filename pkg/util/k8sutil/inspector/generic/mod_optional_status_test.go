//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

//go:build testing

package generic_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	apiv2alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v2alpha1"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient/external"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
)

const (
	arangoDeploymentCRDName   = "arangodeployments.database.arangodb.com"
	arangoDeploymentsResource = "arangodeployments"
)

// staticSubresources is a fixed SubresourceInspector used to drive the wrapper into each mode.
type staticSubresources map[generic.Subresource]bool

func (s staticSubresources) SubresourceEnabled(subresource generic.Subresource) bool {
	return s[subresource]
}

// Test_ModOptionalStatusClient exercises every routing mode of ModOptionalStatusClient against a real
// apiserver (runs only when K3D_ENABLED/TEST_KUBECONFIG points at a cluster - see external.ExternalClient).
//
// It relies on the ArangoDeployment CRD exposing the status subresource ONLY on v2alpha1 and NOT on v1:
//   - v1, inspector reports status disabled     -> writes go straight to the main endpoint.
//   - v2alpha1, inspector reports status enabled -> writes go to the served /status subresource.
//   - v1, inspector reports status enabled       -> the /status subresource 404s and we fall back to
//     the main endpoint.
//
// In every mode the status write must succeed and persist.
func Test_ModOptionalStatusClient(t *testing.T) {
	client, ns := external.ExternalClient(t)

	enabled := staticSubresources{generic.SubresourceStatus: true}
	disabled := staticSubresources{}

	deplsV1 := client.Arango().DatabaseV1().ArangoDeployments(ns)
	deplsV2 := client.Arango().DatabaseV2alpha1().ArangoDeployments(ns)

	// EnsureCRD does not wait for the CRD to be established, and the k3d cluster is freshly created for
	// each run - wait until both served versions actually accept requests before writing.
	waitServable(t, func() error {
		_, err := deplsV1.List(shutdown.Context(), meta.ListOptions{})
		return err
	})
	waitServable(t, func() error {
		_, err := deplsV2.List(shutdown.Context(), meta.ListOptions{})
		return err
	})

	t.Run("Disabled - no subresource, write via main endpoint", func(t *testing.T) {
		depls := deplsV1
		wrapped := generic.NewModOptionalStatusClient[*api.ArangoDeployment](disabled, depls)

		runStatusWrites(t, wrapped,
			&api.ArangoDeployment{
				ObjectMeta: meta.ObjectMeta{GenerateName: "test-disabled-"},
				Spec:       api.DeploymentSpec{Mode: api.NewMode(api.DeploymentModeSingle)},
			},
			func(o *api.ArangoDeployment, v string) { o.Status.AppliedVersion = v },
			func(o *api.ArangoDeployment) string { return o.Status.AppliedVersion },
			func(name string) (*api.ArangoDeployment, error) {
				return depls.Get(shutdown.Context(), name, meta.GetOptions{})
			},
		)
	})

	t.Run("Enabled - subresource served, write via /status", func(t *testing.T) {
		depls := deplsV2
		wrapped := generic.NewModOptionalStatusClient[*apiv2alpha1.ArangoDeployment](enabled, depls)

		runStatusWrites(t, wrapped,
			&apiv2alpha1.ArangoDeployment{
				ObjectMeta: meta.ObjectMeta{GenerateName: "test-enabled-"},
				Spec:       apiv2alpha1.DeploymentSpec{Mode: apiv2alpha1.NewMode(apiv2alpha1.DeploymentModeSingle)},
			},
			func(o *apiv2alpha1.ArangoDeployment, v string) { o.Status.AppliedVersion = v },
			func(o *apiv2alpha1.ArangoDeployment) string { return o.Status.AppliedVersion },
			func(name string) (*apiv2alpha1.ArangoDeployment, error) {
				return depls.Get(shutdown.Context(), name, meta.GetOptions{})
			},
		)
	})

	t.Run("Enabled but subresource absent - fall back to main endpoint", func(t *testing.T) {
		depls := deplsV1
		wrapped := generic.NewModOptionalStatusClient[*api.ArangoDeployment](enabled, depls)

		runStatusWrites(t, wrapped,
			&api.ArangoDeployment{
				ObjectMeta: meta.ObjectMeta{GenerateName: "test-fallback-"},
				Spec:       api.DeploymentSpec{Mode: api.NewMode(api.DeploymentModeSingle)},
			},
			func(o *api.ArangoDeployment, v string) { o.Status.AppliedVersion = v },
			func(o *api.ArangoDeployment) string { return o.Status.AppliedVersion },
			func(name string) (*api.ArangoDeployment, error) {
				return depls.Get(shutdown.Context(), name, meta.GetOptions{})
			},
		)
	})
}

// runStatusWrites creates obj, then writes its status via both UpdateStatus and PatchStatus, asserting
// each call succeeds and is actually persisted on a fresh read - regardless of which endpoint the
// wrapper routed the write to.
func runStatusWrites[S meta.Object](
	t *testing.T,
	wrapped generic.ModOptionalStatusClient[S],
	obj S,
	setApplied func(o S, v string),
	getApplied func(o S) string,
	get func(name string) (S, error),
) {
	created, err := wrapped.Create(shutdown.Context(), obj, meta.CreateOptions{})
	require.NoError(t, err)

	name := created.GetName()
	t.Cleanup(func() {
		_ = wrapped.Delete(context.Background(), name, meta.DeleteOptions{})
	})

	// UpdateStatus.
	setApplied(created, "via-update-status")
	updated, err := wrapped.UpdateStatus(shutdown.Context(), created, meta.UpdateOptions{})
	require.NoError(t, err)
	require.Equal(t, "via-update-status", getApplied(updated))

	got, err := get(name)
	require.NoError(t, err)
	require.Equal(t, "via-update-status", getApplied(got), "UpdateStatus was not persisted")

	// PatchStatus.
	patched, err := wrapped.PatchStatus(shutdown.Context(), name, types.MergePatchType,
		[]byte(`{"status":{"appliedVersion":"via-patch-status"}}`), meta.PatchOptions{})
	require.NoError(t, err)
	require.Equal(t, "via-patch-status", getApplied(patched))

	got, err = get(name)
	require.NoError(t, err)
	require.Equal(t, "via-patch-status", getApplied(got), "PatchStatus was not persisted")
}

// Test_ModOptionalStatusClient_Discovery drives NewCachedSubresourceInspector off real Kubernetes
// discovery (ServerResourcesForGroupVersion) and enables the ArangoDeployment v1 status subresource
// in the middle of the run, asserting that:
//   - discovery reports the correct subresource state before and after the change,
//   - an inspector built before the change keeps its cached (stale) view until SubresourceCacheTTL,
//   - an inspector built after the change reports the subresource as enabled,
//   - the wrapper writes status successfully in both states (main endpoint, then /status).
//
// Runs only when K3D_ENABLED/TEST_KUBECONFIG points at a cluster (see external.ExternalClient).
func Test_ModOptionalStatusClient_Discovery(t *testing.T) {
	client, ns := external.ExternalClient(t)

	gv := api.SchemeGroupVersion.String()
	fetch := discoverStatusSubresource(client, gv)

	// Discovery must report NO status subresource on v1 to start with.
	subs, err := fetch(shutdown.Context())
	require.NoError(t, err)
	require.NotContains(t, subs, generic.SubresourceStatus, "v1 must not expose the status subresource initially")

	// A cached inspector built now reflects that, and the wrapper falls back to the main endpoint.
	before := generic.NewCachedSubresourceInspector(fetch)
	require.False(t, before.SubresourceEnabled(generic.SubresourceStatus))

	depls := client.Arango().DatabaseV1().ArangoDeployments(ns)

	// EnsureCRD does not wait for establishment - make sure v1 is servable before writing.
	waitServable(t, func() error {
		_, err := depls.List(shutdown.Context(), meta.ListOptions{})
		return err
	})

	runStatusWrites(t, generic.NewModOptionalStatusClient[*api.ArangoDeployment](before, depls),
		&api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{GenerateName: "test-disco-before-"},
			Spec:       api.DeploymentSpec{Mode: api.NewMode(api.DeploymentModeSingle)},
		},
		func(o *api.ArangoDeployment, v string) { o.Status.AppliedVersion = v },
		func(o *api.ArangoDeployment) string { return o.Status.AppliedVersion },
		func(name string) (*api.ArangoDeployment, error) {
			return depls.Get(shutdown.Context(), name, meta.GetOptions{})
		},
	)

	// Enable the status subresource on v1 in the middle of the run.
	enableV1StatusSubresource(t, client)

	// Discovery must pick up the change, and the endpoint must be servable again (mutating the CRD
	// re-establishes the resource, so a create can briefly 404 in the meantime).
	require.Eventually(t, func() bool {
		subs, err := fetch(shutdown.Context())
		if err != nil || !containsSubresource(subs, generic.SubresourceStatus) {
			return false
		}
		_, err = depls.List(shutdown.Context(), meta.ListOptions{})
		return err == nil
	}, time.Minute, time.Second, "discovery did not report the added status subresource")

	// The inspector built before the change keeps its cached view until SubresourceCacheTTL elapses.
	require.False(t, before.SubresourceEnabled(generic.SubresourceStatus),
		"inspector built before the change must not flip before SubresourceCacheTTL")

	// A freshly built inspector reflects the new state, and the wrapper now targets the /status subresource.
	after := generic.NewCachedSubresourceInspector(fetch)
	require.True(t, after.SubresourceEnabled(generic.SubresourceStatus))

	runStatusWrites(t, generic.NewModOptionalStatusClient[*api.ArangoDeployment](after, depls),
		&api.ArangoDeployment{
			ObjectMeta: meta.ObjectMeta{GenerateName: "test-disco-after-"},
			Spec:       api.DeploymentSpec{Mode: api.NewMode(api.DeploymentModeSingle)},
		},
		func(o *api.ArangoDeployment, v string) { o.Status.AppliedVersion = v },
		func(o *api.ArangoDeployment) string { return o.Status.AppliedVersion },
		func(name string) (*api.ArangoDeployment, error) {
			return depls.Get(shutdown.Context(), name, meta.GetOptions{})
		},
	)
}

// discoverStatusSubresource returns a fetch function that lists, via Kubernetes discovery, the subresources
// exposed by the arangodeployments resource in the given group/version.
func discoverStatusSubresource(client kclient.Client, gv string) func(context.Context) ([]generic.Subresource, error) {
	return func(context.Context) ([]generic.Subresource, error) {
		list, err := client.Kubernetes().Discovery().ServerResourcesForGroupVersion(gv)
		if err != nil {
			return nil, err
		}

		var subs []generic.Subresource
		for _, r := range list.APIResources {
			if r.Name == arangoDeploymentsResource+"/"+generic.SubresourceStatus.String() {
				subs = append(subs, generic.SubresourceStatus)
			}
		}
		return subs, nil
	}
}

// waitServable blocks until probe returns no error, giving the apiserver time to (re)establish the
// resource endpoint after a CRD install or mutation - EnsureCRD does not wait for it and the k3d cluster
// is recreated for every run.
func waitServable(t *testing.T, probe func() error) {
	require.Eventually(t, func() bool {
		return probe() == nil
	}, 30*time.Second, 200*time.Millisecond, "resource endpoint did not become servable")
}

func containsSubresource(subs []generic.Subresource, subresource generic.Subresource) bool {
	for _, s := range subs {
		if s == subresource {
			return true
		}
	}
	return false
}

// enableV1StatusSubresource enables the status subresource on the v1 version of the ArangoDeployment CRD
// and registers a cleanup that removes it again.
func enableV1StatusSubresource(t *testing.T, client kclient.Client) {
	crds := client.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions()

	setV1Subresources := func(subresources *apiextv1.CustomResourceSubresources) {
		crd, err := crds.Get(shutdown.Context(), arangoDeploymentCRDName, meta.GetOptions{})
		require.NoError(t, err)

		for i := range crd.Spec.Versions {
			if crd.Spec.Versions[i].Name == api.ArangoDeploymentVersion {
				crd.Spec.Versions[i].Subresources = subresources
			}
		}

		_, err = crds.Update(shutdown.Context(), crd, meta.UpdateOptions{})
		require.NoError(t, err)
	}

	t.Cleanup(func() {
		setV1Subresources(nil)
	})

	setV1Subresources(&apiextv1.CustomResourceSubresources{
		Status: &apiextv1.CustomResourceSubresourceStatus{},
	})
}
