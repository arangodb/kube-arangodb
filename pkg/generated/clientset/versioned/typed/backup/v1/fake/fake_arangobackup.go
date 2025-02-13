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

	v1 "github.com/arangodb/kube-arangodb/pkg/apis/backup/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeArangoBackups implements ArangoBackupInterface
type FakeArangoBackups struct {
	Fake *FakeBackupV1
	ns   string
}

var arangobackupsResource = v1.SchemeGroupVersion.WithResource("arangobackups")

var arangobackupsKind = v1.SchemeGroupVersion.WithKind("ArangoBackup")

// Get takes name of the arangoBackup, and returns the corresponding arangoBackup object, and an error if there is any.
func (c *FakeArangoBackups) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.ArangoBackup, err error) {
	emptyResult := &v1.ArangoBackup{}
	obj, err := c.Fake.
		Invokes(testing.NewGetActionWithOptions(arangobackupsResource, c.ns, name, options), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoBackup), err
}

// List takes label and field selectors, and returns the list of ArangoBackups that match those selectors.
func (c *FakeArangoBackups) List(ctx context.Context, opts metav1.ListOptions) (result *v1.ArangoBackupList, err error) {
	emptyResult := &v1.ArangoBackupList{}
	obj, err := c.Fake.
		Invokes(testing.NewListActionWithOptions(arangobackupsResource, arangobackupsKind, c.ns, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.ArangoBackupList{ListMeta: obj.(*v1.ArangoBackupList).ListMeta}
	for _, item := range obj.(*v1.ArangoBackupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested arangoBackups.
func (c *FakeArangoBackups) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchActionWithOptions(arangobackupsResource, c.ns, opts))

}

// Create takes the representation of a arangoBackup and creates it.  Returns the server's representation of the arangoBackup, and an error, if there is any.
func (c *FakeArangoBackups) Create(ctx context.Context, arangoBackup *v1.ArangoBackup, opts metav1.CreateOptions) (result *v1.ArangoBackup, err error) {
	emptyResult := &v1.ArangoBackup{}
	obj, err := c.Fake.
		Invokes(testing.NewCreateActionWithOptions(arangobackupsResource, c.ns, arangoBackup, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoBackup), err
}

// Update takes the representation of a arangoBackup and updates it. Returns the server's representation of the arangoBackup, and an error, if there is any.
func (c *FakeArangoBackups) Update(ctx context.Context, arangoBackup *v1.ArangoBackup, opts metav1.UpdateOptions) (result *v1.ArangoBackup, err error) {
	emptyResult := &v1.ArangoBackup{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateActionWithOptions(arangobackupsResource, c.ns, arangoBackup, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoBackup), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeArangoBackups) UpdateStatus(ctx context.Context, arangoBackup *v1.ArangoBackup, opts metav1.UpdateOptions) (result *v1.ArangoBackup, err error) {
	emptyResult := &v1.ArangoBackup{}
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceActionWithOptions(arangobackupsResource, "status", c.ns, arangoBackup, opts), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoBackup), err
}

// Delete takes name of the arangoBackup and deletes it. Returns an error if one occurs.
func (c *FakeArangoBackups) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(arangobackupsResource, c.ns, name, opts), &v1.ArangoBackup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeArangoBackups) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionActionWithOptions(arangobackupsResource, c.ns, opts, listOpts)

	_, err := c.Fake.Invokes(action, &v1.ArangoBackupList{})
	return err
}

// Patch applies the patch and returns the patched arangoBackup.
func (c *FakeArangoBackups) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.ArangoBackup, err error) {
	emptyResult := &v1.ArangoBackup{}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceActionWithOptions(arangobackupsResource, c.ns, name, pt, data, opts, subresources...), emptyResult)

	if obj == nil {
		return emptyResult, err
	}
	return obj.(*v1.ArangoBackup), err
}
