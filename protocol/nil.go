package protocol

type NilInterop struct{}

func (i NilInterop) Interop(_ []byte) (string, bool) {
	return "", false
}

func (i NilInterop) Protocol() string {
	return "unknown"
}
