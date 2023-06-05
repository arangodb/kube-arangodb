//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	"github.com/arangodb/kube-arangodb/pkg/deployment/resources"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector"
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
		if testCase.config.OperatorImage == "" {
			testCase.config.OperatorImage = testImageOperator
		}

		d, eventRecorder := createTestDeployment(t, testCase.config, testCase.ArangoDeployment)

		startDepl := d.GetStatus()

		errs := 0
		for {
			require.NoError(t, d.acs.CurrentClusterCache().Refresh(context.Background()))
			err := d.resources.EnsureSecrets(context.Background(), d.GetCachedStatus())
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

		f := startDepl.Members.AsList()
		if len(f) == 0 {
			f = d.GetStatus().Members.AsList()
		}

		// Add Expected pod defaults
		if !testCase.DropInit {
			testCase.ExpectedPod = *defaultPodAppender(t, &testCase.ExpectedPod,
				addLifecycle(f[0].Member.ID,
					f[0].Group == api.ServerGroupDBServers && f[0].Member.IsInitialized,
					testCase.ArangoDeployment.Spec.License.GetSecretName(),
					f[0].Group),
				podDataSort())
		}

		// Create custom resource in the fake kubernetes API
		_, err := d.deps.Client.Arango().DatabaseV1().ArangoDeployments(testNamespace).Create(context.Background(), d.currentObject, meta.CreateOptions{})
		require.NoError(t, err)

		if testCase.Resources != nil {
			testCase.Resources(t, d)
		}

		// Set features
		{
			*features.EncryptionRotation().EnabledPointer() = testCase.Features.EncryptionRotation
			*features.Version310().EnabledPointer() = testCase.Features.Version310
			require.Equal(t, testCase.Features.EncryptionRotation, *features.EncryptionRotation().EnabledPointer())
			*features.JWTRotation().EnabledPointer() = testCase.Features.JWTRotation
			*features.TLSSNI().EnabledPointer() = testCase.Features.TLSSNI
			if g := testCase.Features.Graceful; g != nil {
				*features.GracefulShutdown().EnabledPointer() = *g
			} else {
				*features.GracefulShutdown().EnabledPointer() = features.GracefulShutdown().EnabledByDefault()
			}
			*features.TLSRotation().EnabledPointer() = testCase.Features.TLSRotation
		}

		// Set Pending phase
		for _, e := range d.GetStatus().Members.AsList() {
			m := e.Member
			if m.Phase == api.MemberPhaseNone {
				m.Phase = api.MemberPhasePending
				require.NoError(t, d.currentObjectStatus.Members.Update(m, e.Group))
			}
		}

		// Set members
		var loopErr error
		for _, e := range d.GetStatus().Members.AsList() {
			m := e.Member
			group := e.Group
			member := api.ArangoMember{
				ObjectMeta: meta.ObjectMeta{
					Namespace: d.GetNamespace(),
					Name:      m.ArangoMemberName(d.GetName(), group),
				},
				Spec: api.ArangoMemberSpec{
					Group: group,
					ID:    m.ID,
				},
			}

			if _, loopErr = d.acs.CurrentClusterCache().ArangoMemberModInterface().V1().Create(context.Background(), &member, meta.CreateOptions{}); loopErr != nil {
				break
			}

			s := core.Service{
				ObjectMeta: meta.ObjectMeta{
					Name:      member.GetName(),
					Namespace: member.GetNamespace(),
				},
			}

			if _, loopErr = d.ServicesModInterface().Create(context.Background(), &s, meta.CreateOptions{}); loopErr != nil {
				break
			}

			require.NoError(t, d.acs.CurrentClusterCache().Refresh(context.Background()))

			groupSpec := d.GetSpec().GetServerGroupSpec(group)

			image, ok := d.resources.SelectImage(d.GetSpec(), d.GetStatus())
			require.True(t, ok)

			var template *core.PodTemplateSpec
			template, loopErr = d.resources.RenderPodTemplateForMember(context.Background(), d.ACS(), d.GetSpec(), d.GetStatus(), m.ID, image)
			if loopErr != nil {
				break
			}

			checksum, err := resources.ChecksumArangoPod(groupSpec, resources.CreatePodFromTemplate(template))
			require.NoError(t, err)

			podTemplate, err := api.GetArangoMemberPodTemplate(template, checksum)
			require.NoError(t, err)

			member.Status.Template = podTemplate
			member.Spec.Template = podTemplate

			if loopErr = inspector.WithArangoMemberUpdate(context.Background(), d.acs.CurrentClusterCache(), member.GetName(), func(in *api.ArangoMember) (bool, error) {
				in.Spec.Template = podTemplate
				return true, nil
			}); loopErr != nil {
				break
			}

			if loopErr = inspector.WithArangoMemberStatusUpdate(context.Background(), d.acs.CurrentClusterCache(), member.GetName(), func(in *api.ArangoMember) (bool, error) {
				in.Status.Template = podTemplate
				return true, nil
			}); loopErr != nil {
				break
			}
		}
		if loopErr != nil && testCase.ExpectedError != nil && assert.EqualError(t, loopErr, testCase.ExpectedError.Error()) {
			return
		}
		require.NoError(t, err)

		// Act
		require.NoError(t, d.acs.CurrentClusterCache().Refresh(context.Background()))
		err = d.resources.EnsurePods(context.Background(), d.GetCachedStatus())

		// Assert
		if testCase.ExpectedError != nil {

			if !assert.EqualError(t, err, testCase.ExpectedError.Error()) {
				println(fmt.Sprintf("%+v", err))
			}
			return
		}

		require.NoError(t, err)
		pods, err := d.deps.Client.Kubernetes().CoreV1().Pods(testNamespace).List(context.Background(), meta.ListOptions{})
		require.NoError(t, err)
		require.Len(t, pods.Items, 1)
		if util.TypeOrDefault[bool](testCase.CompareChecksum, true) {
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

			status := d.GetStatus()

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

					require.NotNil(t, m.Image)
					require.True(t, m.Image.Equal(d.currentObject.Status.CurrentImage))
				}
				return nil
			}

			d.GetServerGroupIterator().ForeachServerGroupAccepted(checkEachMember, &status)
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
