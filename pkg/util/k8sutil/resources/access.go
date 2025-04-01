//
// DISCLAIMER
//
// Copyright 2024-2025 ArangoDB GmbH, Cologne, Germany
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
	"context"
	"fmt"
	goStrings "strings"
	"sync"

	authorization "k8s.io/api/authorization/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

type AccessRequest struct {
	Verb, Group, Version, Resource, SubResource, Name, Namespace string
}

func (a AccessRequest) Verify(ctx context.Context, client kubernetes.Interface) authorization.SubjectAccessReviewStatus {
	return VerifyAccessRequestStatus(ctx, client, a.Verb, a.Group, a.Version, a.Resource, a.SubResource, a.Name, a.Namespace)
}

func (a AccessRequest) VerifyErr(ctx context.Context, client kubernetes.Interface) error {
	res := a.Verify(ctx, client)
	if res.Allowed {
		return nil
	}
	if res.Reason != "" {
		return errors.Errorf("Unable to access %s: %s", a.String(), res.Reason)
	}

	return errors.Errorf("Unable to access %s", a.String())
}

func (a AccessRequest) String() string {
	gv := a.Version
	if a.Group != "" {
		gv = fmt.Sprintf("%s/%s", a.Group, a.Version)
	}

	res := a.Resource

	if a.SubResource != "" {
		res = fmt.Sprintf("%s/%s", a.Resource, a.SubResource)
	}

	n := a.Name

	if a.Namespace != "" {
		n = fmt.Sprintf("%s/%s", a.Namespace, a.Name)
	}

	return fmt.Sprintf("%s/%s/%s %s", gv, res, n, goStrings.ToUpper(a.Verb))
}

func VerifyAll(ctx context.Context, client kubernetes.Interface, requests ...AccessRequest) error {
	var wg sync.WaitGroup

	errs := make([]error, len(requests))

	for id := range requests {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			errs[id] = requests[id].VerifyErr(ctx, client)
		}(id)
	}

	wg.Wait()

	return errors.Errors(errs...)
}

func VerifyAccessRequestStatus(ctx context.Context, client kubernetes.Interface, verb, group, version, resource, subResource, name, namespace string) authorization.SubjectAccessReviewStatus {
	resp, err := VerifyAccessRequest(ctx, client, verb, group, version, resource, subResource, name, namespace)

	if err != nil {
		return authorization.SubjectAccessReviewStatus{
			Allowed: false,
			Reason:  fmt.Sprintf("Unable to check access: %s", err.Error()),
		}
	}

	return resp.Status
}

func VerifyAccessRequest(ctx context.Context, client kubernetes.Interface, verb, group, version, resource, subResource, name, namespace string) (*authorization.SelfSubjectAccessReview, error) {
	review := authorization.SelfSubjectAccessReview{
		Spec: authorization.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorization.ResourceAttributes{
				Namespace:   namespace,
				Verb:        verb,
				Group:       group,
				Version:     version,
				Resource:    resource,
				Subresource: subResource,
				Name:        name,
			},
		},
	}

	return client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &review, meta.CreateOptions{})
}
