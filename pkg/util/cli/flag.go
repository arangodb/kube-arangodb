//
// DISCLAIMER
//
// Copyright 2025 ArangoDB GmbH, Cologne, Germany
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

package cli

import (
	"reflect"
	"time"

	"github.com/spf13/cobra"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func ValidateFlags(flags ...FlagRegisterer) RunE {
	return func(cmd *cobra.Command, args []string) error {
		return errors.Errors(util.FormatList[FlagRegisterer, error](flags, func(registerer FlagRegisterer) error {
			return errors.Wrapf(registerer.Validate(cmd), "Error while validating arg --%s", registerer.GetName())
		})...)
	}
}

func RegisterFlags(cmd *cobra.Command, flags ...FlagRegisterer) error {
	for _, f := range flags {
		if err := f.Register(cmd); err != nil {
			return err
		}
	}

	return nil
}

type FlagRegisterer interface {
	GetName() string

	Register(cmd *cobra.Command) error
	Validate(cmd *cobra.Command) error
}

type FlagInterface[T any] interface {
	FlagRegisterer
	Get(cmd *cobra.Command) (T, error)
}

type Flag[T any] struct {
	Name        string
	Short       string
	Description string
	Default     T

	Persistent bool

	Check func(in T) error

	Deprecated *string
	Hidden     bool
}

func (f Flag[T]) GetName() string {
	return f.Name
}

func (f Flag[T]) Validate(cmd *cobra.Command) error {
	if cmd.Flags().Lookup(f.Name) == nil {
		return nil
	}

	v, err := f.Get(cmd)
	if err != nil {
		return err
	}

	if f.Check != nil {
		return f.Check(v)
	}

	return nil
}

func (f Flag[T]) Register(cmd *cobra.Command) error {
	flags := cmd.Flags()

	if f.Persistent {
		flags = cmd.PersistentFlags()
	}

	v := reflect.TypeOf(f.Default)

	z := reflect.ValueOf(f.Default).Interface()

	if v == util.TypeOf[string]() {
		v := z.(string)
		if short := f.Short; short == "" {
			flags.String(f.Name, v, f.Description)
		} else {
			flags.StringP(f.Name, short, v, f.Description)
		}
	} else if v == util.TypeOf[bool]() {
		v := z.(bool)
		if short := f.Short; short == "" {
			flags.Bool(f.Name, v, f.Description)
		} else {
			flags.BoolP(f.Name, short, v, f.Description)
		}
	} else if v == util.TypeOf[[]string]() {
		v := z.([]string)
		if short := f.Short; short == "" {
			flags.StringSlice(f.Name, v, f.Description)
		} else {
			flags.StringSliceP(f.Name, short, v, f.Description)
		}
	} else if v == util.TypeOf[time.Duration]() {
		v := z.(time.Duration)
		if short := f.Short; short == "" {
			flags.Duration(f.Name, v, f.Description)
		} else {
			flags.DurationP(f.Name, short, v, f.Description)
		}
	} else {
		return errors.Errorf("Unsupported type for kind: %s", reflect.ValueOf(f.Default).Type().String())
	}

	if q := f.Deprecated; q != nil {
		if err := flags.MarkDeprecated(f.Name, *q); err != nil {
			return err
		}
	}

	if f.Hidden {
		if err := flags.MarkHidden(f.Name); err != nil {
			return err
		}
	}

	return nil
}

func (f Flag[T]) Get(cmd *cobra.Command) (T, error) {

	v := reflect.TypeOf(f.Default)

	if v == util.TypeOf[string]() {
		v, err := cmd.Flags().GetString(f.Name)
		if err != nil {
			return util.Default[T](), err
		}

		q, ok := reflect.ValueOf(v).Interface().(T)
		if !ok {
			return util.Default[T](), errors.Errorf("Unable to parse type for kind: %s", reflect.ValueOf(f.Default).Type().String())
		}

		return q, nil
	} else if v == util.TypeOf[[]string]() {
		v, err := cmd.Flags().GetStringSlice(f.Name)
		if err != nil {
			return util.Default[T](), err
		}

		q, ok := reflect.ValueOf(v).Interface().(T)
		if !ok {
			return util.Default[T](), errors.Errorf("Unable to parse type for kind: %s", reflect.ValueOf(f.Default).Type().String())
		}

		return q, nil
	} else if v == util.TypeOf[bool]() {
		v, err := cmd.Flags().GetBool(f.Name)
		if err != nil {
			return util.Default[T](), err
		}

		q, ok := reflect.ValueOf(v).Interface().(T)
		if !ok {
			return util.Default[T](), errors.Errorf("Unable to parse type for kind: %s", reflect.ValueOf(f.Default).Type().String())
		}

		return q, nil
	} else if v == util.TypeOf[time.Duration]() {
		v, err := cmd.Flags().GetDuration(f.Name)
		if err != nil {
			return util.Default[T](), err
		}

		q, ok := reflect.ValueOf(v).Interface().(T)
		if !ok {
			return util.Default[T](), errors.Errorf("Unable to parse type for kind: %s", reflect.ValueOf(f.Default).Type().String())
		}

		return q, nil
	} else {
		return util.Default[T](), errors.Errorf("Unsupported type for kind: %s", reflect.ValueOf(f.Default).Type().String())
	}
}

func (f Flag[T]) AsInterface() FlagInterface[T] {
	return f
}
