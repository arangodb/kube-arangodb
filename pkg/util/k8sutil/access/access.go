//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package access

import (
	"context"
	"fmt"
	"sync"
	"time"

	authorization "k8s.io/api/authorization/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/cache"
	"github.com/arangodb/kube-arangodb/pkg/util/globals"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
)

var (
	accessCache     = map[string]cache.Cache[authorization.ResourceAttributes, authorization.SubjectAccessReviewStatus]{}
	accessCacheLock sync.Mutex
)

func AccessCache(client kclient.Client) cache.Cache[authorization.ResourceAttributes, authorization.SubjectAccessReviewStatus] {
	accessCacheLock.Lock()
	defer accessCacheLock.Unlock()

	if v, ok := accessCache[client.Name()]; ok {
		return v
	}

	c := cache.NewCache(accessCacheFuncGen(client))
	accessCache[client.Name()] = c
	return c
}

func accessCacheFuncGen(client kclient.Client) func(ctx context.Context, in authorization.ResourceAttributes) (authorization.SubjectAccessReviewStatus, time.Time, error) {
	return func(ctx context.Context, in authorization.ResourceAttributes) (authorization.SubjectAccessReviewStatus, time.Time, error) {
		log := logger.
			Str("Namespace", in.Namespace).
			Str("Verb", in.Verb).
			Str("Group", in.Group).
			Str("Version", in.Version).
			Str("Resource", in.Resource).
			Str("Subresource", in.Subresource).
			Str("Name", in.Name)

		log.Debug("Evaluating access")

		ctx, c := context.WithTimeout(ctx, globals.GetGlobals().Timeouts().Kubernetes().Get())
		defer c()

		review := authorization.SelfSubjectAccessReview{
			Spec: authorization.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &in,
			},
		}

		if resp, err := client.Kubernetes().AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, &review, meta.CreateOptions{}); err != nil {
			log.Err(err).Info("Access check failed")
			return authorization.SubjectAccessReviewStatus{}, util.Default[time.Time](), err
		} else {
			if IsAllowed(resp.Status) {
				log.Debug("Access allowed")
			} else {
				log.Debug("Access denied")
			}
			return resp.Status, time.Now().Add(time.Minute), nil
		}
	}
}

func VerifyAllAccess(ctx context.Context, client kclient.Client, requests ...authorization.ResourceAttributes) bool {
	return IsAllowed(VerifyAllAccessRequest(ctx, client, requests...))
}

func VerifyAllAccessRequest(ctx context.Context, client kclient.Client, requests ...authorization.ResourceAttributes) authorization.SubjectAccessReviewStatus {
	var wg sync.WaitGroup

	errs := make([]authorization.SubjectAccessReviewStatus, len(requests))

	for id := range requests {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			errs[id] = VerifyAccessRequest(ctx, client, requests[id])
		}(id)
	}

	wg.Wait()

	for _, resp := range errs {
		if !resp.Allowed {
			return authorization.SubjectAccessReviewStatus{
				Allowed: false,
				Reason:  "Not allowed by one of the requests",
			}
		}
	}

	return authorization.SubjectAccessReviewStatus{
		Allowed: true,
	}
}

func VerifyAccess(ctx context.Context, client kclient.Client, in authorization.ResourceAttributes) bool {
	return IsAllowed(VerifyAccessRequest(ctx, client, in))
}

func IsAllowed(in authorization.SubjectAccessReviewStatus) bool {
	return in.Allowed && !in.Denied
}

func VerifyAccessRequest(ctx context.Context, client kclient.Client, in authorization.ResourceAttributes) authorization.SubjectAccessReviewStatus {
	resp, err := AccessCache(client).Get(ctx, in)

	if err != nil {
		return authorization.SubjectAccessReviewStatus{
			Allowed:         false,
			Reason:          fmt.Sprintf("Unable to check access: %s", err.Error()),
			EvaluationError: err.Error(),
		}
	}

	return resp
}
