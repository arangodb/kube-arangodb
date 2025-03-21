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

	v1 "github.com/arangodb/kube-arangodb/pkg/apis/deployment/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeArangoDeployments implements ArangoDeploymentInterface
type FakeArangoDeployments struct {
	Fake *FakeDatabaseV1
	ns   string
}

var arangodeploymentsResource = v1.SchemeGroupVersion.WithResource("arangodeployments")

var arangodeploymentsKind = v1.SchemeGroupVersion.WithKind("ArangoDeployment")

// Get takes name of the arangoDeployment, and returns the corresponding arangoDeployment object, and an error if there is any.
func (c *FakeArangoDeployments) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ArangoDeployment, err error) {
	emptyResult := &v1.ArangoDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(arangodeploymentsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoDeployment), err
}

// List takes label and field selectors, and returns the list of ArangoDeployments that match those selectors.
func (c *FakeArangoDeployments) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ArangoDeploymentList, err error) {
	emptyResult := &v1.ArangoDeploymentList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(arangodeploymentsResource, arangodeploymentsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ArangoDeploymentList{ListMeta: obj.(*v1.ArangoDeploymentList).ListMeta}
	for _, item := range obj.(*v1.ArangoDeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested arangoDeployments.
func (c *FakeArangoDeployments) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(arangodeploymentsResource, c.ns, opts))

}

// Create takes the representation of a arangoDeployment and creates it.  Returns the server's representation of the arangoDeployment, and an error, if there is any.
func (c *FakeArangoDeployments) Create(ctx context.Context, arangoDeployment *v1.ArangoDeployment, opts metav1.CreateOptions) (result *v1.ArangoDeployment, err error) {
	emptyResult := &v1.ArangoDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(arangodeploymentsResource, c.ns, arangoDeployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoDeployment), err
}

// Update takes the representation of a arangoDeployment and updates it. Returns the server's representation of the arangoDeployment, and an error, if there is any.
func (c *FakeArangoDeployments) Update(ctx context.Context, arangoDeployment *v1.ArangoDeployment, opts metav1.UpdateOptions) (result *v1.ArangoDeployment, err error) {
	emptyResult := &v1.ArangoDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(arangodeploymentsResource, c.ns, arangoDeployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoDeployment), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeArangoDeployments) UpdateStatus(ctx context.Context, arangoDeployment *v1.ArangoDeployment, opts metav1.UpdateOptions) (result *v1.ArangoDeployment, err error) {
	emptyResult := &v1.ArangoDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(arangodeploymentsResource, "status", c.ns, arangoDeployment, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoDeployment), err
}

// Delete takes name of the arangoDeployment and deletes it. Returns an error if one occurs.
func (c *FakeArangoDeployments) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(arangodeploymentsResource, c.ns, name, opts), &v1.ArangoDeployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeArangoDeployments) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(arangodeploymentsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ArangoDeploymentList{})
	return err
}

// Patch applies the patch and returns the patched arangoDeployment.
func (c *FakeArangoDeployments) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ArangoDeployment, err error) {
	emptyResult := &v1.ArangoDeployment{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(arangodeploymentsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoDeployment), err
}
