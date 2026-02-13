package types

import (
	"sort"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func (x *Role) Deleted() bool {
	return x == nil
}

func (x *Role) Clean() error {
	if x == nil {
		return nil
	}

	sort.Strings(x.Users)
	sort.Strings(x.Policies)

	x.Users = util.UniqueList(x.Users)
	x.Policies = util.UniqueList(x.Policies)

	return nil
}

func (x *Role) Validate() error {
	if x == nil {
		return errors.Errorf("Nil not allowed")
	}

	return nil
}
