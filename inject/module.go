package inject

type Module interface {
	Install(injector *Injector) error
}

var _ Module = ModuleFn(nil)

type ModuleFn func(injector *Injector) error

func (f ModuleFn) Install(injector *Injector) error {
	return f(injector)
}

func (i *Injector) Install(module Module) error {
	return module.Install(i)
}
