package shimsexcluded

import "strings"

func Excluded(value string) string {
	return strings.TrimSpace(value)
}

func Reported(value string) string { // want "avoid a trivial forwarding function"
	return strings.TrimSpace(value)
}
