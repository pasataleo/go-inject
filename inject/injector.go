package inject

import (
	"reflect"

	"github.com/pasataleo/go-errors/errors"
)

type Creator[T any] func(injector *Injector) (T, error)

type Parser[T any] func(values []string) (T, error)

type Module func(injector *Injector) error

type Injector struct {
	// staticBindings and staticTypes are used to store bindings and types for
	// static values.
	staticBindings map[string]interface{}
	staticTypes    map[reflect.Type]interface{}

	// functionBindings and functionTypes are used to store bindings and types
	// for functions.
	functionBindings map[string]Creator[any]
	functionTypes    map[reflect.Type]Creator[any]

	// parsers map flag names and environment variables to parser functions.
	parsers map[string]Parser[any]
}

func NewInjector() *Injector {
	return &Injector{
		staticTypes:      make(map[reflect.Type]interface{}),
		staticBindings:   make(map[string]interface{}),
		functionBindings: make(map[string]Creator[any]),
		functionTypes:    make(map[reflect.Type]Creator[any]),
		parsers:          make(map[string]Parser[any]),
	}
}

func (i *Injector) Install(module Module) error {
	return module(i)
}

func (i *Injector) Inject(injectee interface{}) error {
	ptr := reflect.TypeOf(injectee)
	if ptr.Kind() != reflect.Ptr {
		panic("Inject value must be a pointer")
	}

	t, v := ptr.Elem(), reflect.ValueOf(injectee).Elem()
	if ok := i.HasType(t); ok {
		value, err := i.GetType(t)
		if err != nil {
			return errors.Wrap(err, "Failed to create value")
		}
		v.Set(reflect.ValueOf(value))
		return nil
	}

	if t.Kind() == reflect.Struct {
		// We can try and inject the fields of the struct, but can't do that with any other kind of type.
		for ix := 0; ix < t.NumField(); ix++ {
			fieldType := t.Field(ix)
			fieldValue := v.Field(ix)

			if !fieldType.IsExported() {
				// Can't inject unexported fields
				continue
			}

			// First, check if the field has a tag that contains the name of the binding.
			if inject, ok := fieldType.Tag.Lookup("inject"); ok {
				if inject == "-" {
					// We can skip the field if the tag is "-"
					continue
				}

				value, err := i.Get(inject)
				if err != nil {
					return errors.Wrap(err, "Failed to create value")
				}
				fieldValue.Set(reflect.ValueOf(value))
				continue
			}

			// Second, if the field doesn't have a tag, we can try and inject it by type
			if fieldType.Type.Kind() == reflect.Ptr {
				value := reflect.New(fieldType.Type.Elem())
				if err := i.Inject(value.Interface()); err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				fieldValue.Set(value)
				continue
			} else {
				value := reflect.New(fieldType.Type)
				if err := i.Inject(value.Interface()); err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				fieldValue.Set(value.Elem())
				continue
			}
		}

		return nil
	}

	return errors.Newf(nil, errors.ErrorCodeUnknown, "No binding found for %s", t)
}

func (i *Injector) Get(identifier string) (interface{}, error) {
	if creator, ok := i.functionBindings[identifier]; ok {
		return creator(i)
	}
	if value, ok := i.staticBindings[identifier]; ok {
		return value, nil
	}
	return nil, errors.Newf(nil, errors.ErrorCodeUnknown, "No binding found for %s", identifier)
}

func (i *Injector) Has(identifier string) bool {
	if _, ok := i.functionBindings[identifier]; ok {
		return true
	}
	if _, ok := i.staticBindings[identifier]; ok {
		return true
	}
	if _, ok := i.parsers[identifier]; ok {
		return true
	}
	return false
}

func (i *Injector) GetUnsafe(identifier string) interface{} {
	value, err := i.Get(identifier)
	if err != nil {
		panic(err)
	}
	return value
}

func (i *Injector) GetType(ty reflect.Type) (interface{}, error) {
	if creator, ok := i.functionTypes[ty]; ok {
		return creator(i)
	}
	if value, ok := i.staticTypes[ty]; ok {
		return value, nil
	}
	return nil, errors.Newf(nil, errors.ErrorCodeUnknown, "No binding found for %s", ty)
}

func (i *Injector) HasType(ty reflect.Type) bool {
	if _, ok := i.functionTypes[ty]; ok {
		return true
	}
	if _, ok := i.staticTypes[ty]; ok {
		return true
	}
	return false
}

func (i *Injector) GetTypeUnsafe(ty reflect.Type) interface{} {
	value, err := i.GetType(ty)
	if err != nil {
		panic(err)
	}
	return value
}
