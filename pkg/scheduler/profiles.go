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

package scheduler

import (
	"context"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/generic"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/list"
	"github.com/arangodb/kube-arangodb/pkg/util/strings"
)

func Profiles(ctx context.Context, client generic.ListInterface[*schedulerApi.ArangoProfileList], labels map[string]string, profiles ...string) ([]util.KV[string, schedulerApi.ProfileAcceptedTemplate], string, error) {
	profileList, err := list.APIList[*schedulerApi.ArangoProfileList, *schedulerApi.ArangoProfile](ctx, client, meta.ListOptions{}, func(result *schedulerApi.ArangoProfileList) []*schedulerApi.ArangoProfile {
		q := make([]*schedulerApi.ArangoProfile, len(result.Items))

		for id, e := range result.Items {
			q[id] = e.DeepCopy()
		}

		return q
	})
	if err != nil {
		return nil, "", err
	}

	profileMap := util.ListAsMap(profileList, func(in *schedulerApi.ArangoProfile) string {
		return in.GetName()
	})

	extractedProfiles := util.List[*schedulerApi.ArangoProfile](profileList).Filter(func(a *schedulerApi.ArangoProfile) bool {
		return a != nil && a.Spec.Template != nil
	}).Filter(func(a *schedulerApi.ArangoProfile) bool {
		if a.Spec.Selectors == nil {
			return false
		}

		if !a.Spec.Selectors.Select(labels) {
			return false
		}

		return true
	})

	for _, name := range profiles {
		p, ok := profileMap[name]
		if !ok {
			return nil, "", errors.Errorf("Profile with name `%s` is missing", name)
		}

		extractedProfiles = append(extractedProfiles, p)
	}

	extractedProfiles = extractedProfiles.Unique(func(existing util.List[*schedulerApi.ArangoProfile], o *schedulerApi.ArangoProfile) bool {
		return existing.Contains(func(a *schedulerApi.ArangoProfile) bool {
			return a.GetName() == o.GetName()
		})
	})

	extractedProfiles = extractedProfiles.Sort(func(a, b *schedulerApi.ArangoProfile) bool {
		if ca, cb := a.Spec.Template.GetPriority(), b.Spec.Template.GetPriority(); ca != cb {
			return ca > cb
		}

		if ca, cb := a.GetName(), b.GetName(); ca != cb {
			return ca > cb
		}

		return a.GetCreationTimestamp().After(b.GetCreationTimestamp().Time)
	})

	// Check if everything is valid
	if err := errors.Errors(util.FormatList(extractedProfiles, func(in *schedulerApi.ArangoProfile) error {
		if !in.Status.Conditions.IsTrue(schedulerApi.ReadyCondition) {
			return errors.Errorf("ArangoProfile `%s` is not ready", in.GetName())
		}
		if in.Status.Accepted == nil {
			return errors.Errorf("ArangoProfile `%s` status is nil", in.GetName())
		}

		return nil
	})...); err != nil {
		return nil, "", err
	}

	resultProfiles := util.FormatList(extractedProfiles, func(a *schedulerApi.ArangoProfile) util.KV[string, schedulerApi.ProfileAcceptedTemplate] {
		return util.KV[string, schedulerApi.ProfileAcceptedTemplate]{
			K: a.GetName(),
			V: *a.Status.Accepted,
		}
	})

	return resultProfiles, util.SHA256FromString(strings.Join(util.FormatList(resultProfiles, func(a util.KV[string, schedulerApi.ProfileAcceptedTemplate]) string {
		return a.V.Checksum
	}), "|")), nil
}
