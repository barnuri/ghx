package resolve

import (
	"net/url"
	"os/exec"
	"strings"
)

const originRemote = "origin"

// RemoteURL returns the URL of the git remote for dir, or "" when there is none.
type RemoteURL func(dir string) string

// GitOriginURL asks git, so url.<base>.insteadOf rewrites apply.
func GitOriginURL(dir string) string {
	output, err := exec.Command("git", "-C", dir, "remote", "get-url", originRemote).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// remoteHost returns the host of an ssh:// URL or an scp-like "user@host:path" remote.
func remoteHost(remote string) string {
	if strings.Contains(remote, "://") {
		parsed, err := url.Parse(remote)
		if err != nil {
			return ""
		}
		return parsed.Hostname()
	}
	hostPart, _, found := strings.Cut(remote, ":")
	if !found || strings.ContainsAny(hostPart, `/\`) {
		return ""
	}
	return hostPart[strings.LastIndex(hostPart, "@")+1:]
}

// remoteUser returns the user and host of an http(s) remote such as "https://me@github.com/org/repo".
func remoteUser(remote string) (string, string) {
	parsed, err := url.Parse(remote)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User == nil {
		return "", ""
	}
	return parsed.User.Username(), parsed.Hostname()
}
