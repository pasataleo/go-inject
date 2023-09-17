package inject

import (
	"testing"

	"github.com/pasataleo/go-testing/tests"
)

type NestedTestStruct struct {
	Bool bool `inject:"boolean"`
	Int  int  `inject:"integer"`

	// This field will not be injected
	String string `inject:"-"`
}

type TestStruct struct {
	Bool   bool
	Int    int
	String string

	Struct    NestedTestStruct
	StructPtr *NestedTestStruct
}

func TestInjector_Inject(t *testing.T) {
	injector := NewInjector()

	BindUnsafe(injector, func(injector *Injector) (bool, error) {
		return true, nil
	})
	BindUnsafe(injector, func(injector *Injector) (int, error) {
		return 42, nil
	})
	BindUnsafe(injector, func(injector *Injector) (string, error) {
		return "string", nil
	})

	injector.BindUnsafe("boolean", func(injector *Injector) (interface{}, error) {
		return false, nil
	})
	injector.BindUnsafe("integer", func(injector *Injector) (interface{}, error) {
		return 33, nil
	})

	tests.ExecFn(t, injector.Bind, "boolean", func(injector *Injector) (interface{}, error) {
		return false, nil
	}).NoError()
	tests.ExecFn(t, injector.Bind, "integer", func(injector *Injector) (interface{}, error) {
		return 33, nil
	}).NoError()

	var value TestStruct
	tests.ExecFn(t, injector.Inject, &value).NoError()
	tests.Equals(t, true, value.Bool)
	tests.Equals(t, 42, value.Int)
	tests.Equals(t, "string", value.String)
	tests.Equals(t, false, value.Struct.Bool)
	tests.Equals(t, 33, value.Struct.Int)
	tests.Equals(t, "", value.Struct.String)
	tests.Equals(t, false, value.StructPtr.Bool)
	tests.Equals(t, 33, value.StructPtr.Int)
	tests.Equals(t, "", value.StructPtr.String)
}
