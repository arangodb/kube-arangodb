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
	"fmt"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	authorization "k8s.io/api/authorization/v1"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var logger = logging.Global().RegisterAndGetLogger("crd", logging.Info)

func EnsureCRD(ctx context.Context, client kclient.Client) {
	crdsLock.Lock()
	defer crdsLock.Unlock()

	for crd, spec := range crds {
		getAccess := verifyCRDAccess(ctx, client, crd, "get")

		if !getAccess.Allowed {
			logger.Str("crd", crd).Info("Get Operations is not allowed. Continue")
			continue
		}

		c, err := client.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crd, meta.GetOptions{})
		if err != nil {
			if !errors.IsNotFound(err) {
				logger.Err(err).Str("crd", crd).Warn("Get Operations is not allowed due to error. Continue")
				continue
			}

			createAccess := verifyCRDAccess(ctx, client, crd, "create")

			if !createAccess.Allowed {
				logger.Str("crd", crd).Info("Create Operations is not allowed but CRD is missing. Continue")
				continue
			}

			c = &apiextensions.CustomResourceDefinition{
				ObjectMeta: meta.ObjectMeta{
					Name: crd,
					Labels: map[string]string{
						Version: string(spec.version),
					},
				},
				Spec: spec.spec,
			}

			if _, err := client.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions().Create(ctx, c, meta.CreateOptions{}); err != nil {
				logger.Err(err).Str("crd", crd).Warn("Create Operations is not allowed due to error. Continue")
				continue
			}

			logger.Str("crd", crd).Info("CRD Created")
			continue
		}

		updateAccess := verifyCRDAccess(ctx, client, crd, "update")

		if !updateAccess.Allowed {
			logger.Str("crd", crd).Info("Update Operations is not allowed. Continue")
			continue
		}

		if c.ObjectMeta.Labels == nil {
			c.ObjectMeta.Labels = map[string]string{}
		}

		if v, ok := c.ObjectMeta.Labels[Version]; ok {
			if v != "" {
				if !isUpdateRequired(spec.version, driver.Version(v)) {
					logger.Str("crd", crd).Info("CRD Update not required")
					continue
				}
			}
		}

		c.ObjectMeta.Labels[Version] = string(spec.version)

		c.Spec = spec.spec

		if _, err := client.KubernetesExtensions().ApiextensionsV1().CustomResourceDefinitions().Update(ctx, c, meta.UpdateOptions{}); err != nil {
			logger.Err(err).Str("crd", crd).Warn("Create Operations is not allowed due to error. Continue")
			continue
		}
		logger.Str("crd", crd).Info("CRD Updated")
	}
}

func verifyCRDAccess(ctx context.Context, client kclient.Client, crd string, verb string) authorization.SubjectAccessReviewStatus {
	r, err := verifyCRDAccessRequest(ctx, client, crd, verb)
	if err != nil {
		return authorization.SubjectAccessReviewStatus{
			Allowed: false,
			Reason:  fmt.Sprintf("Unable to check access: %s", err.Error()),
		}
	}

	return r.Status
}

func verifyCRDAccessRequest(ctx context.Context, client kclient.Client, crd string, verb string) (*authorization.SelfSubjectAccessReview, error) {
	review := authorization.SelfSubjectAccessReview{
		Spec: authorization.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorization.ResourceAttributes{
				Verb:     verb,
				Group:    "apiextensions.k8s.io",
				Version:  "v1",
				Resource: "customresourcedefinitions",
				Name:     crd,
			},
		},
	}

	return client.Kubernetes().AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &review, meta.CreateOptions{})
}
