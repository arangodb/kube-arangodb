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

package v1

import (
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/validation"
)

// KubernetesResourceName define name of kubernetes resource including validation function
type KubernetesResourceName string

// AsKubernetesResourceName formats string into AsKubernetesResourceName for validation purposes
func AsKubernetesResourceName(s *string) *KubernetesResourceName {
	if s == nil {
		return nil
	}

	value := KubernetesResourceName(*s)

	return &value
}

// StringP returns string pointer to resource name
func (n *KubernetesResourceName) StringP() *string {
	if n == nil {
		return nil
	}

	value := string(*n)

	return &value
}

// String returns string value of name
func (n *KubernetesResourceName) String() string {
	value := n.StringP()

	if value == nil {
		return ""
	}

	return *value
}

// Validate validate if name is valid kubernetes DNS_LABEL
func (n *KubernetesResourceName) Validate() error {
	if n == nil {
		return errors.Errorf("cannot be undefined")
	}

	name := *n

	if name == "" {
		return errors.Errorf("cannot be empty")
	}

	if err := IsValidName(name.String()); err != nil {
		return err
	}

	return nil
}

// Immutable verify if field changed
func (n *KubernetesResourceName) Immutable(o *KubernetesResourceName) error {
	if o == nil && n == nil {
		return nil
	}

	if o == nil || n == nil {
		return errors.Errorf("field is immutable")
	}

	if *o != *n {
		return errors.Errorf("field is immutable")
	}

	return nil
}

// IsValidName validate name to be a DNS_LABEL.
// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func IsValidName(name string) error {
	if res := validation.IsDNS1123Label(name); len(res) > 0 {
		return errors.Errorf("Validation of label failed: %s", strings.Join(res, ", "))
	}

	return nil
}
