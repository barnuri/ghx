//go:build !windows

package platform

import (
	"os/exec"
	"syscall"
)

// Run replaces the ghx process with gh, so signals, TTY and exit code belong to gh directly.
func Run(binary string, args []string, env []string) (int, error) {
	argv := append([]string{binary}, args...)
	return 0, syscall.Exec(binary, argv, env)
}

func StartDetached(binary string, args ...string) error {
	command := exec.Command(binary, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
