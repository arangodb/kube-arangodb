//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
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

// Code generated by lister-gen. DO NOT EDIT.

package v1beta1

import (
	v1beta1 "github.com/arangodb/kube-arangodb/pkg/apis/scheduler/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ArangoSchedulerBatchJobLister helps list ArangoSchedulerBatchJobs.
// All objects returned here must be treated as read-only.
type ArangoSchedulerBatchJobLister interface {
	// List lists all ArangoSchedulerBatchJobs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.ArangoSchedulerBatchJob, err error)
	// ArangoSchedulerBatchJobs returns an object that can list and get ArangoSchedulerBatchJobs.
	ArangoSchedulerBatchJobs(namespace string) ArangoSchedulerBatchJobNamespaceLister
	ArangoSchedulerBatchJobListerExpansion
}

// arangoSchedulerBatchJobLister implements the ArangoSchedulerBatchJobLister interface.
type arangoSchedulerBatchJobLister struct {
	indexer cache.Indexer
}

// NewArangoSchedulerBatchJobLister returns a new ArangoSchedulerBatchJobLister.
func NewArangoSchedulerBatchJobLister(indexer cache.Indexer) ArangoSchedulerBatchJobLister {
	return &arangoSchedulerBatchJobLister{indexer: indexer}
}

// List lists all ArangoSchedulerBatchJobs in the indexer.
func (s *arangoSchedulerBatchJobLister) List(selector labels.Selector) (ret []*v1beta1.ArangoSchedulerBatchJob, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ArangoSchedulerBatchJob))
	})
	return ret, err
}

// ArangoSchedulerBatchJobs returns an object that can list and get ArangoSchedulerBatchJobs.
func (s *arangoSchedulerBatchJobLister) ArangoSchedulerBatchJobs(namespace string) ArangoSchedulerBatchJobNamespaceLister {
	return arangoSchedulerBatchJobNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ArangoSchedulerBatchJobNamespaceLister helps list and get ArangoSchedulerBatchJobs.
// All objects returned here must be treated as read-only.
type ArangoSchedulerBatchJobNamespaceLister interface {
	// List lists all ArangoSchedulerBatchJobs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1beta1.ArangoSchedulerBatchJob, err error)
	// Get retrieves the ArangoSchedulerBatchJob from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1beta1.ArangoSchedulerBatchJob, error)
	ArangoSchedulerBatchJobNamespaceListerExpansion
}

// arangoSchedulerBatchJobNamespaceLister implements the ArangoSchedulerBatchJobNamespaceLister
// interface.
type arangoSchedulerBatchJobNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ArangoSchedulerBatchJobs in the indexer for a given namespace.
func (s arangoSchedulerBatchJobNamespaceLister) List(selector labels.Selector) (ret []*v1beta1.ArangoSchedulerBatchJob, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1beta1.ArangoSchedulerBatchJob))
	})
	return ret, err
}

// Get retrieves the ArangoSchedulerBatchJob from the indexer for a given namespace and name.
func (s arangoSchedulerBatchJobNamespaceLister) Get(name string) (*v1beta1.ArangoSchedulerBatchJob, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1beta1.Resource("arangoschedulerbatchjob"), name)
	}
	return obj.(*v1beta1.ArangoSchedulerBatchJob), nil
}