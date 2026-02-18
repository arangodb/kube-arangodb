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

package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sidecarSvcAuthzTypes "github.com/arangodb/kube-arangodb/pkg/sidecar/services/authorization/types"
	"github.com/arangodb/kube-arangodb/pkg/util"
)

type evaluator interface {
	Evaluate(name string, action, resource string, context map[string][]string, effect sidecarSvcAuthzTypes.Effect) evaluator
}

type evaluatorImpl struct {
	t      *testing.T
	name   string
	policy *policy
}

func (e evaluatorImpl) Evaluate(name string, action, resource string, context map[string][]string, effect sidecarSvcAuthzTypes.Effect) evaluator {
	e.t.Run(fmt.Sprintf("%s/%s", e.name, name), func(t *testing.T) {
		res, err := e.policy.Evaluate(action, resource, context)
		if err != nil {
			if sidecarSvcAuthzTypes.IsPermissionDenied(err) {
				require.Equal(t, effect, sidecarSvcAuthzTypes.Effect_Deny)
				return
			}

			require.NoError(t, err)
		}

		require.Equal(t, effect, util.BoolSwitch(res, sidecarSvcAuthzTypes.Effect_Allow, sidecarSvcAuthzTypes.Effect_Deny))
	})

	return e
}

func statementEvaluator(t *testing.T, name string, statements ...*sidecarSvcAuthzTypes.PolicyStatement) evaluator {
	pol, err := newPolicy(&sidecarSvcAuthzTypes.Policy{Statements: statements})
	require.NoError(t, err)

	return evaluatorImpl{
		t:      t,
		policy: &pol,
		name:   name,
	}
}

func Test_PolicyEvaluation_Actions(t *testing.T) {
	statementEvaluator(t, "Empty").
		Evaluate(
			"Ensure default deny",
			"test:GetAll",
			"content:file:/data",
			nil,
			sidecarSvcAuthzTypes.Effect_Deny,
		)

	statementEvaluator(t, "Ensure exact grant", &sidecarSvcAuthzTypes.PolicyStatement{
		Effect:    sidecarSvcAuthzTypes.Effect_Allow,
		Resources: []string{"content:file:/data"},
		Actions:   []string{"test:GetAll"},
	}).Evaluate(
		"Exact Granted",
		"test:GetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Same namespace & prefix granted",
		"test:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Deny,
	).Evaluate(
		"Same namespace granted",
		"test:SetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Deny,
	).Evaluate(
		"Different namespace granted",
		"test2:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Deny,
	)

	statementEvaluator(t, "Ensure wildcard grant", &sidecarSvcAuthzTypes.PolicyStatement{
		Effect:    sidecarSvcAuthzTypes.Effect_Allow,
		Resources: []string{"content:file:/data"},
		Actions:   []string{"test:*"},
	}).Evaluate(
		"Exact granted",
		"test:GetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Same namespace & prefix granted",
		"test:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Same namespace granted",
		"test:SetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Different namespace granted",
		"test2:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Deny,
	)

	statementEvaluator(t, "Ensure super wildcard grant", &sidecarSvcAuthzTypes.PolicyStatement{
		Effect:    sidecarSvcAuthzTypes.Effect_Allow,
		Resources: []string{"content:file:/data"},
		Actions:   []string{"*"},
	}).Evaluate(
		"Exact granted",
		"test:GetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Same namespace & prefix granted",
		"test:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Same namespace granted",
		"test:SetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Different namespace granted",
		"test2:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	)

	statementEvaluator(t, "Ensure super wildcard grant with deny", &sidecarSvcAuthzTypes.PolicyStatement{
		Effect:    sidecarSvcAuthzTypes.Effect_Allow,
		Resources: []string{"content:file:/data"},
		Actions:   []string{"*"},
	}, &sidecarSvcAuthzTypes.PolicyStatement{
		Effect:    sidecarSvcAuthzTypes.Effect_Deny,
		Resources: []string{"content:file:/data"},
		Actions:   []string{"test:GetNotAll"},
	}).Evaluate(
		"Exact granted",
		"test:GetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Same namespace & prefix denied",
		"test:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Deny,
	).Evaluate(
		"Same namespace granted",
		"test:SetAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	).Evaluate(
		"Different namespace granted",
		"test2:GetNotAll",
		"content:file:/data",
		nil,
		sidecarSvcAuthzTypes.Effect_Allow,
	)
}
