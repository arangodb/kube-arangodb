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
// Author tomasz@arangodb.con
//

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arangodb/go-driver"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/client"
	"github.com/arangodb/kube-arangodb/pkg/util/arangod"
	"github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
	"github.com/dchest/uniuri"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestSecretHashesRootUser checks if Status.SecretHashes.Users[root] changed after request for it
func TestSecretHashesRootUser(t *testing.T) {
	longOrSkip(t)
	c := client.MustNewClient()
	kubecli := mustNewKubeClient(t)
	ns := getNamespace(t)

	// Prepare deployment config
	depl := newDeployment("test-auth-sng-def-" + uniuri.NewLen(4))
	depl.Spec.Mode = api.NewMode(api.DeploymentModeSingle)
	depl.Spec.SetDefaults(depl.GetName())
	depl.Spec.Bootstrap.PasswordSecretNames[api.UserNameRoot] = api.PasswordSecretNameAuto

	// Create deployment
	apiObject, err := c.DatabaseV1().ArangoDeployments(ns).Create(depl)
	if err != nil {
		t.Fatalf("Create deployment failed: %v", err)
	}
	defer deferedCleanupDeployment(c, depl.GetName(), ns)

	// Wait for deployment to be ready
	depl, err = waitUntilDeployment(c, depl.GetName(), ns, deploymentIsReady())
	if err != nil {
		t.Fatalf("Deployment not running in time: %v", err)
	}

	// Create a database client
	ctx := arangod.WithRequireAuthentication(context.Background())
	client := mustNewArangodDatabaseClient(ctx, kubecli, apiObject, t, nil)

	// Wait for single server available
	if err := waitUntilVersionUp(client, nil); err != nil {
		t.Fatalf("Single server not running returning version in time: %v", err)
	}

	depl, err = waitUntilDeployment(c, depl.GetName(), ns, func(obj *api.ArangoDeployment) error {
		// check if root secret password is set
		secretHashes := obj.Status.SecretHashes
		if secretHashes == nil {
			return fmt.Errorf("field Status.SecretHashes is not set")
		}

		if secretHashes.Users == nil {
			return fmt.Errorf("field Status.SecretHashes.Users is not set")
		}

		if hash, ok := secretHashes.Users[api.UserNameRoot]; !ok {
			return fmt.Errorf("field Status.SecretHashes.Users[root] is not set")
		} else if len(hash) == 0 {
			return fmt.Errorf("field Status.SecretHashes.Users[root] is empty")
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Deployment is not set properly: %v", err)
	}
	rootHashSecret := depl.Status.SecretHashes.Users[api.UserNameRoot]

	secretRootName := string(depl.Spec.Bootstrap.PasswordSecretNames[api.UserNameRoot])
	secretRoot, err := waitUntilSecret(kubecli, secretRootName, ns, nil, time.Second)
	if err != nil {
		t.Fatalf("Root secret '%s' not found: %v", secretRootName, err)
	}

	secretRoot.Data[constants.SecretPassword] = []byte("1")
	_, err = kubecli.CoreV1().Secrets(ns).Update(secretRoot)
	if err != nil {
		t.Fatalf("Root secret '%s' has not been changed: %v", secretRootName, err)
	}

	err = retry.Retry(func() error {
		// check if root secret hash has changed
		depl, err = c.DatabaseV1().ArangoDeployments(ns).Get(depl.GetName(), metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get deployment: %v", err)
		}

		if rootHashSecret == depl.Status.SecretHashes.Users[api.UserNameRoot] {
			return maskAny(errors.New("field Status.SecretHashes.Users[root] has not been changed yet"))
		}
		return nil
	}, deploymentReadyTimeout)
	if err != nil {
		t.Fatalf("%v", err)
	}

	// Check if password changed
	auth := driver.BasicAuthentication(api.UserNameRoot, "1")
	_, err = client.Connection().SetAuthentication(auth)
	if err != nil {
		t.Fatalf("The password for user '%s' has not been changed: %v", api.UserNameRoot, err)
	}
	_, err = client.Version(context.Background())
	if err != nil {
		t.Fatalf("can not get version after the password has been changed")
	}

	//Cleanup
	removeDeployment(c, depl.GetName(), ns)
}
