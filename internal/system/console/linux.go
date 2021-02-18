// +build !windows

package console

func isTerminal(fd uintptr) bool {
	return false
}

func clear(fd uintptr) error {
	return nil
}
