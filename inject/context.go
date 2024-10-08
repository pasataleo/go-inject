package inject

import (
	"reflect"

	"github.com/pasataleo/go-errors/errors"
)

var (
	_ Context = (*staticContext[any])(nil)
	_ Context = (*functionContext[any])(nil)
)

type Context interface {
	ToSafe(injector *Injector, args ...string) error
	To(injector *Injector, args ...string)
}

type staticContext[T any] struct {
	value T
}

func (c *staticContext[T]) ToSafe(injector *Injector, args ...string) error {
	if len(args) == 0 {
		var value [0]T
		t := reflect.TypeOf(value).Elem()

		if injector.HasType(t) {
			return errors.Newf(nil, errors.ErrorCodeUnknown, "Binding already exists for %s", t)
		}

		injector.staticTypes[t] = c.value
	}

	for _, arg := range args {
		if injector.Has(arg) {
			return errors.Newf(nil, errors.ErrorCodeUnknown, "Binding already exists for %s", arg)
		}

		injector.staticBindings[arg] = c.value
	}

	return nil
}

func (c *staticContext[T]) To(injector *Injector, args ...string) {
	if err := c.ToSafe(injector, args...); err != nil {
		panic(err)
	}
}

type functionContext[T any] struct {
	creator Creator[T]
}

func (c *functionContext[T]) ToSafe(injector *Injector, args ...string) error {
	if len(args) == 0 {
		var value [0]T
		t := reflect.TypeOf(value).Elem()

		if injector.HasType(t) {
			return errors.Newf(nil, errors.ErrorCodeUnknown, "Binding already exists for %s", t)
		}

		injector.functionTypes[t] = func(injector *Injector) (any, error) {
			return c.creator(injector)
		}
	}

	for _, arg := range args {
		if injector.Has(arg) {
			return errors.Newf(nil, errors.ErrorCodeUnknown, "Binding already exists for %s", arg)
		}

		injector.functionBindings[arg] = func(injector *Injector) (any, error) {
			return c.creator(injector)
		}
	}
	return nil
}

func (c *functionContext[T]) To(injector *Injector, args ...string) {
	if err := c.ToSafe(injector, args...); err != nil {
		panic(err)
	}
}
