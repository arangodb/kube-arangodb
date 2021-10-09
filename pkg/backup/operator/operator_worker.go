//
// DISCLAIMER
//
// Copyright 2016-2021 ArangoDB GmbH, Cologne, Germany
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

import (
	"github.com/arangodb/kube-arangodb/pkg/backup/operator/operation"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (o *operator) worker() {
	for o.processNextItem() {

	}
}

func (o *operator) processNextItem() bool {
	defer func() {
		// Recover from panic to not shutdown whole operator
		if err := recover(); err != nil {
			e := o.logger.Error()

			switch obj := err.(type) {
			case error:
				e = e.AnErr("err", obj)
			case string:
				e = e.Str("err", obj)
			case int:
				e = e.Int("err", obj)
			default:
				e.Interface("err", obj)
			}

			e.Msgf("Recovered from panic")
		}
	}()

	obj, shutdown := o.workqueue.Get()

	if shutdown {
		return false
	}

	err := o.processObject(obj)

	if err != nil {
		o.logger.Error().Err(err).Interface("object", obj).Msgf("Error during object handling")
		return true
	}

	return true
}

func (o *operator) processObject(obj interface{}) error {
	defer o.workqueue.Done(obj)
	var item operation.Item
	var key string
	var ok bool
	var err error

	if key, ok = obj.(string); !ok {
		o.workqueue.Forget(obj)
		return nil
	}

	if item, err = operation.NewItemFromString(key); err != nil {
		o.workqueue.Forget(obj)
		return nil
	}

	if item.Operation != operation.Update {
		item.Operation = operation.Update
		o.workqueue.Forget(obj)
		o.workqueue.Add(item.String())
		return nil
	}

	o.objectProcessed.Inc()

	o.logger.Trace().Msgf("Received Item Action: %s, Type: %s/%s/%s, Namespace: %s, Name: %s",
		item.Operation,
		item.Group,
		item.Version,
		item.Kind,
		item.Namespace,
		item.Name)

	if err = o.processItem(item); err != nil {
		o.workqueue.AddRateLimited(key)
		return errors.Newf("error syncing '%s': %s, requeuing", key, err.Error())
	}

	o.logger.Trace().Msgf("Processed Item Action: %s, Type: %s/%s/%s, Namespace: %s, Name: %s",
		item.Operation,
		item.Group,
		item.Version,
		item.Kind,
		item.Namespace,
		item.Name)

	o.workqueue.Forget(obj)
	return nil
}

func (o *operator) processItem(item operation.Item) error {
	for _, handler := range o.handlers {
		if handler.CanBeHandled(item) {
			return handler.Handle(item)
		}
	}

	return nil
}
