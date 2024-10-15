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

package helm

import (
	"helm.sh/helm/v3/pkg/release"
	"time"
)

func fromHelmRelease(in *release.Release) Release {
	var r Release

	if in != nil {
		r.Name = in.Name
		r.Version = in.Version
		r.Namespace = in.Namespace
		r.Labels = in.Labels

		r.Info = fromHelmReleaseInfo(in.Info)
	}

	return r
}

func fromHelmReleaseInfo(in *release.Info) ReleaseInfo {
	var r ReleaseInfo

	if in != nil {
		r.FirstDeployed = in.FirstDeployed.Time
		r.LastDeployed = in.LastDeployed.Time
		r.Deleted = in.Deleted.Time
		r.Description = in.Description
		r.Status = in.Status
		r.Notes = in.Notes
	}

	return r
}

type Release struct {
	Name string

	Info ReleaseInfo

	Version   int
	Namespace string
	Labels    map[string]string
}

type ReleaseInfo struct {
	FirstDeployed time.Time
	LastDeployed  time.Time
	Deleted       time.Time
	Description   string
	Status        release.Status
	Notes         string
}

type UninstallRelease struct {
	Release Release
	Info    string
}
