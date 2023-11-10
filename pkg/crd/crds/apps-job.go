//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package crds

import (
	_ "embed"
	"github.com/arangodb/kube-arangodb/pkg/util"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/arangodb/go-driver"
)

const (
	AppsJobVersion = driver.Version("1.0.1")
)

func init() {
	if err := yaml.Unmarshal(appsJobs, &appsJobsCRD); err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(appsJobsSchema, &appsJobsCRDSchema); err != nil {
		panic(err)
	}
}

func DefaultAppsJobOptions() CRDOptions {
	return CRDOptions{
		WithSchema: util.NewType(false),
	}
}

func AppsJob() *apiextensions.CustomResourceDefinition {
	return AppsJobWithOptions(nil)
}

func AppsJobWithOptions(options *CRDOptions) *apiextensions.CustomResourceDefinition {
	return appsJobWithOptions(options)
}

func appsJobWithOptions(options *CRDOptions) *apiextensions.CustomResourceDefinition {
	return extendCRDWithSchema(appsJobsCRD.DeepCopy(), options.Merge(DefaultAppsJobOptions()), appsJobsCRDSchema)
}

func AppsJobDefinitionWithOptions(options *CRDOptions) Definition {
	return Definition{
		Version: AppsJobVersion,
		CRD:     AppsJobWithOptions(options),
	}
}

func AppsJobDefinition() Definition {
	return AppsJobDefinitionWithOptions(nil)
}

var appsJobsCRDSchema CRDSchemas

//go:embed apps-job.schema.yaml
var appsJobsSchema []byte

var appsJobsCRD apiextensions.CustomResourceDefinition

//go:embed apps-job.yaml
var appsJobs []byte
