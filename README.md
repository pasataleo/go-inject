# go-inject

This my dependency injection library.

## Usage

```go
package main

import (
	"fmt"

	"github.com/pasataleo/go-inject/inject"
)

func main() {
	injector := inject.NewInjector()
	inject.BindFn[string](func(_ *inject.Injector) (string, error) {
		return "Hello World", nil
	}).ToUnsafe(injector)

	var value string
	if err := injector.Inject(&value); err != nil {
		panic(err) // something went wrong (e.g. value is not a pointer)
	}

	fmt.Println(value) // Hello World
}
```

### Bindings

You can bind creator functions and values directly to both string keywords and types.

- `BindFn[Type](func(*inject.Injector) (Type, error)).To(injector)` - Bind a creator function to a type.
- `BindFn[Type](func(*inject.Injector) (Type, error)).To(injector, "tag")` - Bind a creator function to a keyword.
- `BindValue[Type](value Type).To(injector)` - Bind a specific value to a type.
- `BindValue[Type](value Type).To(injector, "tag")` - Bind a specific value to a keyword.

### Injecting structs

You can use the `inject` tag on fields within structs to inject values into them.

```go
package main


import (
	"fmt"

	"github.com/pasataleo/go-inject/inject"
)

type MyStruct struct {
	// StringValue value will be injected by keyword.
    StringValue string `inject:"value"` 
		
	// BoolValue will be injected by type.	
	BoolValue bool
		
	// Skip will not be injected.	
	Skip string `inject:"-"`
}

func main() {
    injector := inject.NewInjector()
    inject.BindFn[string](func(_ *inject.Injector) (string, error) {
        return "Hello World", nil
    }).ToUnsafe(injector, "value") // Anything tagged as "value" will be injected with "Hello World".
    inject.BindValue[bool](true).ToUnsafe(injector) // All bound booleans will return true.

    var value MyStruct
    if err := injector.Inject(&value); err != nil {
        panic(err) // something went wrong (e.g. value is not a pointer)
    }

    fmt.Println(value.StringValue) // Hello World
    fmt.Println(value.BoolValue) // true
    fmt.Println(value.Skip) // ""
}
```

### Modules

You can use the `Install(module Module)` method to install a module into the injector. This allows you to group bindings
together and reuse them in multiple injectors.
