package console

// IsTerminal is used to check is in terminal.
func IsTerminal(handle uintptr) bool {
	return isTerminal(handle)
}

// Clear is used to clean console buffer.
func Clear(handle uintptr) error {
	return clear(handle)
}
