package stubsdeclarations

type handler struct{}

func (handler) Empty() {} // want "function has no implementation"

var emptyCallback = func() {} // want "function has no implementation"
