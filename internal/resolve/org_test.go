package resolve

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoteOrg(t *testing.T) {
	cases := map[string]string{
		"https://github.com/chkp-acme/repo.git": "chkp-acme",
		"https://github.com/chkp-acme/repo":     "chkp-acme",
		"git@github.com:chkp-acme/repo.git":     "chkp-acme",
		"ssh://git@github.com/chkp-acme/repo":   "chkp-acme",
		"ssh://git@gh1:2222/chkp-acme/repo":     "chkp-acme",
		"https://github.com/repo.git":           "",
		"git@github.com:repo.git":               "",
		"/srv/git/org/repo.git":                 "",
		"./relative/path:with-colon":            "",
		"":                                      "",
	}
	for remote, want := range cases {
		if got := remoteOrg(remote); got != want {
			t.Errorf("remoteOrg(%q) = %q, want %q", remote, got, want)
		}
	}
}

func TestOrgWildcardPicksProfile(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Orgs = map[string]string{"chkp-*": "work"}
	remote := func(string) string { return "git@github.com:chkp-acme/repo.git" }
	got := Profile(cfg, filepath.Join(root, "personal/blog"), "", remote)
	if got.ProfileName != "work" || got.Source != orgSource+"chkp-acme" {
		t.Fatalf("unexpected match %+v", got)
	}
}

func TestOrgMatchIsCaseInsensitive(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Orgs = map[string]string{"CHKP-*": "work"}
	remote := func(string) string { return "https://github.com/Chkp-Acme/repo.git" }
	if got := Profile(cfg, filepath.Join(root, "personal/blog"), "", remote); got.ProfileName != "work" {
		t.Fatalf("got %q, want work", got.ProfileName)
	}
}

func TestOrgExactBeatsWildcard(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Orgs = map[string]string{"chkp-*": "work", "chkp-labs": "personal"}
	remote := func(string) string { return "https://github.com/chkp-labs/repo.git" }
	if got := Profile(cfg, filepath.Join(root, "work/api"), "", remote); got.ProfileName != "personal" {
		t.Fatalf("got %q, want personal", got.ProfileName)
	}
}

func TestOrgLongerWildcardBeatsShorter(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Orgs = map[string]string{"chkp-*": "personal", "chkp-x*": "work"}
	remote := func(string) string { return "https://github.com/chkp-xyz/repo.git" }
	if got := Profile(cfg, filepath.Join(root, "personal/blog"), "", remote); got.ProfileName != "work" {
		t.Fatalf("got %q, want work", got.ProfileName)
	}
}

func TestOrgUnmatchedFallsBackToDirs(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Orgs = map[string]string{"chkp-*": "personal"}
	remote := func(string) string { return "https://github.com/other-org/repo.git" }
	if got := Profile(cfg, filepath.Join(root, "work/api"), "", remote); got.ProfileName != "work" {
		t.Fatalf("got %q, want work", got.ProfileName)
	}
}

func TestOrgBeatsDirs(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Orgs = map[string]string{"chkp-*": "personal"}
	remote := func(string) string { return "https://github.com/chkp-acme/repo.git" }
	if got := Profile(cfg, filepath.Join(root, "work/api"), "", remote); got.ProfileName != "personal" {
		t.Fatalf("got %q, want personal", got.ProfileName)
	}
}

func TestSSHHostBeatsOrg(t *testing.T) {
	root, cfg := newTree(t)
	cfg.SSHHosts = map[string]string{"gh-personal": "personal"}
	cfg.Orgs = map[string]string{"chkp-*": "work"}
	remote := func(string) string { return "git@gh-personal:chkp-acme/repo.git" }
	if got := Profile(cfg, filepath.Join(root, "personal/blog"), "", remote); got.ProfileName != "personal" {
		t.Fatalf("got %q, want personal", got.ProfileName)
	}
}

func TestRepoFileBeatsOrg(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Orgs = map[string]string{"chkp-*": "work"}
	if err := os.WriteFile(filepath.Join(root, "personal", RepoFileName), []byte("personal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	remote := func(string) string { return "https://github.com/chkp-acme/repo.git" }
	if got := Profile(cfg, filepath.Join(root, "personal"), "", remote); got.ProfileName != "personal" {
		t.Fatalf("got %q, want personal", got.ProfileName)
	}
}

func TestOrgSkipsGitWhenUnconfigured(t *testing.T) {
	root, cfg := newTree(t)
	remote := func(string) string {
		t.Fatal("remote lookup ran without ssh_hosts, remote_user or orgs")
		return ""
	}
	Profile(cfg, filepath.Join(root, "work"), "", remote)
}
