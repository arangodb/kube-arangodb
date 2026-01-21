//
// DISCLAIMER
//
// Copyright 2024-2026 ArangoDB GmbH, Cologne, Germany
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

package policies

import (
	"context"
	goStrings "strings"

	admission "k8s.io/api/admission/v1"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	schedulerApi "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/patch"
	"github.com/arangodb/kube-arangodb/pkg/logging"
	"github.com/arangodb/kube-arangodb/pkg/scheduler"
	"github.com/arangodb/kube-arangodb/pkg/util"
	utilConstants "github.com/arangodb/kube-arangodb/pkg/util/constants"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
	kerrors "github.com/arangodb/kube-arangodb/pkg/util/k8sutil/kerrors"
	"github.com/arangodb/kube-arangodb/pkg/util/kclient"
	"github.com/arangodb/kube-arangodb/pkg/webhook"
)

func Handler(client kclient.Client) webhook.Handler[*core.Pod] {
	return handler{
		client: client,
	}
}

var _ webhook.MutationHandler[*core.Pod] = handler{}

type handler struct {
	client kclient.Client
}

func (h handler) CanHandle(ctx context.Context, log logging.Logger, t webhook.AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) bool {
	if request == nil {
		return false
	}

	if request.Operation != admission.Create {
		return false
	}

	if new == nil {
		return false
	}

	if _, ok := new.GetLabels()[utilConstants.ProfilesDeployment]; ok {
		return true
	}

	if _, ok := new.GetLabels()[utilConstants.ProfilesApplyLabel]; ok {
		return true
	}

	return false
}

func (h handler) Mutate(ctx context.Context, log logging.Logger, t webhook.AdmissionRequestType, request *admission.AdmissionRequest, old, new *core.Pod) (webhook.MutationResponse, error) {
	if !h.CanHandle(ctx, log, t, request, old, new) {
		return webhook.MutationResponse{}, errors.Errorf("Object cannot be handled")
	}

	labels := new.GetLabels()
	annotations := new.GetAnnotations()

	if v, ok := labels[utilConstants.ProfilesDeployment]; ok {
		if _, err := h.client.Arango().DatabaseV1().ArangoDeployments(request.Namespace).Get(ctx, v, meta.GetOptions{}); err != nil {
			if kerrors.IsNotFound(err) {
				return webhook.MutationResponse{
					ValidationResponse: webhook.NewValidationResponse(false, "ArangoDeployment %s/%s not found", request.Namespace, v),
				}, nil
			}
			return webhook.MutationResponse{
				ValidationResponse: webhook.NewValidationResponse(false, "Unable to get ArangoDeployment %s/%s: %s", request.Namespace, v, err.Error()),
			}, nil
		}
	}

	allProfiles := util.FlattenLists(goStrings.Split(labels[utilConstants.ProfilesList], ","), goStrings.Split(annotations[utilConstants.ProfilesList], ","))
	profiles := util.FilterList(util.FormatList(allProfiles, func(s string) string {
		return goStrings.TrimSpace(s)
	}), func(s string) bool {
		return s != ""
	})

	calculatedProfiles, profilesChecksum, err := scheduler.Profiles(ctx, h.client.Arango().SchedulerV1beta1().ArangoProfiles(new.GetNamespace()), labels, profiles...)
	if err != nil {
		return webhook.MutationResponse{
			ValidationResponse: webhook.NewValidationResponse(false, "Unable to get ArangoProfiles: %s", err.Error()),
		}, nil
	}

	var template core.PodTemplateSpec

	template.Labels = new.GetLabels()
	template.Annotations = new.GetAnnotations()
	new.Spec.DeepCopyInto(&template.Spec)

	if template.Annotations == nil {
		template.Annotations = map[string]string{}
	}

	template.Annotations[utilConstants.ProfilesAnnotationApplied] = "true"
	template.Annotations[utilConstants.ProfilesAnnotationChecksum] = profilesChecksum
	template.Annotations[utilConstants.ProfilesAnnotationProfiles] = goStrings.Join(util.FormatList(calculatedProfiles, func(a util.KV[string, schedulerApi.ProfileAcceptedTemplate]) string {
		return a.K
	}), ",")

	if err := schedulerApi.ProfileTemplates(util.FormatList(calculatedProfiles, func(a util.KV[string, schedulerApi.ProfileAcceptedTemplate]) *schedulerApi.ProfileTemplate {
		return a.V.Template
	})).RenderOnTemplate(&template); err != nil {
		return webhook.MutationResponse{
			ValidationResponse: webhook.NewValidationResponse(false, "Unable to get apply ArangoProfiles: %s", err.Error()),
		}, nil
	}

	return webhook.MutationResponse{
		ValidationResponse: webhook.ValidationResponse{Allowed: true},
		Patch: []patch.Item{
			patch.ItemReplace(patch.NewPath("metadata", "labels"), template.Labels),
			patch.ItemReplace(patch.NewPath("metadata", "annotations"), template.Annotations),
			patch.ItemReplace(patch.NewPath("spec"), template.Spec),
		},
	}, nil
}
