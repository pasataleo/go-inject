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

func createBool(_ *Injector) (bool, error) {
	return true, nil
}

func createInt(value int) Creator[int] {
	return func(injector *Injector) (int, error) {
		return value, nil
	}
}

func TestInjector_Inject(t *testing.T) {
	injector := NewInjector()

	BindFn(createBool).ToUnsafe(injector)
	BindFn(createInt(42)).ToUnsafe(injector)
	BindValue("string").ToUnsafe(injector)

	BindValue(false).ToUnsafe(injector, "boolean")
	BindFn(createInt(33)).ToUnsafe(injector, "integer")

	var value TestStruct
	tests.ExecFn(t, injector.Inject, &value).NoError()
	tests.Value(t, value.Bool).True()
	tests.Value(t, value.Int).Equals(42)
	tests.Value(t, value.String).Equals("string")
	tests.Value(t, value.Struct.Bool).False()
	tests.Value(t, value.Struct.Int).Equals(33)
	tests.Value(t, value.Struct.String).Empty()
	tests.Value(t, value.StructPtr.Bool).False()
	tests.Value(t, value.StructPtr.Int).Equals(33)
	tests.Value(t, value.StructPtr.String).Empty()
}
