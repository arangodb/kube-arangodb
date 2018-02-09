//
// DISCLAIMER
//
// Copyright 2018 ArangoDB GmbH, Cologne, Germany
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
package fake

import (
	v1alpha "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeArangoClusters implements ArangoClusterInterface
type FakeArangoClusters struct {
	Fake *FakeClusterV1alpha
	ns   string
}

var arangoclustersResource = schema.GroupVersionResource{Group: "cluster.arangodb.com", Version: "v1alpha", Resource: "arangoclusters"}

var arangoclustersKind = schema.GroupVersionKind{Group: "cluster.arangodb.com", Version: "v1alpha", Kind: "ArangoCluster"}

// Get takes name of the arangoCluster, and returns the corresponding arangoCluster object, and an error if there is any.
func (c *FakeArangoClusters) Get(name string, options v1.GetOptions) (result *v1alpha.ArangoCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(arangoclustersResource, c.ns, name), &v1alpha.ArangoCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.ArangoCluster), err
}

// List takes label and field selectors, and returns the list of ArangoClusters that match those selectors.
func (c *FakeArangoClusters) List(opts v1.ListOptions) (result *v1alpha.ArangoClusterList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(arangoclustersResource, arangoclustersKind, c.ns, opts), &v1alpha.ArangoClusterList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha.ArangoClusterList{}
	for _, item := range obj.(*v1alpha.ArangoClusterList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested arangoClusters.
func (c *FakeArangoClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(arangoclustersResource, c.ns, opts))

}

// Create takes the representation of a arangoCluster and creates it.  Returns the server's representation of the arangoCluster, and an error, if there is any.
func (c *FakeArangoClusters) Create(arangoCluster *v1alpha.ArangoCluster) (result *v1alpha.ArangoCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(arangoclustersResource, c.ns, arangoCluster), &v1alpha.ArangoCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.ArangoCluster), err
}

// Update takes the representation of a arangoCluster and updates it. Returns the server's representation of the arangoCluster, and an error, if there is any.
func (c *FakeArangoClusters) Update(arangoCluster *v1alpha.ArangoCluster) (result *v1alpha.ArangoCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(arangoclustersResource, c.ns, arangoCluster), &v1alpha.ArangoCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.ArangoCluster), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeArangoClusters) UpdateStatus(arangoCluster *v1alpha.ArangoCluster) (*v1alpha.ArangoCluster, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(arangoclustersResource, "status", c.ns, arangoCluster), &v1alpha.ArangoCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.ArangoCluster), err
}

// Delete takes name of the arangoCluster and deletes it. Returns an error if one occurs.
func (c *FakeArangoClusters) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(arangoclustersResource, c.ns, name), &v1alpha.ArangoCluster{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeArangoClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(arangoclustersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha.ArangoClusterList{})
	return err
}

// Patch applies the patch and returns the patched arangoCluster.
func (c *FakeArangoClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha.ArangoCluster, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(arangoclustersResource, c.ns, name, data, subresources...), &v1alpha.ArangoCluster{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha.ArangoCluster), err
}
