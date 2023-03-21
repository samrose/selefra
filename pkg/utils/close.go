package utils

var toClose map[string]func() = make(map[string]func())

func RegisterClose(name string, close func()) {
	toClose[name] = close
}

func MultiRegisterClose(m map[string]func()) {
	for name, fn := range m {
		toClose[name] = fn
	}
}

func Close() {
	for _, cleanFn := range toClose {
		cleanFn()
	}
}
