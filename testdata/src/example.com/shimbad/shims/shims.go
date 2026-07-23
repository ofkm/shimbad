package shims

import "strings"

const separator = ","

type label string

func local(value string) string {
	return value
}

func Forward(value string) string { // want "avoid a trivial forwarding function"
	return local(value)
}

func External(value string) string { // want "avoid a trivial forwarding function"
	return strings.TrimSpace(value)
}

func Composed(value string) string { // want "avoid a trivial forwarding function"
	return strings.ToLower(strings.TrimSpace(value))
}

func WithDefault(values []string) string { // want "avoid a trivial forwarding function"
	return strings.Join(values, separator)
}

func Emit(value string) { // want "avoid a trivial forwarding function"
	println(value)
}

func Pointer(value string) *string { // want "avoid a trivial forwarding function"
	return new(value)
}

func Convert(value label) string { // want "avoid a trivial forwarding function"
	return string(value)
}

func identity[T any](value T) T {
	return value
}

func Generic(value string) string { // want "avoid a trivial forwarding function"
	return identity[string](value)
}

func Variadic(values ...string) string { // want "avoid a trivial forwarding function"
	return strings.Join(values, separator)
}

func Validated(value string) string {
	if value == "" {
		return ""
	}
	return local(value)
}

type formatter struct{}

func (formatter) Trim(value string) string {
	return strings.TrimSpace(value)
}

func MethodCall(value string) string {
	return formatter{}.Trim(value)
}

var transform = strings.TrimSpace

func FunctionValue(value string) string {
	return transform(value)
}

func Duplicate(value string) string {
	return strings.Join([]string{value, value}, separator)
}

func Ignore(value, fallback string) string {
	return local(value)
}

func NoParameters() string {
	return strings.TrimSpace(" value ")
}

func Recursive(value string) string {
	return Recursive(value)
}

type record struct {
	value string
}

func Field(value record) string {
	return local(value.value)
}

func Underscore(_ string) string {
	return local("")
}
