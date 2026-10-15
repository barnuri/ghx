//go:build windows

package platform

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const (
	createNewProcessGroup = 0x00000200
	detachedProcess       = 0x00000008
)

// Run starts gh as a child and mirrors its exit code. Ctrl+C reaches gh through the shared
// console, so ghx only ignores it to stay alive until gh exits.
func Run(binary string, args []string, env []string) (int, error) {
	signal.Ignore(os.Interrupt)
	command := exec.Command(binary, args...)
	command.Env = env
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := command.Run()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), nil
	}
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func StartDetached(binary string, args ...string) error {
	command := exec.Command(binary, args...)
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewProcessGroup | detachedProcess}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
