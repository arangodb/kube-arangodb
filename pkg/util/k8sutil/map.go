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

package k8sutil

import (
	"regexp"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
)

const (
	kubernetesAnnotationMatch = ".*kubernetes\\.io/.*"
	arangoAnnotationMatch     = ".*arangodb\\.com/.*"
)

var (
	kubernetesAnnotationRegex *regexp.Regexp
	arangoAnnotationRegex     *regexp.Regexp
)

func init() {
	r, err := regexp.Compile(kubernetesAnnotationMatch)
	if err != nil {
		panic(err)
	}

	kubernetesAnnotationRegex = r

	r, err = regexp.Compile(arangoAnnotationMatch)
	if err != nil {
		panic(err)
	}

	arangoAnnotationRegex = r
}

func isFilteredBlockedAnnotation(key string) bool {
	switch key {
	case deployment.ArangoDeploymentPodRotateAnnotation:
		return true
	default:
		return false
	}
}

func filterBlockedAnnotations(m map[string]string) map[string]string {
	n := map[string]string{}

	for key, value := range m {
		if isFilteredBlockedAnnotation(key) {
			continue
		}

		n[key] = value
	}

	return n
}

// MergeAnnotations into one annotations map
func MergeAnnotations(annotations ...map[string]string) map[string]string {
	ret := map[string]string{}

	for _, annotationMap := range annotations {
		if annotationMap == nil {
			continue
		}

		for annotation, value := range annotationMap {
			ret[annotation] = value
		}
	}

	return ret
}

// IsSecuredAnnotation check if annotation key is from secured namespace
func IsSecuredAnnotation(key string) bool {
	return kubernetesAnnotationRegex.MatchString(key) || arangoAnnotationRegex.MatchString(key)
}

func GetSecuredAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		return map[string]string{}
	}

	filteredAnnotations := map[string]string{}

	for key, value := range annotations {
		if !IsSecuredAnnotation(key) {
			continue
		}

		filteredAnnotations[key] = value
	}

	return filteredAnnotations
}

func filterActualAnnotations(actual, expected map[string]string) map[string]string {
	if actual == nil {
		return nil
	}

	if expected == nil {
		expected = map[string]string{}
	}

	actualFiltered := map[string]string{}

	for key, value := range actual {
		if _, ok := expected[key]; IsSecuredAnnotation(key) && !ok {
			continue
		}

		actualFiltered[key] = value
	}

	return actualFiltered
}

// CompareAnnotations will compare annotations, but will ignore secured annotations which are present in
// actual but not specified in expected map
// It will also filter out blocked annotations
func CompareAnnotations(actual, expected map[string]string) bool {
	return compareAnnotations(filterBlockedAnnotations(actual), filterBlockedAnnotations(expected))
}

func compareAnnotations(actual, expected map[string]string) bool {
	actualFiltered := filterActualAnnotations(actual, expected)

	if actualFiltered == nil && expected == nil {
		return true
	}

	if (actualFiltered == nil && expected != nil && len(expected) == 0) ||
		(expected == nil && actualFiltered != nil && len(actualFiltered) == 0) {
		return true
	}

	if actualFiltered == nil || expected == nil {
		return false
	}

	if len(actualFiltered) != len(expected) {
		return false
	}

	for key, value := range expected {
		existingValue, existing := actualFiltered[key]

		if !existing {
			return false
		}

		if existingValue != value {
			return false
		}
	}

	return true
}
