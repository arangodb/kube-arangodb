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

package inspector

import (
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/poddisruptionbudget"
	policy "k8s.io/api/policy/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func (i *inspector) IteratePodDisruptionBudgets(action poddisruptionbudget.PodDisruptionBudgetAction, filters ...poddisruptionbudget.PodDisruptionBudgetFilter) error {
	for _, podDisruptionBudget := range i.PodDisruptionBudgets() {
		if err := i.iteratePodDisruptionBudget(podDisruptionBudget, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iteratePodDisruptionBudget(podDisruptionBudget *policy.PodDisruptionBudget, action poddisruptionbudget.PodDisruptionBudgetAction, filters ...poddisruptionbudget.PodDisruptionBudgetFilter) error {
	for _, filter := range filters {
		if !filter(podDisruptionBudget) {
			return nil
		}
	}

	return action(podDisruptionBudget)
}

func (i *inspector) PodDisruptionBudgets() []*policy.PodDisruptionBudget {
	i.lock.Lock()
	defer i.lock.Unlock()

	var r []*policy.PodDisruptionBudget
	for _, podDisruptionBudget := range i.podDisruptionBudgets {
		r = append(r, podDisruptionBudget)
	}

	return r
}

func (i *inspector) PodDisruptionBudget(name string) (*policy.PodDisruptionBudget, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	podDisruptionBudget, ok := i.podDisruptionBudgets[name]
	if !ok {
		return nil, false
	}

	return podDisruptionBudget, true
}

func podDisruptionBudgetsToMap(k kubernetes.Interface, namespace string) (map[string]*policy.PodDisruptionBudget, error) {
	podDisruptionBudgets, err := getPodDisruptionBudgets(k, namespace, "")
	if err != nil {
		return nil, err
	}

	podDisruptionBudgetMap := map[string]*policy.PodDisruptionBudget{}

	for _, podDisruptionBudget := range podDisruptionBudgets {
		_, exists := podDisruptionBudgetMap[podDisruptionBudget.GetName()]
		if exists {
			return nil, errors.Newf("PodDisruptionBudget %s already exists in map, error received", podDisruptionBudget.GetName())
		}

		podDisruptionBudgetMap[podDisruptionBudget.GetName()] = podDisruptionBudgetPointer(podDisruptionBudget)
	}

	return podDisruptionBudgetMap, nil
}

func podDisruptionBudgetPointer(podDisruptionBudget policy.PodDisruptionBudget) *policy.PodDisruptionBudget {
	return &podDisruptionBudget
}

func getPodDisruptionBudgets(k kubernetes.Interface, namespace, cont string) ([]policy.PodDisruptionBudget, error) {
	podDisruptionBudgets, err := k.PolicyV1beta1().PodDisruptionBudgets(namespace).List(meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})

	if err != nil {
		return nil, err
	}

	if podDisruptionBudgets.Continue != "" {
		nextPodDisruptionBudgetsLayer, err := getPodDisruptionBudgets(k, namespace, podDisruptionBudgets.Continue)
		if err != nil {
			return nil, err
		}

		return append(podDisruptionBudgets.Items, nextPodDisruptionBudgetsLayer...), nil
	}

	return podDisruptionBudgets.Items, nil
}

func FilterPodDisruptionBudgetsByLabels(labels map[string]string) poddisruptionbudget.PodDisruptionBudgetFilter {
	return func(podDisruptionBudget *policy.PodDisruptionBudget) bool {
		for key, value := range labels {
			v, ok := podDisruptionBudget.Labels[key]
			if !ok {
				return false
			}

			if v != value {
				return false
			}
		}

		return true
	}
}
