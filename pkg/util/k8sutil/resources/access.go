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
	"context"
	"fmt"

	authorization "k8s.io/api/authorization/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

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
