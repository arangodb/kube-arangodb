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
// Author Tomasz Mielech
//

package inspector

import (
	"context"

	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil"

	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/arangomember"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (i *inspector) IterateArangoMembers(action arangomember.ArangoMemberAction, filters ...arangomember.ArangoMemberFilter) error {
	for _, arangoMember := range i.ArangoMembers() {
		if err := i.iterateArangoMembers(arangoMember, action, filters...); err != nil {
			return err
		}
	}
	return nil
}

func (i *inspector) iterateArangoMembers(arangoMember *api.ArangoMember, action arangomember.ArangoMemberAction, filters ...arangomember.ArangoMemberFilter) error {
	for _, filter := range filters {
		if !filter(arangoMember) {
			return nil
		}
	}

	return action(arangoMember)
}

func (i *inspector) ArangoMembers() []*api.ArangoMember {
	i.lock.Lock()
	defer i.lock.Unlock()

	var r []*api.ArangoMember
	for _, arangoMember := range i.arangoMembers {
		r = append(r, arangoMember)
	}

	return r
}

func (i *inspector) ArangoMember(name string) (*api.ArangoMember, bool) {
	i.lock.Lock()
	defer i.lock.Unlock()

	arangoMember, ok := i.arangoMembers[name]
	if !ok {
		return nil, false
	}

	return arangoMember, true
}

func arangoMembersToMap(ctx context.Context, k versioned.Interface, namespace string) (map[string]*api.ArangoMember, error) {
	arangoMembers, err := getArangoMembers(ctx, k, namespace, "")
	if err != nil {
		return nil, err
	}

	arangoMemberMap := map[string]*api.ArangoMember{}

	for _, arangoMember := range arangoMembers {
		_, exists := arangoMemberMap[arangoMember.GetName()]
		if exists {
			return nil, errors.Newf("ArangoMember %s already exists in map, error received", arangoMember.GetName())
		}

		arangoMemberMap[arangoMember.GetName()] = arangoMemberPointer(arangoMember)
	}

	return arangoMemberMap, nil
}

func arangoMemberPointer(pod api.ArangoMember) *api.ArangoMember {
	return &pod
}

func getArangoMembers(ctx context.Context, k versioned.Interface, namespace, cont string) ([]api.ArangoMember, error) {
	ctxChild, cancel := context.WithTimeout(ctx, k8sutil.GetRequestTimeout())
	arangoMembers, err := k.DatabaseV1().ArangoMembers(namespace).List(ctxChild, meta.ListOptions{
		Limit:    128,
		Continue: cont,
	})
	cancel()

	if err != nil {
		return nil, err
	}

	if arangoMembers.Continue != "" {
		nextArangoMembersLayer, err := getArangoMembers(ctx, k, namespace, arangoMembers.Continue)
		if err != nil {
			return nil, err
		}

		return append(arangoMembers.Items, nextArangoMembersLayer...), nil
	}

	return arangoMembers.Items, nil
}
