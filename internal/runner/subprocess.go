package runner

import (
	"bytes"
	"context"
	"os/exec"
)

type SubprocessExecutor struct {
	Command string
	Args    []string
}

func (e SubprocessExecutor) Execute(ctx context.Context, inputPath string, outputDir string) ExecutionResult {
	args := append([]string{}, e.Args...)
	args = append(args, inputPath, outputDir)
	cmd := exec.CommandContext(ctx, e.Command, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitCode = 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}
	if err != nil && stderr.Len() > 0 {
		err = redactExecError{message: stderr.String()}
	}
	return ExecutionResult{Stdout: bytes.NewReader(stdout.Bytes()), ExitCode: exitCode, Err: err}
}

type redactExecError struct{ message string }

func (e redactExecError) Error() string { return redact(e.message) }
