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

package crd

import (
	"context"
	"time"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/retry"
)

// WaitReady waits for a check to be ready.
func WaitReady(check func() error) error {
	if err := retry.Retry(check, time.Second*30); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// WaitCRDReady waits for a custom resource definition with given name to be ready.
func WaitCRDReady(clientset apiextensionsclient.Interface, crdName string) error {
	op := func() error {
		crd, err := clientset.ApiextensionsV1().CustomResourceDefinitions().Get(context.Background(), crdName, meta.GetOptions{})
		if err != nil {
			return errors.WithStack(err)
		}
		for _, cond := range crd.Status.Conditions {
			switch cond.Type {
			case apiextensionsv1.Established:
				if cond.Status == apiextensionsv1.ConditionTrue {
					return nil
				}
			case apiextensionsv1.NamesAccepted:
				if cond.Status == apiextensionsv1.ConditionFalse {
					return errors.WithStack(errors.Newf("Name conflict: %v", cond.Reason))
				}
			}
		}
		return errors.WithStack(errors.Newf("Retry needed"))
	}
	return WaitReady(op)
}
