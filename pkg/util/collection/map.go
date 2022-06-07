//
// DISCLAIMER
//
// Copyright 2016-2022 ArangoDB GmbH, Cologne, Germany
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

package collection

import (
	"regexp"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"
)

const (
	kubernetesIOMatch = ".*kubernetes\\.io/.*"
	arangoDBMatch     = ".*arangodb\\.com/.*"
)

var (
	kubernetesIORegex *regexp.Regexp
	arangoDBRegex     *regexp.Regexp

	reservedLabels = RestrictedList{
		k8sutil.LabelKeyArangoDeployment,
		k8sutil.LabelKeyArangoLocalStorage,
		k8sutil.LabelKeyApp,
		k8sutil.LabelKeyRole,
		k8sutil.LabelKeyArangoExporter,
	}
)

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

func Compare(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for ak, av := range a {
		bv, ok := b[ak]
		if !ok {
			return false
		}
		if av != bv {
			return false
		}
	}

	return true
}

func NewRestrictedList(param ...string) RestrictedList {
	return param
}

func ReservedLabels() RestrictedList {
	l := RestrictedList{}
	l = append(l, reservedLabels...)
	return l
}

type RestrictedList []string

func (r RestrictedList) IsRestricted(s string) bool {
	for _, i := range r {
		if match, err := regexp.MatchString(i, s); err != nil {
			continue
		} else if match {
			return true
		}

		if i == s {
			return true
		}
	}

	return false
}

func (r RestrictedList) Filter(m map[string]string) map[string]string {
	z := map[string]string{}

	for k, v := range m {
		if r.IsRestricted(k) {
			continue
		}

		z[k] = v
	}

	return z
}

func init() {
	r, err := regexp.Compile(kubernetesIOMatch)
	if err != nil {
		panic(err)
	}

	kubernetesIORegex = r

	r, err = regexp.Compile(arangoDBMatch)
	if err != nil {
		panic(err)
	}

	arangoDBRegex = r
}

func LabelsPatch(mode api.LabelsMode, expected map[string]string, actual map[string]string, ignored ...string) patch.Patch {
	return getFieldPatch(mode, "labels", expected, actual, func(k, v string) bool {
		if reservedLabels.IsRestricted(k) {
			return true
		}

		if arangoDBRegex.MatchString(k) {
			return true
		}

		if kubernetesIORegex.MatchString(k) {
			return true
		}

		if NewRestrictedList(ignored...).IsRestricted(k) {
			return true
		}

		return false
	})
}

func AnnotationsPatch(mode api.LabelsMode, expected map[string]string, actual map[string]string, ignored ...string) patch.Patch {
	return getFieldPatch(mode, "annotations", expected, actual, func(k, v string) bool {
		if kubernetesIORegex.MatchString(k) {
			return true
		}

		if arangoDBRegex.MatchString(k) {
			return true
		}

		if NewRestrictedList(ignored...).IsRestricted(k) {
			return true
		}

		return false
	})
}

func getFieldPatch(mode api.LabelsMode, section string, expected map[string]string, actual map[string]string, filtered func(k, v string) bool) patch.Patch {
	p := patch.NewPatch()

	switch mode {
	case api.LabelsDisabledMode:
		break
	case api.LabelsAppendMode:
		if len(actual) == 0 {
			return patch.NewPatch(patch.ItemReplace(patch.NewPath("metadata", section), expected))
		}
		for e, v := range expected {
			if a, ok := actual[e]; !ok {
				p.ItemAdd(patch.NewPath("metadata", section, e), v)
			} else if v != a {
				p.ItemReplace(patch.NewPath("metadata", section, e), v)
			}
		}
	case api.LabelsReplaceMode:
		if len(actual) == 0 {
			return patch.NewPatch(patch.ItemReplace(patch.NewPath("metadata", section), expected))
		}
		for e, v := range expected {
			if a, ok := actual[e]; !ok {
				p.ItemAdd(patch.NewPath("metadata", section, e), v)
			} else if v != a {
				p.ItemReplace(patch.NewPath("metadata", section, e), v)
			}
		}

		for a, v := range actual {
			if filtered != nil {
				if filtered(a, v) {
					continue
				}
			}

			if _, ok := expected[a]; !ok {
				p.ItemRemove(patch.NewPath("metadata", section, a))
			}
		}
	}

	if len(p) == 0 {
		return nil
	}

	// Add map init
	if actual == nil {
		newP := patch.NewPatch()
		newP.ItemAdd(patch.NewPath("metadata", section), []string{})
		p = append(newP, p...)
	}

	return p
}
