package inject

func BindValue[T any](value T) Context {
	return &staticContext[T]{
		value: value,
	}
}

func BindFn[T any](creator Creator[T]) Context {
	return &functionContext[T]{
		creator: creator,
	}
}
