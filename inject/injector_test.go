package inject

import (
	"testing"

	"github.com/pasataleo/go-testing/tests"
)

type NestedTestStruct struct {
	Bool     bool   `inject:"boolean"`
	Int      int    `inject:"integer"`
	Optional string `inject:"string;optional"`

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

	// First, all un-named boolean fields will be created
	// by the createBool function.
	BindFn(createBool).To(injector)

	// Then, all un-named integer fields will be created
	// by the createInt function.
	BindFn(createInt(42)).To(injector)

	// Finally, all un-named string fields will be assigned
	// the value "string".
	BindValue("string").To(injector)

	// Now, we'll create some specific bindings.

	// First, any field with the tag "boolean" will be set to false.
	BindValue(false).To(injector, "boolean")

	// Then, any field with the tag "integer" will be set to 33.
	BindFn(createInt(33)).To(injector, "integer")

	var value TestStruct
	tests.ExecuteE(injector.Inject(&value)).NoError(t)

	tests.Execute(value.Bool).Equal(t, true)
	tests.Execute(value.Int).Equal(t, 42)
	tests.Execute(value.String).Equal(t, "string")

	// The two nested structs should have the same values.

	tests.Execute(value.Struct.Bool).Equal(t, false)
	tests.Execute(value.Struct.Int).Equal(t, 33)
	tests.Execute(value.Struct.String).Equal(t, "")
	tests.Execute(value.Struct.Optional).Equal(t, "")

	tests.Execute(value.StructPtr.Bool).Equal(t, false)
	tests.Execute(value.StructPtr.Int).Equal(t, 33)
	tests.Execute(value.StructPtr.String).Equal(t, "")
	tests.Execute(value.StructPtr.Optional).Equal(t, "")
}

func TestInjector_Direct(t *testing.T) {
	injector := NewInjector()

	tests.ExecuteE(Binder[bool](injector, "boolean")("irrelevant", true)).NoError(t)
	tests.ExecuteE(DirectBinder[bool](injector)("relevant", true)).NoError(t)

	tests.Execute2E(injector.Get("boolean")).NoError(t).Equal(t, true)
	tests.Execute2E(injector.Get("relevant")).NoError(t).Equal(t, true)
	tests.Execute(injector.Has("irrelevant")).Equal(t, false)
}
