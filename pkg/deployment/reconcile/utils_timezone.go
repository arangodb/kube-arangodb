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

package reconcile

import (
	"github.com/arangodb/kube-arangodb/pkg/deployment/pod"
	"github.com/arangodb/kube-arangodb/pkg/generated/timezones"
	"github.com/arangodb/kube-arangodb/pkg/util"
	secretv1 "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/inspector/secret/v1"
)

const defaultTimezone = "UTC"

func GetTimezone(tz *string) (timezones.Timezone, bool) {
	if tz == nil {
		return timezones.GetTimezone(defaultTimezone)
	}
	return timezones.GetTimezone(*tz)
}

func IsTimezoneValid(cache secretv1.Inspector, name string, timezone timezones.Timezone) bool {
	sn := pod.TimezoneSecret(name)

	tzd, ok := timezone.GetData()
	if !ok {
		// Unable to get TZ Data, so ignoring
		return true
	}

	if s, ok := cache.GetSimple(sn); ok {
		// Secret exists, verify
		if v, ok := s.Data[pod.TimezoneNameKey]; ok {
			if string(v) != timezone.Name {
				return false
			}
		} else {
			return false
		}
		if v, ok := s.Data[pod.TimezoneDataKey]; ok {
			if util.SHA256(v) != util.SHA256(tzd) {
				return false
			}
		} else {
			return false
		}
		if v, ok := s.Data[pod.TimezoneTZKey]; ok {
			if string(v) != timezone.Zone {
				return false
			}
		} else {
			return false
		}
	} else {
		return false
	}

	return true
}
