package main

var builtins = map[string]BuiltinFunc{}

func register(name string, fn BuiltinFunc) {
	builtins[name] = fn
}

func registerAlias(alias, target string) {
	fn, ok := builtins[target]
	if !ok {
		panic("registerAlias: target not found: " + target)
	}
	builtins[alias] = fn
}
