package exec

import (
	"context"
	"os/exec"
)

// Run is used to call program wait it until exit, get the output and the exit code.
func Run(name string, arg ...string) (string, int, error) {
	cmd := exec.Command(name, arg...) // #nosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), 1, err
	}
	return string(output), cmd.ProcessState.ExitCode(), nil
}

// RunContext is used to call program, context is used to kill the program.
func RunContext(ctx context.Context, name string, args ...string) (string, int, error) {
	cmd := exec.CommandContext(ctx, name, args...) // #nosec
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), 1, err
	}
	return string(output), cmd.ProcessState.ExitCode(), nil
}

// Command returns the Cmd struct to execute the named program with the given arguments.
func Command(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...) // #nosec
}

// CommandContext is like Command but includes a context.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...) // #nosec
}

// RunCommand is used to run exec.Cmd and get exit code.
func RunCommand(cmd *exec.Cmd) (int, error) {
	err := cmd.Run()
	if err != nil {
		return 1, err
	}
	return cmd.ProcessState.ExitCode(), nil
}
