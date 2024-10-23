package storage

import (
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arangodb/kube-arangodb/pkg/apis/apps"
	appsApi "github.com/arangodb/kube-arangodb/pkg/apis/apps/v1"
	fakeClientSet "github.com/arangodb/kube-arangodb/pkg/generated/clientset/versioned/fake"
	operator "github.com/arangodb/kube-arangodb/pkg/operatorV2"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/event"
	"github.com/arangodb/kube-arangodb/pkg/operatorV2/operation"
)

func newFakeHandler() *handler {
	f := fakeClientSet.NewSimpleClientset()
	k := fake.NewSimpleClientset()

	h := &handler{
		client:        f,
		kubeClient:    k,
		eventRecorder: event.NewEventRecorder("mock", k).NewInstance(Group(), Version(), Kind()),
		operator:      operator.NewOperator("mock", "mock", "mock"),
	}

	return h
}

func newItem(o operation.Operation, namespace, name string) operation.Item {
	return operation.Item{
		Group:   appsApi.SchemeGroupVersion.Group,
		Version: appsApi.SchemeGroupVersion.Version,
		Kind:    apps.ArangoJobResourceKind,

		Operation: o,

		Namespace: namespace,
		Name:      name,
	}
}
