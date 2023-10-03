package inject

import (
	"strings"

	"github.com/pasataleo/go-errors/errors"
)

// Examples:
//   --boolean
//   --no-boolean
//   --boolean=true
//   --boolean=false
//   --boolean=1
//   --boolean=0
//   --string=value
//   --integer=123
//	 --float=123.456
//   --float=123.456e-78
//   --list=1 --list=2 --list=3
//   --list=1,2,3

type parser[T any] interface {
	Parse(string) error
	Value() T
}

// Parse parses the arguments that are supported by the injector and returns the
// remaining arguments.
func (i *Injector) Parse(args []string) ([]string, error) {
	var remaining []string

	values := make(map[string][]string)
	addValue := func(name, value string) {
		if _, ok := values[name]; !ok {
			values[name] = []string{}
		}
		values[name] = append(values[name], value)
	}

	skipNextArg := false
	for ix, arg := range args {
		if skipNextArg {
			skipNextArg = false
			continue
		}

		name, ok := isFlagName(arg)
		if !ok {
			remaining = append(remaining, arg)
			continue
		}

		var value string
		name, value, ok = isFlagValue(name)
		if !ok {
			// Then the value is in next argument.
			if ix+1 > len(args) {
				// Then we don't have a next argument.
				value = ""
			} else {

				nextArg := args[ix+1]
				if _, ok := isFlagName(nextArg); ok {
					// Then the next argument is a flag name so we can't use it
					// as a value.
					value = ""
				} else {
					skipNextArg = true
					value = nextArg
				}
			}
		}

		if _, ok := i.parsers[name]; !ok {
			// Then the parser is not registered.
			remaining = append(remaining, arg)
			if skipNextArg {
				remaining = append(remaining, value)
			}
			continue
		}

		addValue(name, value)
	}

	for name, values := range values {
		parser := i.parsers[name]
		value, err := parser(values)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse flag %s", name)
		}
		// Write the parsed values into the static bindings of the injector.
		i.staticBindings[name] = value
	}

	return remaining, nil
}

func isFlagName(arg string) (string, bool) {
	if name, ok := strings.CutPrefix(arg, "--"); ok {
		return name, true
	}

	if name, ok := strings.CutPrefix(arg, "-"); ok {
		return name, true
	}

	return arg, false
}

func isFlagValue(arg string) (string, string, bool) {
	if name, value, ok := strings.Cut(arg, "="); ok {
		return name, value, true
	}
	return arg, "", false
}
