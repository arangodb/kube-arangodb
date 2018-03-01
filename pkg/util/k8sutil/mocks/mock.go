package mocks

import "github.com/stretchr/testify/mock"

type MockGetter interface {
	AsMock() *mock.Mock
}

// AsMock performs a typeconversion to *Mock.
func AsMock(obj interface{}) *mock.Mock {
	return obj.(MockGetter).AsMock()
}
