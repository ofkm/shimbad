package shimsdeclarations

import "strings"

type formatter struct{}

func (formatter) Trim(value string) string { // want "avoid a trivial forwarding function"
	return strings.TrimSpace(value)
}

var Trim = func(value string) string { // want "avoid a trivial forwarding function"
	return strings.TrimSpace(value)
}
