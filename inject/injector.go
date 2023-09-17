package inject

import (
	"reflect"

	"github.com/pasataleo/go-errors/errors"
)

type Creator[T any] func(injector *Injector) (T, error)

type Module func(injector *Injector) error

type Injector struct {
	bindings map[string]Creator[any]
	types    map[reflect.Type]Creator[any]
}

func NewInjector() *Injector {
	return &Injector{
		bindings: make(map[string]Creator[any]),
		types:    make(map[reflect.Type]Creator[any]),
	}
}

func (i *Injector) Inject(injectee interface{}) error {
	ptr := reflect.TypeOf(injectee)
	if ptr.Kind() != reflect.Ptr {
		panic("Inject value must be a pointer")
	}

	t := ptr.Elem()
	if creator, ok := i.types[t]; ok {
		value, err := creator(i)
		if err != nil {
			return errors.Wrap(err, "Failed to create value")
		}
		reflect.ValueOf(injectee).Elem().Set(reflect.ValueOf(value))
		return nil
	}
	v := reflect.ValueOf(injectee).Elem()

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

				if creator, ok := i.bindings[inject]; ok {
					value, err := creator(i)
					if err != nil {
						return errors.Wrap(err, "Failed to create value")
					}
					fieldValue.Set(reflect.ValueOf(value))
					continue
				} else {
					return errors.Newf(nil, errors.ErrorCodeUnknown, "No binding found for %s", inject)
				}
			}

			// Second, if the field doesn't have a tag, we can try and inject it by type
			if fieldType.Type.Kind() == reflect.Ptr {
				value := reflect.New(fieldType.Type.Elem())
				if err := i.Inject(value.Interface()); err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				fieldValue.Set(value)
			} else {
				value := reflect.New(fieldType.Type)
				if err := i.Inject(value.Interface()); err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				fieldValue.Set(value.Elem())
			}

		}

		return nil
	}

	return errors.Newf(nil, errors.ErrorCodeUnknown, "No binding found for %s", t)
}

func (i *Injector) Get(identifier string) (interface{}, error) {
	if creator, ok := i.bindings[identifier]; ok {
		return creator(i)
	}
	return nil, errors.Newf(nil, errors.ErrorCodeUnknown, "No binding found for %s", identifier)
}

func (i *Injector) GetUnsafe(identifier string) interface{} {
	value, err := i.Get(identifier)
	if err != nil {
		panic(err)
	}
	return value
}

func (i *Injector) GetType(ty reflect.Type) (interface{}, error) {
	if creator, ok := i.types[ty]; ok {
		return creator(i)
	}
	return nil, errors.Newf(nil, errors.ErrorCodeUnknown, "No binding found for %s", ty)
}

func (i *Injector) GetTypeUnsafe(ty reflect.Type) interface{} {
	value, err := i.GetType(ty)
	if err != nil {
		panic(err)
	}
	return value
}

func (i *Injector) Bind(identifier string, creator Creator[any]) error {
	if _, ok := i.bindings[identifier]; ok {
		return errors.Newf(nil, errors.ErrorCodeUnknown, "Binding already exists for %s", identifier)
	}
	i.bindings[identifier] = creator
	return nil
}

func (i *Injector) BindUnsafe(identifier string, creator Creator[any]) {
	if err := i.Bind(identifier, creator); err != nil {
		panic(err)
	}
}

func Bind[T any](injector *Injector, creator Creator[T]) error {
	return injector.bindType(reflect.TypeOf(creator).Out(0), func(injector *Injector) (any, error) {
		return creator(injector)
	})
}

func BindUnsafe[T any](injector *Injector, creator Creator[T]) {
	if err := Bind(injector, creator); err != nil {
		panic(err)
	}
}

func (i *Injector) bindType(ty reflect.Type, creator Creator[any]) error {
	if _, ok := i.types[ty]; ok {
		return errors.Newf(nil, errors.ErrorCodeUnknown, "Binding already exists for %s", ty)
	}
	i.types[ty] = creator
	return nil
}

func (i *Injector) bindTypeUnsafe(ty reflect.Type, creator Creator[any]) {
	if err := i.bindType(ty, creator); err != nil {
		panic(err)
	}
}

func (i *Injector) Install(module Module) error {
	return module(i)
}
