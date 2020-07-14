package features

import "github.com/arangodb/go-driver"

var _ Feature = &feature{}

type Feature interface {
	Name() string
	Description() string
	LongDescription() string
	Version() driver.Version
	EnterpriseRequired() bool
	EnabledByDefault() bool
	EnabledPointer() *bool
}

type feature struct {
	name, description, longDescription string
	version driver.Version
	enterpriseRequired,	enabledByDefault, enabled bool
}

func (f feature) LongDescription() string {
	return f.longDescription
}

func (f *feature) EnabledPointer() *bool {
	return &f.enabled
}

func (f feature) Version() driver.Version {
	return f.version
}

func (f feature) EnterpriseRequired() bool {
	return f.enterpriseRequired
}

func (f feature) EnabledByDefault() bool {
	return f.enabledByDefault
}

func (f feature) Name() string {
	return f.name
}

func (f feature) Description() string {
	return f.description
}