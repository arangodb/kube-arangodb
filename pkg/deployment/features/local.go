//
// DISCLAIMER
//
// Copyright 2016-2024 ArangoDB GmbH, Cologne, Germany
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
	"strings"
	"sync"

	"github.com/spf13/cobra"

	"github.com/arangodb/go-driver"

	"github.com/arangodb/kube-arangodb/pkg/util"
)

const prefixArg = "deployment.feature"

const MinSupportedArangoDBVersion = "3.8.0"

var features Features
var featuresLock sync.Mutex
var enableAll = false

func registerFeature(f Feature) {
	featuresLock.Lock()
	defer featuresLock.Unlock()

	if _, ok := features.Get(f.Name()); ok {
		panic("Feature already registered")
	}

	features = append(features, f)

	features.Sort()
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

	for _, feature := range features {
		iterator(feature.Name(), feature)
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

		minVersion, maxVersion := feature.Version()

		if minVersion == "" {
			minVersion = MinSupportedArangoDBVersion
		}

		versionDesc := fmt.Sprintf(">= %s", minVersion)
		if maxVersion != NoVersionLimit {
			versionDesc = fmt.Sprintf(">= %s, < %s", minVersion, maxVersion)
		}

		if feature.EnterpriseRequired() {
			z = fmt.Sprintf("%s - Required ArangoDB EE %s", feature.Description(), versionDesc)
		} else {
			z = fmt.Sprintf("%s - Required ArangoDB %s", feature.Description(), versionDesc)
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
		if deps := feature.Dependencies(); len(deps) > 0 {
			names := make([]string, len(deps))
			for id := range names {
				names[id] = deps[id].Name()
			}
			println(fmt.Sprintf("Dependencies: %s", strings.Join(names, ", ")))
		}
		if feature.EnabledByDefault() {
			println("Enabled: true")
		} else {
			println("Enabled: false")
		}
		if min, max := feature.Version(); min != NoVersionLimit && max != NoVersionLimit {
			println(fmt.Sprintf("ArangoDB Version Required: >= %s, <= %s", min, max))
		} else if min != NoVersionLimit {
			println(fmt.Sprintf("ArangoDB Version Required: >= %s", min))
		} else if max != NoVersionLimit {
			println(fmt.Sprintf("ArangoDB Version Required: <= %s", max))
		}

		if feature.EnterpriseRequired() {
			println("ArangoDB Edition Required: Enterprise")
		} else {
			println("ArangoDB Edition Required: Community, Enterprise")
		}

		if deps := feature.GetDependencies(); len(deps) > 0 {
			println(fmt.Sprintf("Dependencies: %v", deps))
		}

		if ok, reason := feature.Deprecated(); ok {
			println(fmt.Sprintf("Deprecated: %s", reason))
		}

		println()
	}
}

// Supported returns false when:
// - feature is disabled.
// - any feature dependency is disabled.
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

	for _, dependency := range f.Dependencies() {
		if !Supported(dependency, v, enterprise) {
			return false
		}
	}

	min, max := f.Version()
	if min != NoVersionLimit {
		if v.CompareTo(min) < 0 {
			return false
		}
	}

	if max != NoVersionLimit {
		if v.CompareTo(max) >= 0 {
			return false
		}
	}

	return true
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

// EnableDependencies enables dependencies for features if it is required.
func EnableDependencies() {
	for {
		var changed bool

		for _, f := range features {
			if !f.Enabled() {
				continue
			}

			for _, depName := range f.GetDependencies() {
				// Don't use `dependency.Enabled` here because `constValue` is involved here, and it can not be changed.
				if enableDependencyByName(depName) {
					// Dependency is changed so list of features must be iterated once again, because this
					// dependency can turn on other dependencies.
					changed = true
				}
			}
		}

		if !changed {
			return
		}
	}
}

// enableDependencyByName enables dependency by name of a feature.
func enableDependencyByName(name string) bool {
	for _, f := range features {
		if name != f.Name() {
			continue
		}

		if ep := f.EnabledPointer(); !*ep {
			*ep = true
			return true
		}
	}

	return false
}
