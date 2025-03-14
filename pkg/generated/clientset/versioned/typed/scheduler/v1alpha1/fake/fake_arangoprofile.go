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

	v1alpha1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeArangoProfiles implements ArangoProfileInterface
type FakeArangoProfiles struct {
	Fake *FakeSchedulerV1alpha1
	ns   string
}

var arangoprofilesResource = v1alpha1.SchemeGroupVersion.WithResource("arangoprofiles")

var arangoprofilesKind = v1alpha1.SchemeGroupVersion.WithKind("ArangoProfile")

// Get takes name of the arangoProfile, and returns the corresponding arangoProfile object, and an error if there is any.
func (c *FakeArangoProfiles) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ArangoProfile, err error) {
	emptyResult := &v1alpha1.ArangoProfile{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(arangoprofilesResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ArangoProfile), err
}

// List takes label and field selectors, and returns the list of ArangoProfiles that match those selectors.
func (c *FakeArangoProfiles) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ArangoProfileList, err error) {
	emptyResult := &v1alpha1.ArangoProfileList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(arangoprofilesResource, arangoprofilesKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.ArangoProfileList{ListMeta: obj.(*v1alpha1.ArangoProfileList).ListMeta}
	for _, item := range obj.(*v1alpha1.ArangoProfileList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested arangoProfiles.
func (c *FakeArangoProfiles) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(arangoprofilesResource, c.ns, opts))

}

// Create takes the representation of a arangoProfile and creates it.  Returns the server's representation of the arangoProfile, and an error, if there is any.
func (c *FakeArangoProfiles) Create(ctx context.Context, arangoProfile *v1alpha1.ArangoProfile, opts v1.CreateOptions) (result *v1alpha1.ArangoProfile, err error) {
	emptyResult := &v1alpha1.ArangoProfile{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(arangoprofilesResource, c.ns, arangoProfile, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ArangoProfile), err
}

// Update takes the representation of a arangoProfile and updates it. Returns the server's representation of the arangoProfile, and an error, if there is any.
func (c *FakeArangoProfiles) Update(ctx context.Context, arangoProfile *v1alpha1.ArangoProfile, opts v1.UpdateOptions) (result *v1alpha1.ArangoProfile, err error) {
	emptyResult := &v1alpha1.ArangoProfile{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(arangoprofilesResource, c.ns, arangoProfile, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ArangoProfile), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeArangoProfiles) UpdateStatus(ctx context.Context, arangoProfile *v1alpha1.ArangoProfile, opts v1.UpdateOptions) (result *v1alpha1.ArangoProfile, err error) {
	emptyResult := &v1alpha1.ArangoProfile{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(arangoprofilesResource, "status", c.ns, arangoProfile, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ArangoProfile), err
}

// Delete takes name of the arangoProfile and deletes it. Returns an error if one occurs.
func (c *FakeArangoProfiles) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(arangoprofilesResource, c.ns, name, opts), &v1alpha1.ArangoProfile{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeArangoProfiles) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(arangoprofilesResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.ArangoProfileList{})
	return err
}

// Patch applies the patch and returns the patched arangoProfile.
func (c *FakeArangoProfiles) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ArangoProfile, err error) {
	emptyResult := &v1alpha1.ArangoProfile{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(arangoprofilesResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1alpha1.ArangoProfile), err
}
