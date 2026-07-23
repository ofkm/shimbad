package stubs

import (
	"errors"
	"fmt"
)

func Empty() {} // want "function has no implementation"

func EmptyWithComment() { // want "function has no implementation"
	// Intentionally left blank is still a no-op implementation.
}

func DiscardOnly(value string) { // want "function has no implementation"
	_ = value
}

func Constant(value string) int { // want "function ignores its inputs"
	return 0
}

func StaticValues(value string) (string, []string, error) { // want "function ignores its inputs"
	_ = value
	return "", []string{}, nil
}

func BareReturn(value string) (result int) { // want "function ignores its inputs"
	return
}

func PanicTODO() { // want "placeholder panic"
	panic("TODO")
}

func PanicConcatenated() { // want "placeholder panic"
	panic("not " + "implemented")
}

func PanicFormatted(operation string) { // want "placeholder panic"
	panic(fmt.Errorf("NYI: %s", operation))
}

func PlaceholderError() error { // want "placeholder not-implemented result"
	return errors.New("not implemented")
}

func PlaceholderPair(name string) (string, error) { // want "placeholder not-implemented result"
	return "", fmt.Errorf("TODO: support %s", name)
}

func NoArgConstant() int {
	return 1
}

func Identity(value string) string {
	return value
}

func Validated(value string) int {
	if value == "" {
		return 0
	}
	return 1
}

func GeneralPanic(value string) {
	panic("invalid value: " + value)
}

func Sentinel(value string) error {
	return errors.New("invalid value")
}

func Computed(value string) int {
	return len(value)
}

type handler struct{}

func (handler) Empty() {}

var emptyCallback = func() {}
