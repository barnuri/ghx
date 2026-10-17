package resolve

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/barnuri/ghx/internal/config"
)

const (
	RepoFileName     = ".ghx"
	overrideSource   = "GHX_PROFILE"
	defaultSource    = "default"
	sshHostSource    = "ssh host "
	remoteUserSource = "remote user "
)

// Profile picks the profile for cwd in this order: GHX_PROFILE, the nearest .ghx file,
// the ssh host or https user of the origin remote, the deepest configured dir containing cwd, then default.
// No match means plain gh.
func Profile(cfg config.Config, cwd string, override string, remoteURL RemoteURL) Match {
	if override != "" {
		return named(cfg, override, overrideSource)
	}
	current := canonical(cwd)
	if name, file := nearestRepoFile(current); name != "" {
		return named(cfg, name, file)
	}
	if match := fromOrigin(cfg, current, remoteURL); match.Found() {
		return match
	}
	if match := deepestDir(cfg, current); match.Found() {
		return match
	}
	if cfg.Default != "" {
		return named(cfg, cfg.Default, defaultSource)
	}
	return Match{}
}

// named treats a name with no matching profile as a gh user on github.com, so a .ghx file
// works without any global config.
func named(cfg config.Config, name string, source string) Match {
	profile, ok := cfg.Profiles[name]
	if !ok {
		profile = config.Profile{User: name}
	}
	return Match{ProfileName: name, Profile: profile, Source: source}
}

// fromOrigin runs git only when ssh_hosts or remote_user is configured, so other setups never pay for it.
func fromOrigin(cfg config.Config, current string, remoteURL RemoteURL) Match {
	if len(cfg.SSHHosts) == 0 && !cfg.RemoteUser {
		return Match{}
	}
	remote := remoteURL(current)
	if match := sshHost(cfg, remote); match.Found() {
		return match
	}
	if cfg.RemoteUser {
		return userInRemote(cfg, remote)
	}
	return Match{}
}

func sshHost(cfg config.Config, remote string) Match {
	host := remoteHost(remote)
	profileName, ok := cfg.SSHHosts[host]
	if !ok {
		return Match{}
	}
	return named(cfg, profileName, sshHostSource+host)
}

// userInRemote prefers the profile with that login and host, else uses the login directly.
func userInRemote(cfg config.Config, remote string) Match {
	user, host := remoteUser(remote)
	if user == "" {
		return Match{}
	}
	source := remoteUserSource + user + "@" + host
	for name, profile := range cfg.Profiles {
		if profile.User == user && profile.HostOrDefault() == host {
			return Match{ProfileName: name, Profile: profile, Source: source}
		}
	}
	profile := config.Profile{User: user}
	if host != config.DefaultHost {
		profile.Host = host
	}
	return Match{ProfileName: user, Profile: profile, Source: source}
}

func deepestDir(cfg config.Config, current string) Match {
	best := Match{}
	bestLength := -1
	for dir, profileName := range cfg.Dirs {
		base := canonical(config.ExpandHome(dir))
		if !contains(base, current) || len(base) <= bestLength {
			continue
		}
		best = named(cfg, profileName, base)
		bestLength = len(base)
	}
	return best
}

func nearestRepoFile(start string) (string, string) {
	for dir := start; ; dir = filepath.Dir(dir) {
		file := filepath.Join(dir, RepoFileName)
		if name := readRepoFile(file); name != "" {
			return name, file
		}
		if filepath.Dir(dir) == dir {
			return "", ""
		}
	}
}

func readRepoFile(file string) string {
	content, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			return line
		}
	}
	return ""
}

func canonical(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		absolute = path
	}
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		return resolved
	}
	return filepath.Clean(absolute)
}

func contains(base string, target string) bool {
	relative, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)))
}
