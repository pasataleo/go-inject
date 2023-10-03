package inject

import (
	"reflect"

	"github.com/pasataleo/go-errors/errors"
)

type Context interface {
	To(injector *Injector, args ...string) error
	ToUnsafe(injector *Injector, args ...string)
}

type staticContext[T any] struct {
	value T
}

func (c *staticContext[T]) To(injector *Injector, args ...string) error {
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

func (c *staticContext[T]) ToUnsafe(injector *Injector, args ...string) {
	if err := c.To(injector, args...); err != nil {
		panic(err)
	}
}

type functionContext[T any] struct {
	creator Creator[T]
}

func (c *functionContext[T]) To(injector *Injector, args ...string) error {
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

func (c *functionContext[T]) ToUnsafe(injector *Injector, args ...string) {
	if err := c.To(injector, args...); err != nil {
		panic(err)
	}
}

type parserContext[T any] struct {
	parser Parser[T]
}

func (c *parserContext[T]) To(injector *Injector, args ...string) error {
	if len(args) == 0 {
		return errors.Newf(nil, errors.ErrorCodeUnknown, "Parser requires at least one argument")
	}
	for _, arg := range args {
		if injector.Has(arg) {
			return errors.Newf(nil, errors.ErrorCodeUnknown, "Binding already exists for %s", arg)
		}

		injector.parsers[arg] = func(values []string) (any, error) {
			return c.parser(values)
		}
	}
	return nil
}

func (c *parserContext[T]) ToUnsafe(injector *Injector, args ...string) {
	if err := c.To(injector, args...); err != nil {
		panic(err)
	}
}
