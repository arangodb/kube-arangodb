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
// Author Adam Janikowski
//

package deployment

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/arangodb/kube-arangodb/pkg/deployment/features"

	"github.com/rs/zerolog/log"

	"github.com/arangodb/kube-arangodb/pkg/deployment/resources/inspector"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	core "k8s.io/api/core/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func runTestCases(t *testing.T, testCases ...testCaseStruct) {
	// This esure idempotency in generated outputs
	for i := 0; i < 25; i++ {
		t.Run(fmt.Sprintf("Iteration %d", i), func(t *testing.T) {
			for _, testCase := range testCases {
				runTestCase(t, testCase)
			}
		})
	}
}

func runTestCase(t *testing.T, testCase testCaseStruct) {
	t.Run(testCase.Name, func(t *testing.T) {
		// Arrange
		d, eventRecorder := createTestDeployment(testCase.config, testCase.ArangoDeployment)

		errs := 0
		for {
			cache, err := inspector.NewInspector(d.GetKubeCli(), d.GetNamespace())
			require.NoError(t, err)
			err = d.resources.EnsureSecrets(log.Logger, cache)
			if err == nil {
				break
			}

			if errs > 25 {
				require.NoError(t, err)
			}

			errs++

			if errors.IsReconcile(err) {
				continue
			}

			require.NoError(t, err)
		}

		if testCase.Helper != nil {
			testCase.Helper(t, d, &testCase)
		}

		// Create custom resource in the fake kubernetes API
		_, err := d.deps.DatabaseCRCli.DatabaseV1().ArangoDeployments(testNamespace).Create(d.apiObject)
		require.NoError(t, err)

		if testCase.Resources != nil {
			testCase.Resources(t, d)
		}

		// Set features
		{
			*features.EncryptionRotation().EnabledPointer() = testCase.Features.EncryptionRotation
			require.Equal(t, testCase.Features.EncryptionRotation, *features.EncryptionRotation().EnabledPointer())
			*features.JWTRotation().EnabledPointer() = testCase.Features.JWTRotation
			*features.TLSSNI().EnabledPointer() = testCase.Features.TLSSNI
			*features.TLSRotation().EnabledPointer() = testCase.Features.TLSRotation
		}

		// Act
		cache, err := inspector.NewInspector(d.GetKubeCli(), d.GetNamespace())
		require.NoError(t, err)
		err = d.resources.EnsurePods(cache)

		// Assert
		if testCase.ExpectedError != nil {

			if !assert.EqualError(t, err, testCase.ExpectedError.Error()) {
				println(fmt.Sprintf("%+v", err))
			}
			return
		}

		require.NoError(t, err)
		pods, err := d.deps.KubeCli.CoreV1().Pods(testNamespace).List(metav1.ListOptions{})
		require.NoError(t, err)
		require.Len(t, pods.Items, 1)
		if util.BoolOrDefault(testCase.CompareChecksum, true) {
			compareSpec(t, testCase.ExpectedPod.Spec, pods.Items[0].Spec)
		}
		require.Equal(t, testCase.ExpectedPod.Spec, pods.Items[0].Spec)
		require.Equal(t, testCase.ExpectedPod.ObjectMeta, pods.Items[0].ObjectMeta)

		if len(testCase.ExpectedEvent) > 0 {
			select {
			case msg := <-eventRecorder.Events:
				assert.Contains(t, msg, testCase.ExpectedEvent)
			default:
				assert.Fail(t, "expected event", "expected event with message '%s'", testCase.ExpectedEvent)
			}

			status, version := d.GetStatus()
			assert.Equal(t, int32(1), version)

			checkEachMember := func(group api.ServerGroup, groupSpec api.ServerGroupSpec, status *api.MemberStatusList) error {
				for _, m := range *status {
					require.Equal(t, api.MemberPhaseCreated, m.Phase)

					_, exist := m.Conditions.Get(api.ConditionTypeReady)
					require.Equal(t, false, exist)
					_, exist = m.Conditions.Get(api.ConditionTypeTerminated)
					require.Equal(t, false, exist)
					_, exist = m.Conditions.Get(api.ConditionTypeTerminating)
					require.Equal(t, false, exist)
					_, exist = m.Conditions.Get(api.ConditionTypeAgentRecoveryNeeded)
					require.Equal(t, false, exist)
					_, exist = m.Conditions.Get(api.ConditionTypeAutoUpgrade)
					require.Equal(t, false, exist)
				}
				return nil
			}

			d.GetServerGroupIterator().ForeachServerGroup(checkEachMember, &status)
		}
	})
}

func compareSpec(t *testing.T, a, b core.PodSpec) {
	ac, err := k8sutil.GetPodSpecChecksum(a)
	require.NoError(t, err)

	bc, err := k8sutil.GetPodSpecChecksum(b)
	require.NoError(t, err)

	aj, err := json.Marshal(a)
	require.NoError(t, err)

	bj, err := json.Marshal(b)
	require.NoError(t, err)

	require.Equal(t, string(aj), string(bj))
	require.Equal(t, ac, bc)
}
