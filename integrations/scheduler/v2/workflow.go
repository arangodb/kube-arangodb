//
// DISCLAIMER
//
// Copyright 2026 ArangoDB GmbH, Cologne, Germany
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

package v2

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	pbSchedulerV2 "github.com/arangodb/kube-arangodb/integrations/scheduler/v2/definition"
	platformApi "github.com/arangodb/kube-arangodb/pkg/apis/platform/v1beta1"
	sharedApi "github.com/arangodb/kube-arangodb/pkg/apis/shared/v1"
	"github.com/arangodb/kube-arangodb/pkg/deployment/features"
	platformv1beta1 "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/typed/platform/v1beta1"
	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/k8sutil/helm"
)

// workflowEnabled reports whether the SchedulerV2 Integration should manage
// ArangoPlatformWorkflow resources instead of calling Helm directly.
func (i *implementation) workflowEnabled() bool {
	return features.SchedulerV2Workflow().Enabled()
}

func (i *implementation) workflows() platformv1beta1.ArangoPlatformWorkflowInterface {
	return i.kclient.Arango().PlatformV1beta1().ArangoPlatformWorkflows(i.cfg.Namespace)
}

// installV2Workflow ensures the desired ArangoPlatformWorkflow and returns immediately.
// The Helm release is reconciled asynchronously by the operator; callers observe
// convergence via Status.
func (i *implementation) installV2Workflow(ctx context.Context, in *pbSchedulerV2.SchedulerV2InstallV2Request) (*pbSchedulerV2.SchedulerV2InstallV2Response, error) {
	spec, err := i.workflowSpec(in.GetChart(), in.GetValues(), nil)
	if err != nil {
		logger.Err(err).Warn("Unable to render values: InstallV2")
		return nil, status.Errorf(codes.Internal, "Unable to render values: InstallV2: %s", err.Error())
	}

	wf, err := i.ensureWorkflow(ctx, in.GetName(), spec, in.GetOptions().GetLabels())
	if err != nil {
		logger.Err(err).Warn("Unable to apply workflow: InstallV2")
		return nil, status.Errorf(codes.Internal, "Unable to apply workflow: InstallV2: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2InstallV2Response{
		Release: newChartReleaseFromWorkflow(wf),
	}, nil
}

// upgradeV2Workflow ensures the desired ArangoPlatformWorkflow and returns immediately.
func (i *implementation) upgradeV2Workflow(ctx context.Context, in *pbSchedulerV2.SchedulerV2UpgradeV2Request) (*pbSchedulerV2.SchedulerV2UpgradeV2Response, error) {
	var upgrade *platformApi.ArangoPlatformWorkflowSpecUpgrade

	maxHistory := i.cfg.MaxHistory
	if in.GetOptions() != nil && in.GetOptions().MaxHistory != nil {
		maxHistory = int(in.GetOptions().GetMaxHistory())
	}
	upgrade = &platformApi.ArangoPlatformWorkflowSpecUpgrade{
		MaxHistory: util.NewType(maxHistory),
	}

	spec, err := i.workflowSpec(in.GetChart(), in.GetValues(), upgrade)
	if err != nil {
		logger.Err(err).Warn("Unable to render values: UpgradeV2")
		return nil, status.Errorf(codes.Internal, "Unable to render values: UpgradeV2: %s", err.Error())
	}

	before, err := i.workflows().Get(ctx, in.GetName(), meta.GetOptions{})
	if err != nil && !apiErrors.IsNotFound(err) {
		logger.Err(err).Warn("Unable to read workflow: UpgradeV2")
		return nil, status.Errorf(codes.Internal, "Unable to read workflow: UpgradeV2: %s", err.Error())
	}

	wf, err := i.ensureWorkflow(ctx, in.GetName(), spec, in.GetOptions().GetLabels())
	if err != nil {
		logger.Err(err).Warn("Unable to apply workflow: UpgradeV2")
		return nil, status.Errorf(codes.Internal, "Unable to apply workflow: UpgradeV2: %s", err.Error())
	}

	var r pbSchedulerV2.SchedulerV2UpgradeV2Response

	if before != nil {
		r.Before = newChartReleaseFromWorkflow(before)
	}
	r.After = newChartReleaseFromWorkflow(wf)

	return &r, nil
}

// uninstallWorkflow removes the ArangoPlatformWorkflow. The operator finalizer
// uninstalls the underlying Helm release.
func (i *implementation) uninstallWorkflow(ctx context.Context, name string) (*pbSchedulerV2.SchedulerV2UninstallResponse, error) {
	if err := i.workflows().Delete(ctx, name, meta.DeleteOptions{}); err != nil {
		if apiErrors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, "Release `%s` not found", name)
		}

		logger.Err(err).Warn("Unable to apply workflow: Uninstall")
		return nil, status.Errorf(codes.Internal, "Unable to apply workflow: Uninstall: %s", err.Error())
	}

	return &pbSchedulerV2.SchedulerV2UninstallResponse{
		Info: "Release deletion scheduled",
	}, nil
}

