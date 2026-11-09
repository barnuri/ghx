package gitident

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/barnuri/ghx/internal/config"
)

const (
	nameKey       = "user.name"
	emailKey      = "user.email"
	unsetExitCode = 1
)

// Lookup returns the GitHub name and email used for keys the config leaves empty.
type Lookup func() (config.Identity, error)

// Apply writes identity to the repo-local git config of dir. Mode "once" fills only unset keys,
// "always" also replaces a different value. Outside a git repo it does nothing.
func Apply(identity config.Identity, dir string, lookup Lookup) error {
	if !identity.Enabled() || !insideRepo(dir) {
		return nil
	}
	fromGithub := sync.OnceValues(lookup)
	valueOf := func(explicit string, pick func(config.Identity) string) func() (string, error) {
		return func() (string, error) {
			if explicit != "" || !identity.FromGithub {
				return explicit, nil
			}
			fetched, err := fromGithub()
			return pick(fetched), err
		}
	}
	nameValue := valueOf(identity.Name, func(fetched config.Identity) string { return fetched.Name })
	if err := applyKey(dir, nameKey, identity.Mode, nameValue); err != nil {
		return err
	}
	emailValue := valueOf(identity.Email, func(fetched config.Identity) string { return fetched.Email })
	return applyKey(dir, emailKey, identity.Mode, emailValue)
}

// applyKey reads the wanted value only when it may be written, so a GitHub lookup runs only then.
func applyKey(dir string, key string, mode string, wantedValue func() (string, error)) error {
	current, isSet, err := localValue(dir, key)
	if err != nil {
		return err
	}
	if isSet && mode == config.IdentityOnce {
		return nil
	}
	wanted, err := wantedValue()
	if err != nil {
		return err
	}
	if wanted == "" || (isSet && current == wanted) {
		return nil
	}
	if output, err := exec.Command("git", "-C", dir, "config", "--local", key, wanted).CombinedOutput(); err != nil {
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
