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

package v2

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	pbImplStorageV2SharedGCS "github.com/arangodb/kube-arangodb/integrations/storage/v2/shared/gcs"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/util/shutdown"
	"github.com/arangodb/kube-arangodb/pkg/util/tests"
)

func gcsConfiguration(t *testing.T, mods ...util.ModR[Configuration]) Configuration {
	var scfg pbImplStorageV2SharedGCS.Configuration

	scfg.Client = tests.GetGCSClientConfig(t)
	scfg.BucketName = tests.GetGCSBucket(t)
	scfg.BucketPrefix = fmt.Sprintf("tmp/unit-test/%s/", uuid.NewUUID())

	var cfg Configuration

	cfg.Type = ConfigurationTypeGCS
	cfg.GCS = scfg

	return cfg.With(mods...)
}

func gcsKubernetesObject(t *testing.T, mods ...util.Mod[platformApi.ArangoPlatformStorage]) (string, string, kclient.Client) {
	client := kclient.NewFakeClient()

	config := tests.GetGCSClientConfig(t)
	bucketName := tests.GetGCSBucket(t)
	bucketPrefix := fmt.Sprintf("tmp/unit-test-object/%s/", uuid.NewUUID())

	creds, err := client.Kubernetes().CoreV1().Secrets(tests.FakeNamespace).Create(shutdown.Context(), &core.Secret{
		ObjectMeta: meta.ObjectMeta{
			Name:      "credentials",
			Namespace: tests.FakeNamespace,
		},
		Data: map[string][]byte{
			utilConstants.SecretCredentialsServiceAccount: []byte(config.Provider.ServiceAccount.JSON),
		},
	}, meta.CreateOptions{})
	require.NoError(t, err)

	obj := &platformApi.ArangoPlatformStorage{
		ObjectMeta: meta.ObjectMeta{
			Name:      "storage",
			Namespace: tests.FakeNamespace,
		},
		Spec: platformApi.ArangoPlatformStorageSpec{
			Backend: &platformApi.ArangoPlatformStorageSpecBackend{
				GCS: &platformApi.ArangoPlatformStorageSpecBackendGCS{
					ProjectID:    util.NewType(config.ProjectID),
					BucketName:   util.NewType(bucketName),
					BucketPrefix: util.NewType(bucketPrefix),
					CredentialsSecret: &sharedApi.Object{
						Name: creds.GetName(),
					},
				},
			},
		},
	}

	util.ApplyMods(obj, mods...)

	obj, err = client.Arango().PlatformV1beta1().ArangoPlatformStorages(tests.FakeNamespace).Create(shutdown.Context(), obj, meta.CreateOptions{})
	require.NoError(t, err)

	return obj.GetName(), obj.GetNamespace(), client
}

func Test_GCS_Handler(t *testing.T) {
	testConfiguration(t, gcsConfiguration, func(in Configuration) Configuration {
		in.GCS.MaxListKeys = util.NewType[int64](32)
		return in
	})
}

func Test_GCS_Object(t *testing.T) {
	testObject(t, gcsKubernetesObject)
}
