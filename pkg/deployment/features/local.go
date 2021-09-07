//
// DISCLAIMER
//
// Copyright 2020-2021 ArangoDB GmbH, Cologne, Germany
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

package features

import (
	"fmt"
	"sync"

	"github.com/arangodb/go-driver"
	"github.com/spf13/cobra"
)

var features = map[string]Feature{}
var featuresLock sync.Mutex
var enableAll = false

func registerFeature(f Feature) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	if f == nil {
		panic("Feature cannot be nil")
	}

	if _, ok := features[f.Name()]; ok {
		panic("Feature already registered")
	}

	features[f.Name()] = f
}

var internalCMD = &cobra.Command{
	Use:   "features",
	Short: "Describe all operator features",
	Run:   cmdRun,
}

func Init(cmd *cobra.Command) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	cmd.AddCommand(internalCMD)

	f := cmd.Flags()

	f.BoolVar(&enableAll, "deployment.feature.all", false, "Enable ALL Features")

	for _, feature := range features {
		z := ""

		if v := feature.Version(); v != "" || feature.EnterpriseRequired() {
			if v != "" && feature.EnterpriseRequired() {
				z = fmt.Sprintf("%s - Required version %s and Enterprise Edition", feature.Description(), v)
			} else if v != "" {
				z = fmt.Sprintf("%s - Required version %s", feature.Description(), v)
			} else if feature.EnterpriseRequired() {
				z = fmt.Sprintf("%s - Required Enterprise Edition", feature.Description())
			} else {
				z = feature.Description()
			}
		}

		featureName := fmt.Sprintf("deployment.feature.%s", feature.Name())
		if ok, reason := feature.Deprecated(); ok {
			f.MarkDeprecated(featureName, reason)
		}

		f.BoolVar(feature.EnabledPointer(), featureName, feature.EnabledByDefault(), z)
	}
}

func cmdRun(_ *cobra.Command, _ []string) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	for _, feature := range features {
		println(fmt.Sprintf("Feature: %s", feature.Name()))
		println(fmt.Sprintf("Description: %s", feature.Description()))
		if feature.EnabledByDefault() {
			println("Enabled: true")
		} else {
			println("Enabled: false")
		}
		if v := feature.Version(); v != "" {
			println(fmt.Sprintf("ArangoDB Version Required: >= %s", v))
		}

		if feature.EnterpriseRequired() {
			println("ArangoDB Edition Required: Enterprise")
		} else {
			println("ArangoDB Edition Required: Community, Enterprise")
		}

		if ok, reason := feature.Deprecated(); ok {
			println(fmt.Sprintf("Deprecated: %s", reason))
		}

		println()
	}
}

func Supported(f Feature, v driver.Version, enterprise bool) bool {
	return f.Enabled() && ((f.EnterpriseRequired() && enterprise) || !f.EnterpriseRequired()) && v.CompareTo(f.Version()) >= 0
}
