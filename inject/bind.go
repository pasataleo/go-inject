package inject

import "strconv"

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

func BindFlag[T any](parser Parser[T]) Context {
	return &parserContext[T]{
		parser: parser,
	}
}

func BindStringFlag() Context {
	return BindFlag(func(values []string) (string, error) {
		// If we have duplicate flags, then we use the last one.
		return values[len(values)-1], nil
	})
}

func BindIntegerFlag() Context {
	return BindFlag(func(values []string) (int, error) {
		// If we have duplicate flags, then we use the last one.
		return strconv.Atoi(values[len(values)-1])
	})
}

func BindFloatFlag() Context {
	return BindFlag(func(values []string) (float64, error) {
		// If we have duplicate flags, then we use the last one.
		return strconv.ParseFloat(values[len(values)-1], 64)
	})
}

func BindBoolFlag() Context {
	return BindFlag(func(values []string) (bool, error) {
		arg := values[len(values)-1]
		if len(arg) == 0 {
			// If we have no value for the flag, then we assume it's true.
			return true, nil
		}
		// If we have duplicate flags, then we use the last one.
		return strconv.ParseBool(values[len(values)-1])
	})
}
