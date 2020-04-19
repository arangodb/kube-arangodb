//
// DISCLAIMER
//
// Copyright 2020 ArangoDB GmbH, Cologne, Germany
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
//

package webhook

import (
	"encoding/json"
	"fmt"

	"github.com/arangodb/kube-arangodb/pkg/apis/deployment"
	api "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	"github.com/arangodb/kube-arangodb/pkg/util/log"
	admission "k8s.io/api/admission/v1beta1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	if err := RegisterHandler(arangoDeployment{}); err != nil {
		panic(err)
	}
}

var _ ValidationCreateHandler = arangoDeployment{}

type arangoDeployment struct {
}

func (p arangoDeployment) asObject(obj runtime.RawExtension) (*api.ArangoDeployment, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var o api.ArangoDeployment

	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}

	return &o, nil
}

func (p arangoDeployment) ValidateCreate(log log.Factory, request *admission.AdmissionRequest) (bool, string) {
	_, err := p.asObject(request.Object) // Only check if object can be parsed
	if err != nil {
		return false, fmt.Sprintf("Unable to parse object as ObjectMeta: %s", err.Error())
	}

	return true, ""
}

func (p arangoDeployment) CanBeHandled(gvk meta.GroupVersionKind) bool {
	// Handle ArangoDeployment
	return gvk.Group == deployment.ArangoDeploymentGroupName &&
		gvk.Kind == deployment.ArangoDeploymentResourceKind &&
		(gvk.Version == api.ArangoDeploymentVersion || gvk.Version == "v1alpha")
}

func (p arangoDeployment) Name() string {
	return "arango-deployment-handler"
}
