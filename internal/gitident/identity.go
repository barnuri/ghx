package gitident

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/barnuri/ghx/internal/config"
)

const (
	nameKey       = "user.name"
	emailKey      = "user.email"
	unsetExitCode = 1
)

// Apply writes identity to the repo-local git config of dir. Mode "once" fills only unset keys,
// "always" also replaces a different value. Outside a git repo it does nothing.
func Apply(identity config.Identity, dir string) error {
	if identity.IsEmpty() || !insideRepo(dir) {
		return nil
	}
	if err := applyKey(dir, nameKey, identity.Name, identity.Mode); err != nil {
		return err
	}
	return applyKey(dir, emailKey, identity.Email, identity.Mode)
}

func applyKey(dir string, key string, value string, mode string) error {
	if value == "" {
		return nil
	}
	current, isSet, err := localValue(dir, key)
	if err != nil {
		return err
	}
	if isSet && (mode == config.IdentityOnce || current == value) {
		return nil
	}
	if output, err := exec.Command("git", "-C", dir, "config", "--local", key, value).CombinedOutput(); err != nil {
		return fmt.Errorf("set git %s: %w: %s", key, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func insideRepo(dir string) bool {
	return exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Run() == nil
}

func localValue(dir string, key string) (string, bool, error) {
	output, err := exec.Command("git", "-C", dir, "config", "--local", "--get", key).Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == unsetExitCode {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read git %s: %w", key, err)
	}
	return strings.TrimSpace(string(output)), true, nil
}
