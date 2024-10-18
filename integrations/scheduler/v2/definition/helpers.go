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

package definition

import (
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

func (i SchedulerV2ReleaseInfoStatus) AsHelmStatus() release.Status {
	switch i {
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_UNKNOWN_UNSPECIFIED:
		return release.StatusUnknown
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_DEPLOYED:
		return release.StatusDeployed
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_UNINSTALLED:
		return release.StatusUninstalled
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_SUPERSEDED:
		return release.StatusSuperseded
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_FAILED:
		return release.StatusFailed
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_UNINSTALLING:
		return release.StatusUninstalling
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_PENDINGINSTALL:
		return release.StatusPendingInstall
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_PENDINGUPGRADE:
		return release.StatusPendingUpgrade
	case SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_PENDINGROLLBACK:
		return release.StatusPendingRollback
	default:
		return release.StatusUnknown
	}
}

func FromHelmStatus(in release.Status) SchedulerV2ReleaseInfoStatus {
	switch in {
	case release.StatusUnknown:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_UNKNOWN_UNSPECIFIED
	case release.StatusDeployed:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_DEPLOYED
	case release.StatusUninstalled:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_UNINSTALLED
	case release.StatusSuperseded:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_SUPERSEDED
	case release.StatusFailed:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_FAILED
	case release.StatusUninstalling:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_UNINSTALLING
	case release.StatusPendingInstall:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_PENDINGINSTALL
	case release.StatusPendingUpgrade:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_PENDINGUPGRADE
	case release.StatusPendingRollback:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_PENDINGROLLBACK
	default:
		return SchedulerV2ReleaseInfoStatus_SCHEDULER_V2_RELEASE_INFO_STATUS_UNKNOWN_UNSPECIFIED
	}
}

func (i *SchedulerV2InstallRequestOptions) Options() []util.Mod[action.Install] {
	if i == nil {
		return nil
	}

	var opts []util.Mod[action.Install]

	if v := i.GetLabels(); len(v) > 0 {
		opts = append(opts, func(in *action.Install) {
			in.Labels = v
		})
	}

	return opts
}

func (i *SchedulerV2UpgradeRequestOptions) Options() []util.Mod[action.Upgrade] {
	if i == nil {
		return nil
	}

	var opts []util.Mod[action.Upgrade]

	if v := i.GetLabels(); len(v) > 0 {
		opts = append(opts, func(in *action.Upgrade) {
			in.Labels = v
		})
	}

	return opts
}

func (i *SchedulerV2ListRequestOptions) Options() []util.Mod[action.List] {
	if i == nil {
		return nil
	}

	var opts []util.Mod[action.List]

	if v := i.GetSelectors(); len(v) > 0 {
		opts = append(opts, func(in *action.List) {
			s := labels.NewSelector()

			for k, v := range v {
				if r, err := labels.NewRequirement(k, selection.DoubleEquals, []string{v}); err == nil && r != nil {
					s = s.Add(*r)
				}
			}

			in.Selector = s.String()
		})
	}

	return opts
}

func (i *SchedulerV2UninstallRequestOptions) Options() []util.Mod[action.Uninstall] {
	if i == nil {
		return nil
	}

	var opts []util.Mod[action.Uninstall]

	return opts
}

func (i *SchedulerV2TestRequestOptions) Options() []util.Mod[action.ReleaseTesting] {
	if i == nil {
		return nil
	}

	var opts []util.Mod[action.ReleaseTesting]

	return opts
}

func (i *SchedulerV2GVK) AsHelmResource() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   i.GetGroup(),
		Version: i.GetVersion(),
		Kind:    i.GetKind(),
	}
}

func (i *SchedulerV2ReleaseInfoResource) AsHelmResource() helm.Resource {
	if i == nil {
		return helm.Resource{}
	}

	return helm.Resource{
		GroupVersionKind: i.GetGvk().AsHelmResource(),
		Name:             i.GetName(),
		Namespace:        i.GetNamespace(),
	}
}
