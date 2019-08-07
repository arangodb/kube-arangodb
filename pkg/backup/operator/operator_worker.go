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
// Author Adam Janikowski
//

package operator

import "fmt"

func (o *operator) worker() {
	for o.processNextItem() {

	}
}

func (o *operator) processNextItem() bool {
	obj, shutdown := o.workqueue.Get()

	if shutdown {
		return false
	}

	err := o.processObject(obj)

	if err != nil {
		return true
	}

	return true
}

func (o *operator) processObject(obj interface{}) error {
	defer o.workqueue.Done(obj)
	var item Item
	var key string
	var ok bool
	var err error

	if key, ok = obj.(string); !ok {
		o.workqueue.Forget(obj)
		return nil
	}

	if item, err = NewItemFromString(key); !ok {
		o.workqueue.Forget(obj)
		return nil
	}

	if err = o.processItem(item); err != nil {
		o.workqueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
	}

	o.workqueue.Forget(obj)
	return nil
}

func (o *operator) processItem(item Item) error {
	for _, handler := range o.handlers {
		if handler.CanBeHandled(item) {
			return handler.Handle(item)
		}
	}

	return nil
}
