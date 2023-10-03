package inject

import (
	"os"
	"strings"

	"github.com/pasataleo/go-errors/errors"
)

// IncludeEnvironmentVariables parses all environment variables and adds them
// to the injector if we have a binding for the environment variable name.
func (i *Injector) IncludeEnvironmentVariables() error {
	for _, arg := range os.Environ() {
		name, value, ok := strings.Cut(arg, "=")
		if !ok {
			return errors.Newf(nil, errors.ErrorCodeUnknown, "Invalid environment variable: %s", arg)
		}

		parser, ok := i.parsers[name]
		if !ok {
			continue
		}

		parsed, err := parser([]string{value})
		if err != nil {
			return errors.Wrap(err, "Failed to parse environment variable")
		}

		i.staticBindings[name] = parsed
	}
	return nil
}
