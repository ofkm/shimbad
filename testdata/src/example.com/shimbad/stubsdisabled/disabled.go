package stubsdisabled

import "errors"

func Empty() {}

func Constant(value string) int {
	return 0
}

func Panic() {
	panic("TODO")
}

func Placeholder() error {
	return errors.New("not implemented")
}
