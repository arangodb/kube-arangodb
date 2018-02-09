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
package v1alpha

import (
	v1alpha "github.com/arangodb/k8s-operator/pkg/apis/arangodb/v1alpha"
	scheme "github.com/arangodb/k8s-operator/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ArangoClustersGetter has a method to return a ArangoClusterInterface.
// A group's client should implement this interface.
type ArangoClustersGetter interface {
	ArangoClusters(namespace string) ArangoClusterInterface
}

// ArangoClusterInterface has methods to work with ArangoCluster resources.
type ArangoClusterInterface interface {
	Create(*v1alpha.ArangoCluster) (*v1alpha.ArangoCluster, error)
	Update(*v1alpha.ArangoCluster) (*v1alpha.ArangoCluster, error)
	UpdateStatus(*v1alpha.ArangoCluster) (*v1alpha.ArangoCluster, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha.ArangoCluster, error)
	List(opts v1.ListOptions) (*v1alpha.ArangoClusterList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha.ArangoCluster, err error)
	ArangoClusterExpansion
}

// arangoClusters implements ArangoClusterInterface
type arangoClusters struct {
	client rest.Interface
	ns     string
}

// newArangoClusters returns a ArangoClusters
func newArangoClusters(c *ClusterV1alphaClient, namespace string) *arangoClusters {
	return &arangoClusters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the arangoCluster, and returns the corresponding arangoCluster object, and an error if there is any.
func (c *arangoClusters) Get(name string, options v1.GetOptions) (result *v1alpha.ArangoCluster, err error) {
	result = &v1alpha.ArangoCluster{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("arangoclusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ArangoClusters that match those selectors.
func (c *arangoClusters) List(opts v1.ListOptions) (result *v1alpha.ArangoClusterList, err error) {
	result = &v1alpha.ArangoClusterList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("arangoclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested arangoClusters.
func (c *arangoClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("arangoclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a arangoCluster and creates it.  Returns the server's representation of the arangoCluster, and an error, if there is any.
func (c *arangoClusters) Create(arangoCluster *v1alpha.ArangoCluster) (result *v1alpha.ArangoCluster, err error) {
	result = &v1alpha.ArangoCluster{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("arangoclusters").
		Body(arangoCluster).
		Do().
		Into(result)
	return
}

// Update takes the representation of a arangoCluster and updates it. Returns the server's representation of the arangoCluster, and an error, if there is any.
func (c *arangoClusters) Update(arangoCluster *v1alpha.ArangoCluster) (result *v1alpha.ArangoCluster, err error) {
	result = &v1alpha.ArangoCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("arangoclusters").
		Name(arangoCluster.Name).
		Body(arangoCluster).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *arangoClusters) UpdateStatus(arangoCluster *v1alpha.ArangoCluster) (result *v1alpha.ArangoCluster, err error) {
	result = &v1alpha.ArangoCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("arangoclusters").
		Name(arangoCluster.Name).
		SubResource("status").
		Body(arangoCluster).
		Do().
		Into(result)
	return
}

// Delete takes name of the arangoCluster and deletes it. Returns an error if one occurs.
func (c *arangoClusters) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("arangoclusters").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *arangoClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("arangoclusters").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched arangoCluster.
func (c *arangoClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha.ArangoCluster, err error) {
	result = &v1alpha.ArangoCluster{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("arangoclusters").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
