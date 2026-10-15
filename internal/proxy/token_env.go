package proxy

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/barnuri/ghx/internal/config"
)

const (
	tokenEnv           = "GH_TOKEN"
	enterpriseTokenEnv = "GH_ENTERPRISE_TOKEN"
	hostEnv            = "GH_HOST"
)

// NeedsToken is false for gh auth commands: they manage stored logins and refuse to run
// while GH_TOKEN is set. git-credential is the exception because git uses it to push.
func NeedsToken(args []string) bool {
	if len(args) == 0 || args[0] != "auth" {
		return true
	}
	return len(args) > 1 && args[1] == "git-credential"
}

func HasExplicitToken(env []string) bool {
	return lookup(env, tokenEnv) != "" || lookup(env, enterpriseTokenEnv) != ""
}

func FetchToken(ghPath string, profile config.Profile) (string, error) {
	var stderr bytes.Buffer
	host := profile.HostOrDefault()
	command := exec.Command(ghPath, "auth", "token", "--hostname", host, "--user", profile.User)
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("gh has no token for user %s on %s: %s. Log in with: gh auth login --hostname %s",
			profile.User, host, strings.TrimSpace(stderr.String()), host)
	}
	return strings.TrimSpace(string(output)), nil
}

func WithProfile(env []string, profile config.Profile, token string) []string {
	host := profile.HostOrDefault()
	if host == config.DefaultHost {
		return withVar(env, tokenEnv, token)
	}
	return withVar(withVar(env, enterpriseTokenEnv, token), hostEnv, host)
}

// withVar drops existing entries first: execve keeps duplicates and gh reads the first one.
func withVar(env []string, key string, value string) []string {
	prefix := key + "="
	result := make([]string, 0, len(env)+1)
	for _, entry := range env {
		if !strings.HasPrefix(entry, prefix) {
			result = append(result, entry)
		}
	}
	return append(result, prefix+value)
}

func lookup(env []string, key string) string {
	prefix := key + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			return entry[len(prefix):]
		}
	}
	return ""
}