// ensureWorkflow creates or updates the ArangoPlatformWorkflow with the desired spec.
func (i *implementation) ensureWorkflow(ctx context.Context, name string, spec platformApi.ArangoPlatformWorkflowSpec, extraLabels map[string]string) (*platformApi.ArangoPlatformWorkflow, error) {
	if name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Name cannot be empty")
	}

	labels := map[string]string{
		LabelArangoDBDeploymentName: i.cfg.Deployment,
	}
	for k, v := range extraLabels {
		labels[k] = v
	}

	client := i.workflows()

	existing, err := client.Get(ctx, name, meta.GetOptions{})
	if err != nil {
		if !apiErrors.IsNotFound(err) {
			return nil, err
		}

		return client.Create(ctx, &platformApi.ArangoPlatformWorkflow{
			ObjectMeta: meta.ObjectMeta{
				Name:      name,
				Namespace: i.cfg.Namespace,
				Labels:    labels,
			},
			Spec: spec,
		}, meta.CreateOptions{})
	}

	existing.Spec = spec
	if existing.Labels == nil {
		existing.Labels = map[string]string{}
	}
	for k, v := range labels {
		existing.Labels[k] = v
	}

	return client.Update(ctx, existing, meta.UpdateOptions{})
}

// workflowSpec renders the workflow spec from the request. Chart Overrides are
// merged by the operator handler, so only the request values are stored here.
func (i *implementation) workflowSpec(chart string, rawValues [][]byte, upgrade *platformApi.ArangoPlatformWorkflowSpecUpgrade) (platformApi.ArangoPlatformWorkflowSpec, error) {
	spec := platformApi.ArangoPlatformWorkflowSpec{
		Deployment: &sharedApi.Object{Name: i.cfg.Deployment},
		Chart:      &sharedApi.Object{Name: chart},
		Upgrade:    upgrade,
	}

	values, err := mergeWorkflowValues(rawValues)
	if err != nil {
		return platformApi.ArangoPlatformWorkflowSpec{}, err
	}
	if len(values) > 0 {
		spec.Values = values
	}

	return spec, nil
}

func mergeWorkflowValues(in [][]byte) (sharedApi.Any, error) {
	rawValues := make([]helm.Values, 0, len(in))
	for _, v := range in {
		if len(v) > 0 {
			rawValues = append(rawValues, v)
		}
	}

	if len(rawValues) == 0 {
		return nil, nil
	}

	values, err := helm.NewMergeRawValues(helm.MergeMaps, rawValues...)
	if err != nil {
		return nil, err
	}

	return sharedApi.Any(values), nil
}

func newChartReleaseFromWorkflow(in *platformApi.ArangoPlatformWorkflow) *pbSchedulerV2.SchedulerV2Release {
	if in == nil {
		return nil
	}

	var rel pbSchedulerV2.SchedulerV2Release

	rel.Name = in.GetName()
	rel.Namespace = in.GetNamespace()
	rel.Labels = in.GetLabels()

	if len(in.Status.Values) > 0 {
		rel.Values = in.Status.Values
	}

	if r := in.Status.Release; r != nil {
		rel.Version = int64(r.Version)
		rel.Info = newChartReleaseInfoFromWorkflow(r)
	}

	return &rel
}

func newChartReleaseInfoFromWorkflow(in *platformApi.ArangoPlatformWorkflowStatusRelease) *pbSchedulerV2.SchedulerV2ReleaseInfo {
	if in == nil {
		return nil
	}

	var info pbSchedulerV2.SchedulerV2ReleaseInfo

	info.Description = in.Info.Description
	info.Status = pbSchedulerV2.FromHelmStatus(in.Info.Status)

	if t := in.Info.FirstDeployed; t != nil {
		info.FirstDeployed = timestamppb.New(t.Time)
	}
	if t := in.Info.LastDeployed; t != nil {
		info.LastDeployed = timestamppb.New(t.Time)
	}

	return &info
}
