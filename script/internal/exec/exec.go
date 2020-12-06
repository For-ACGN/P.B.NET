package exec

import (
	"context"
	"os/exec"
)

// Run is used to call program wait it until exit, get the output and the exit code.
func Run(name string, arg ...string) (output string, code int, err error) {
	cmd := exec.Command(name, arg...) // #nosec
	out, err := cmd.CombinedOutput()
	output = string(out)
	if err != nil {
		code = 1
		return
	}
	code = cmd.ProcessState.ExitCode()
	return
}

// RunContext is used to call program wait it until exit, get the output and the
// exit code with a context.
func RunContext(ctx context.Context, name string, arg ...string) (output string, code int, err error) {
	cmd := exec.CommandContext(ctx, name, arg...) // #nosec
	out, err := cmd.CombinedOutput()
	output = string(out)
	if err != nil {
		code = 1
		return
	}
	code = cmd.ProcessState.ExitCode()
	return
}
