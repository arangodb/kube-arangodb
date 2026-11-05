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

package platform

import (
	"fmt"
	"sort"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/cache"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/version"
)

// reservedRequirementKeys lists the reserved requirement keys in a fixed order so condition messages are
// stable across reconciles (map iteration order is otherwise random). Any requirement key that is not
// reserved is treated as a reference to another ArangoPlatformChart.
var reservedRequirementKeys = []string{
	platformApi.ChartRequirementArangoDB,
	platformApi.ChartRequirementOperator,
}

func isReservedRequirement(key string) bool {
	for _, k := range reservedRequirementKeys {
		if k == key {
			return true
		}
	}

	return false
}

// EvaluateRequirements evaluates a chart's platform requirements and returns the failure messages (empty
// if all are satisfied), sorted so the result is stable across reconciles. Reserved keys are evaluated
// against the bound deployment's ArangoDB version and the operator's own version; every other key names an
// ArangoPlatformChart in the same namespace whose resolved version, looked up in the shared chart cache,
// must satisfy the constraint.
func EvaluateRequirements(charts cache.ObjectCache[*platformApi.ArangoPlatformChart], namespace string, requirements platformApi.ChartDetailsPlatformRequirements, depl *api.ArangoDeployment) []string {
	failures := make([]string, 0, len(requirements))

	reserved := evaluateReservedRequirements(requirements, depl)
	for _, key := range reservedRequirementKeys {
		if err := reserved[key]; err != nil {
			failures = append(failures, fmt.Sprintf("%s: %s", key, err.Error()))
		}
	}

	customKeys := make([]string, 0, len(requirements))
	for key := range requirements {
		if isReservedRequirement(key) {
			continue
		}
		customKeys = append(customKeys, key)
	}
	sort.Strings(customKeys)

	for _, key := range customKeys {
		constraint := requirements[key]
		if constraint.Empty() {
			continue
		}

		if err := evaluateChartRequirement(charts, namespace, key, constraint); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %s", key, err.Error()))
		}
	}

	return failures
}

// evaluateChartRequirement checks a single custom chart requirement against the shared chart cache.
func evaluateChartRequirement(charts cache.ObjectCache[*platformApi.ArangoPlatformChart], namespace, name string, constraint util.VersionConstrain) error {
	chart, ok := charts.Get(operation.ItemName{Namespace: namespace, Name: name})
	if !ok {
		return fmt.Errorf("chart is not available")
	}

	var ver string
	if info := chart.Status.Info; info != nil {
		ver = info.Details.GetVersion()
	}

	if ver == "" {
		return fmt.Errorf("version is not known yet")
	}

	if valid, err := constraint.Validate(util.Version(ver)); err != nil {
		return fmt.Errorf("version %s is invalid for constraint %q: %w", ver, constraint, err)
	} else if !valid {
		return fmt.Errorf("version %s does not satisfy required %q", ver, constraint)
	}

	return nil
}

// evaluateReservedRequirements checks the reserved ArangoDB and Operator requirement constraints. The
// returned map holds one entry per declared reserved requirement: a nil value means it is fulfilled, a
// non-nil error describes why it is not. An absent or empty constraint imposes no requirement and produces
// no entry, so an empty map means fulfilled.
func evaluateReservedRequirements(requirements platformApi.ChartDetailsPlatformRequirements, depl *api.ArangoDeployment) map[string]error {
	arangoDBVersion, arangoDBKnown := deploymentArangoDBVersion(depl)

	results := make(map[string]error, len(reservedRequirementKeys))

	for _, reserved := range []struct {
		key     string
		version util.Version
		known   bool
	}{
		{platformApi.ChartRequirementArangoDB, arangoDBVersion, arangoDBKnown},
		{platformApi.ChartRequirementOperator, util.Version(version.GetVersionV1().Version), true},
	} {
		constraint, ok := requirements[reserved.key]
		if !ok || constraint.Empty() {
			continue
		}

		if !reserved.known {
			results[reserved.key] = fmt.Errorf("version is not known yet")
			continue
		}

		if valid, err := constraint.Validate(reserved.version); err != nil {
			results[reserved.key] = fmt.Errorf("version %s is invalid for constraint %q: %w", reserved.version, constraint, err)
		} else if !valid {
			results[reserved.key] = fmt.Errorf("version %s does not satisfy required %q", reserved.version, constraint)
		} else {
			results[reserved.key] = nil
		}
	}

	return results
}

// deploymentArangoDBVersion returns the ArangoDB server version reported by the deployment's current image
// and whether it is known yet (it is unset until the deployment has pulled and inspected its image).
func deploymentArangoDBVersion(depl *api.ArangoDeployment) (util.Version, bool) {
	if depl == nil || depl.Status.CurrentImage == nil {
		return "", false
	}

	return util.Version(depl.Status.CurrentImage.ArangoDBVersion), true
}
