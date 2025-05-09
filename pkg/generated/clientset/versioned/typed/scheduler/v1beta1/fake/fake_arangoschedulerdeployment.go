//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1beta1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeArangoSchedulerDeployments implements ArangoSchedulerDeploymentInterface
type FakeArangoSchedulerDeployments struct {
	Fake *FakeSchedulerV1beta1
	ns   string
}

var arangoschedulerdeploymentsResource = v1beta1.SchemeGroupVersion.WithResource("arangoschedulerdeployments")

var arangoschedulerdeploymentsKind = v1beta1.SchemeGroupVersion.WithKind("ArangoSchedulerDeployment")

// Get takes name of the arangoSchedulerDeployment, and returns the corresponding arangoSchedulerDeployment object, and an error if there is any.
func (c *FakeArangoSchedulerDeployments) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1beta1.ArangoSchedulerDeployment, err error) {
	emptyResult := &v1beta1.ArangoSchedulerDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(arangoschedulerdeploymentsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ArangoSchedulerDeployment), err
}

// List takes label and field selectors, and returns the list of ArangoSchedulerDeployments that match those selectors.
func (c *FakeArangoSchedulerDeployments) List(ctx context.Context, opts v1.ListOptions) (result *v1beta1.ArangoSchedulerDeploymentList, err error) {
	emptyResult := &v1beta1.ArangoSchedulerDeploymentList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(arangoschedulerdeploymentsResource, arangoschedulerdeploymentsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.ArangoSchedulerDeploymentList{ListMeta: obj.(*v1beta1.ArangoSchedulerDeploymentList).ListMeta}
	for _, item := range obj.(*v1beta1.ArangoSchedulerDeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested arangoSchedulerDeployments.
func (c *FakeArangoSchedulerDeployments) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(arangoschedulerdeploymentsResource, c.ns, opts))

}

// Create takes the representation of a arangoSchedulerDeployment and creates it.  Returns the server's representation of the arangoSchedulerDeployment, and an error, if there is any.
func (c *FakeArangoSchedulerDeployments) Create(ctx context.Context, arangoSchedulerDeployment *v1beta1.ArangoSchedulerDeployment, opts v1.CreateOptions) (result *v1beta1.ArangoSchedulerDeployment, err error) {
	emptyResult := &v1beta1.ArangoSchedulerDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(arangoschedulerdeploymentsResource, c.ns, arangoSchedulerDeployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ArangoSchedulerDeployment), err
}

// Update takes the representation of a arangoSchedulerDeployment and updates it. Returns the server's representation of the arangoSchedulerDeployment, and an error, if there is any.
func (c *FakeArangoSchedulerDeployments) Update(ctx context.Context, arangoSchedulerDeployment *v1beta1.ArangoSchedulerDeployment, opts v1.UpdateOptions) (result *v1beta1.ArangoSchedulerDeployment, err error) {
	emptyResult := &v1beta1.ArangoSchedulerDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(arangoschedulerdeploymentsResource, c.ns, arangoSchedulerDeployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ArangoSchedulerDeployment), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeArangoSchedulerDeployments) UpdateStatus(ctx context.Context, arangoSchedulerDeployment *v1beta1.ArangoSchedulerDeployment, opts v1.UpdateOptions) (result *v1beta1.ArangoSchedulerDeployment, err error) {
	emptyResult := &v1beta1.ArangoSchedulerDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(arangoschedulerdeploymentsResource, "status", c.ns, arangoSchedulerDeployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ArangoSchedulerDeployment), err
}

// Delete takes name of the arangoSchedulerDeployment and deletes it. Returns an error if one occurs.
func (c *FakeArangoSchedulerDeployments) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(arangoschedulerdeploymentsResource, c.ns, name, opts), &v1beta1.ArangoSchedulerDeployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeArangoSchedulerDeployments) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(arangoschedulerdeploymentsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1beta1.ArangoSchedulerDeploymentList{})
	return err
}

// Patch applies the patch and returns the patched arangoSchedulerDeployment.
func (c *FakeArangoSchedulerDeployments) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1beta1.ArangoSchedulerDeployment, err error) {
	emptyResult := &v1beta1.ArangoSchedulerDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(arangoschedulerdeploymentsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1beta1.ArangoSchedulerDeployment), err
}

// GetScale takes name of the arangoSchedulerDeployment, and returns the corresponding scale object, and an error if there is any.
func (c *FakeArangoSchedulerDeployments) GetScale(ctx context.Context, arangoSchedulerDeploymentName string, options v1.GetOptions) (result *autoscalingv1.Scale, err error) {
	emptyResult := &autoscalingv1.Scale{}
	obj, err := c.Fake.
		Invokes(testing.NewGetSubresourceActionWithOptions(arangoschedulerdeploymentsResource, c.ns, "scale", arangoSchedulerDeploymentName, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*autoscalingv1.Scale), err
}

// UpdateScale takes the representation of a scale and updates it. Returns the server's representation of the scale, and an error, if there is any.
func (c *FakeArangoSchedulerDeployments) UpdateScale(ctx context.Context, arangoSchedulerDeploymentName string, scale *autoscalingv1.Scale, opts v1.UpdateOptions) (result *autoscalingv1.Scale, err error) {
	emptyResult := &autoscalingv1.Scale{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(arangoschedulerdeploymentsResource, "scale", c.ns, scale, opts), &autoscalingv1.Scale{})

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*autoscalingv1.Scale), err
}
