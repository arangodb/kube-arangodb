//
// DISCLAIMER
//
// Copyright 2021 ArangoDB GmbH, Cologne, Germany
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

import "github.com/arangodb/kube-arangodb/pkg/util/errors"

func NewReconcile() Reconcile {
	return &reconcile{}
}

type Reconcile interface {
	Reconcile() error
	Required()
	IsRequired() bool
	WithError(err error) error
}

type reconcile struct {
	required bool
}

func (r *reconcile) Reconcile() error {
	if r.required {
		return errors.Reconcile()
	}

	return nil
}

func (r *reconcile) Required() {
	r.required = true
}

func (r *reconcile) IsRequired() bool {
	return r.required
}

func (r *reconcile) WithError(err error) error {
	if err == nil {
		return nil
	}

	if errors.IsReconcile(err) {
		r.Required()
		return nil
	}

	return err
}
