// +build go1.10, !go1.12

package os

// ExitCode returns the exit code of the exited process, or -1
// if the process hasn't exited or was terminated by a signal.
//
// From go1.12
func (p *ProcessState) ExitCode() int {
	// return -1 if the process hasn't started.
	if p == nil {
		return -1
	}
	return p.status.ExitStatus()
}

