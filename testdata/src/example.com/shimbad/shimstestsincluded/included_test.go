package shimstestsincluded

func TrimForTest(value string) string { // want "avoid a trivial forwarding function"
	return trim(value)
}
