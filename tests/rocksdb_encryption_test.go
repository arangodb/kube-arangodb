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
// Author Ewout Prangsma
//

package tests

import (
	"context"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dchest/uniuri"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1alpha"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

// TestRocksDBEncryptionSingle tests the creating of a single server deployment
// with RocksDB & Encryption.
func TestRocksDBEncryptionSingle(t *testing.T) {
	longOrSkip(t)
	image := getEnterpriseImageOrSkip(t)
	c := client.MustNewInCluster()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepull enterprise images
	assert.NoError(t, prepullArangoImage(kubecli, image, ns))

	// Prepare deployment config
	depl := newDeployment("test-rocksdb-enc-sng-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.Image = util.NewString(image)
	depl.Spec.StorageEngine = api.NewStorageEngine(api.StorageEngineRocksDB)
	depl.Spec.RocksDB.Encryption.KeySecretName = util.NewString(strings.ToLower(uniuri.New()))

	// Create encryption key secret
	key := make([]byte, 32)
	rand.Read(key)
	if err := k8sutil.CreateEncryptionKeySecret(kubecli.CoreV1(), depl.Spec.RocksDB.Encryption.GetKeySecretName(), ns, key); err != nil {
		t.Fatalf("Create encryption key secret failed: %v", err)
	}

	// Create deployment
	_, err := c.DatabaseV1alpha().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	apiObject, err := waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create database client
	ctx := context.Background()
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	// Cleanup
	removeDeployment(c, depl.GetName(), ns)
	removeSecret(kubecli, depl.Spec.RocksDB.Encryption.GetKeySecretName(), ns)
}
