//
// DISCLAIMER
//
// Copyright 2016-2023 ArangoDB GmbH, Cologne, Germany
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

package features

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

const prefixArg = "deployment.feature"

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

// Iterator defines feature definition iterator
type Iterator func(name string, feature Feature)

// Iterate allows to iterate over all registered functions
func Iterate(iterator Iterator) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	for name, feature := range features {
		iterator(name, feature)
	}
}

// Init initializes all registered features.
// If a feature is not provided via process's argument, then it is taken from environment variable
// or from enabled by default setting.
func Init(cmd *cobra.Command) error {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	cmd.AddCommand(internalCMD)

	f := cmd.Flags()

	featureArgName := GetFeatureArgName("all")
	f.BoolVar(&enableAll, featureArgName, isEnabledFeatureFromEnv(featureArgName), "Enable ALL Features")

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

		featureArgName = GetFeatureArgName(feature.Name())
		enabled := feature.EnabledByDefault() || isEnabledFeatureFromEnv(featureArgName)
		f.BoolVar(feature.EnabledPointer(), featureArgName, enabled, z)

		if ok, reason := feature.Deprecated(); ok {
			if err := f.MarkDeprecated(featureArgName, reason); err != nil {
				return err
			}
		}

		if feature.Hidden() {
			if err := f.MarkHidden(featureArgName); err != nil {
				return err
			}
		}
	}

	f.StringVar(&configMapName, "features-config-map-name", DefaultFeaturesConfigMap, "Name of the Feature Map ConfigMap")

	return nil
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

// Supported returns false when:
// - feature is disabled.
// - a given version is lower than minimum feature version.
// - feature expects enterprise but a given enterprise arg is not true.
func Supported(f Feature, v driver.Version, enterprise bool) bool {
	if !f.Enabled() {
		return false
	}

	if f.EnterpriseRequired() && !enterprise {
		// This feature requires enterprise version but current version is not enterprise.
		return false
	}

	return v.CompareTo(f.Version()) >= 0
}

// GetFeatureMap returns all features' arguments names.
func GetFeatureMap() map[string]bool {
	args := make(map[string]bool, len(features))
	for _, f := range features {
		args[util.NormalizeEnv(GetFeatureArgName(f.Name()))] = f.Enabled()
	}

	return args
}

// GetFeatureArgName returns feature process argument name.
func GetFeatureArgName(featureName string) string {
	return fmt.Sprintf("%s.%s", prefixArg, featureName)
}

// isEnabledFeatureFromEnv returns true if argument is enabled as an environment variable.
func isEnabledFeatureFromEnv(arg string) bool {
	return os.Getenv(util.NormalizeEnv(arg)) == Enabled
}
