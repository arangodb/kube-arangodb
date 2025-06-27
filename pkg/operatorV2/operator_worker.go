//
// DISCLAIMER
//
// Copyright 2016-2025 ArangoDB GmbH, Cologne, Germany
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

package operator

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/arangodb/kube-arangodb/pkg/generated/metric_descriptions"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
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
			e := loggerWorker.Str("type", "worker")

			switch obj := err.(type) {
			case error:
				e = e.Err(obj)
			case string:
				e = e.Str("err", obj)
			case int:
				e = e.Int("err", obj)
			default:
				e.Interface("err", obj)
			}

			if v := debug.Stack(); len(v) != 0 {
				e = e.Str("stack", string(v))
			}

			e.Error("Recovered from panic")
		}
	}()

	obj, shutdown := o.workqueue.Get()

	if shutdown {
		return false
	}

	err := o.processObject(obj)

	if err != nil {
		loggerWorker.Stack().Err(err).Interface("object", obj).Error("Error during object handling: %v", err)
		return true
	}

	return true
}

func (o *operator) processObject(item operation.Item) error {
	defer o.workqueue.Done(item)
	var err error

	if item.Operation != operation.Update {
		o.workqueue.Forget(item)
		item.Operation = operation.Update
		o.workqueue.Add(item)
		return nil
	}

	metric_descriptions.GlobalArangodbOperatorObjectsProcessedCounter().Inc(metric_descriptions.NewArangodbOperatorObjectsProcessedInput(o.operator.name))

	loggerWorker.Trace("Received Item Action: %s, Type: %s/%s/%s, Namespace: %s, Name: %s",
		item.Operation,
		item.Group,
		item.Version,
		item.Kind,
		item.Namespace,
		item.Name)

	if err = o.processItem(item); err != nil {
		o.workqueue.AddRateLimited(item)

		if !IsReconcile(err) {
			message := fmt.Sprintf("error syncing '%s': %s, re-queuing", item.String(), err.Error())
			loggerWorker.Debug(message)
			return errors.Errorf(message)
		}

		return nil
	}

	loggerWorker.Trace("Processed Item Action: %s, Type: %s/%s/%s, Namespace: %s, Name: %s",
		item.Operation,
		item.Group,
		item.Version,
		item.Kind,
		item.Namespace,
		item.Name)

	o.workqueue.Forget(item)
	return nil
}

func (o *operator) processItem(item operation.Item) error {
	for _, handler := range o.handlers {
		if handler.CanBeHandled(item) {
			return o.processItemWithCTX(item, handler)
		}
	}

	return nil
}

func (o *operator) processItemWithCTX(item operation.Item, handler Handler) error {
	ctx, c := WithHandlerTimeout(context.Background(), handler)
	defer c()

	return handler.Handle(ctx, item)
}
