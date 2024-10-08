package inject

import (
	"reflect"

	"github.com/pasataleo/go-errors/errors"
)

type Creator[T any] func(injector *Injector) (T, error)

type Injector struct {
	// staticBindings and staticTypes are used to store bindings and types for static values.
	staticBindings map[string]interface{}
	staticTypes    map[reflect.Type]interface{}

	// functionBindings and functionTypes are used to store bindings and types for functions.
	functionBindings map[string]Creator[any]
	functionTypes    map[reflect.Type]Creator[any]
}

func NewInjector() *Injector {
	return &Injector{
		staticTypes:      make(map[reflect.Type]interface{}),
		staticBindings:   make(map[string]interface{}),
		functionBindings: make(map[string]Creator[any]),
		functionTypes:    make(map[reflect.Type]Creator[any]),
	}
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

				tag := parseTags(inject)

				if i.Has(tag.key) {
					value, err := i.Get(tag.key)
					if err != nil {
						return errors.Wrap(err, "Failed to create value")
					}
					fieldValue.Set(reflect.ValueOf(value))
					continue
				}

				if tag.optional {
					// Then it's fine, we just won't set a value for the field.
					continue
				} else {
					return errors.Newf(nil, errors.ErrorCodeUnknown, "Missing binding for %s", tag.key)
				}
			}

			// Second, if the field doesn't have a tag, we can try and inject it by type
			if i.HasType(fieldType.Type) {
				value, err := i.GetType(fieldType.Type)
				if err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				fieldValue.Set(reflect.ValueOf(value))
				continue
			} else if fieldType.Type.Kind() == reflect.Ptr && i.HasType(fieldType.Type.Elem()) {
				// If the field is a pointer, we can try and inject it by the underlying type.
				value, err := i.GetType(fieldType.Type.Elem())
				if err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				fieldValue.Set(reflect.ValueOf(value))
				continue
			}

			// Finally, if all else has failed and the type is a structure or a pointer to a structure, we can try and
			// inject the fields into the structure recursively.
			if fieldType.Type.Kind() == reflect.Struct {
				if err := i.Inject(fieldValue.Addr().Interface()); err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				continue
			} else if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
				value := reflect.New(fieldType.Type.Elem())
				if err := i.Inject(value.Interface()); err != nil {
					return errors.Wrapf(err, "Failed to inject field %s", fieldType.Name)
				}
				fieldValue.Set(value)
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
