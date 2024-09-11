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

package integrations

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/arangodb/kube-arangodb/pkg/util"
	"github.com/arangodb/kube-arangodb/pkg/util/errors"
)

func NewFlagEnvHandler(fs *flag.FlagSet) FlagEnvHandler {
	return flagEnvHandler{
		fs: fs,
	}
}

type FlagEnvHandler interface {
	WithPrefix(prefix string) FlagEnvHandler

	StringVar(p *string, name string, value string, usage string) error
	String(name string, value string, usage string) error

	StringSliceVar(p *[]string, name string, value []string, usage string) error
	StringSlice(name string, value []string, usage string) error

	BoolVar(p *bool, name string, value bool, usage string) error
	Bool(name string, value bool, usage string) error

	Uint16Var(p *uint16, name string, value uint16, usage string) error
	Uint16(name string, value uint16, usage string) error

	DurationVar(p *time.Duration, name string, value time.Duration, usage string) error
	Duration(name string, value time.Duration, usage string) error
}

type flagEnvHandler struct {
	prefix string
	fs     *flag.FlagSet
}

func (f flagEnvHandler) StringVar(p *string, name string, value string, usage string) error {
	v, err := parseEnvToString(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.StringVar(p, f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) String(name string, value string, usage string) error {
	v, err := parseEnvToString(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.String(f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) StringSliceVar(p *[]string, name string, value []string, usage string) error {
	v, err := parseEnvToStringArray(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.StringSliceVar(p, f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) StringSlice(name string, value []string, usage string) error {
	v, err := parseEnvToStringArray(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.StringSlice(f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) BoolVar(p *bool, name string, value bool, usage string) error {
	v, err := parseEnvToBool(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.BoolVar(p, f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) Bool(name string, value bool, usage string) error {
	v, err := parseEnvToBool(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.Bool(f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) DurationVar(p *time.Duration, name string, value time.Duration, usage string) error {
	v, err := parseEnvToDuration(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.DurationVar(p, f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) Duration(name string, value time.Duration, usage string) error {
	v, err := parseEnvToDuration(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.Duration(f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) Uint16Var(p *uint16, name string, value uint16, usage string) error {
	v, err := parseEnvToUint16(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.Uint16Var(p, f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) Uint16(name string, value uint16, usage string) error {
	v, err := parseEnvToUint16(f.getEnv(name), value)
	if err != nil {
		return err
	}

	f.fs.Uint16(f.name(name), v, f.varDesc(name, usage))

	return nil
}

func (f flagEnvHandler) varDesc(name string, dest string) string {
	return fmt.Sprintf("%s (Env: %s)", dest, f.getEnv(name))
}

func (f flagEnvHandler) getEnv(n string) string {
	z := f.name(n)

	z = strings.ReplaceAll(z, ".", "_")
	z = strings.ReplaceAll(z, "-", "_")

	return strings.ToUpper(z)
}
func (f flagEnvHandler) name(n string) string {
	if f.prefix == "" {
		return n
	}
	if n == "" {
		return f.prefix
	}
	return fmt.Sprintf("%s.%s", f.prefix, n)
}

func (f flagEnvHandler) WithPrefix(prefix string) FlagEnvHandler {
	return flagEnvHandler{
		prefix: f.name(prefix),
		fs:     f.fs,
	}
}

func parseEnvToDuration(env string, def time.Duration) (time.Duration, error) {
	return parseEnvToType(env, def, time.ParseDuration)
}

func parseEnvToUint16(env string, def uint16) (uint16, error) {
	return parseEnvToType(env, def, func(in string) (uint16, error) {
		v, err := strconv.ParseUint(in, 10, 16)
		return uint16(v), err
	})
}

func parseEnvToBool(env string, def bool) (bool, error) {
	return parseEnvToType(env, def, strconv.ParseBool)
}

func parseEnvToStringArray(env string, def []string) ([]string, error) {
	return parseEnvToType(env, def, func(in string) ([]string, error) {
		return strings.Split(in, ","), nil
	})
}

func parseEnvToString(env string, def string) (string, error) {
	return parseEnvToType(env, def, func(in string) (string, error) {
		return in, nil
	})
}

func parseEnvToType[T any](env string, def T, parser func(in string) (T, error)) (T, error) {
	if v, ok := os.LookupEnv(env); ok {
		if q, err := parser(v); err != nil {
			return util.Default[T](), errors.Wrapf(err, "Unable to parse env `%s` as %s", env, reflect.TypeOf(def).String())
		} else {
			return q, nil
		}
	}

	return def, nil
}
