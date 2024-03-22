//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

package resources

import (
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func SelectLabels(selector *meta.LabelSelector, labels map[string]string) bool {
	if selector == nil {
		return false
	}

	for k, v := range selector.MatchLabels {
		if v2, ok := labels[k]; !ok || v2 != v {
			return false
		}
	}

	for _, req := range selector.MatchExpressions {
		switch req.Operator {
		case meta.LabelSelectorOpIn:
			if len(req.Values) == 0 {
				return false
			}

			if v, ok := labels[req.Key]; !ok {
				return false
			} else if !strings.ListContains(req.Values, v) {
				return false
			}
		case meta.LabelSelectorOpNotIn:
			if len(req.Values) == 0 {
				return false
			}

			if v, ok := labels[req.Key]; ok {
				if strings.ListContains(req.Values, v) {
					return false
				}
			}
		case meta.LabelSelectorOpExists:
			if _, ok := labels[req.Key]; !ok {
				return false
			}
		case meta.LabelSelectorOpDoesNotExist:
			if _, ok := labels[req.Key]; ok {
				return false
			}
		default:
			return false
		}
	}

	return true
}
