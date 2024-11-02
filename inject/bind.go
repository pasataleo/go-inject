package inject

// BindValue returns a Context that contains an already computed value.
func BindValue[T any](value T) Context {
	return &staticContext[T]{
		value: value,
	}
}

// BindFn returns a Context that contains a function that will create a value.
func BindFn[T any](creator Creator[T]) Context {
	return &functionContext[T]{
		creator: creator,
	}
}

// Binder returns a function that will bind a value to the injector when it is called.
func Binder[T any](injector *Injector, args ...string) func(string, T) error {
	return func(_ string, value T) error {
		return BindValue(value).ToSafe(injector, args...)
	}
}

// DirectBinder returns a function that will bind a value to the injector when it is called.
//
// DirectBinder accepts the name of the binding from the later caller, instead of the current caller.
func DirectBinder[T any](injector *Injector) func(string, T) error {
	return func(name string, value T) error {
		return BindValue(value).ToSafe(injector, name)
	}
}
