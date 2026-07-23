package stubscustomplaceholder

func Pending() { // want "placeholder panic"
	panic("pending implementation")
}

func DefaultMarkerIsOverridden() {
	panic("TODO")
}
