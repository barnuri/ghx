package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseValid(t *testing.T) {
	cfg, err := Parse([]byte(`
manage_gh: true
default: personal
profiles:
  work: {user: me-work, host: github.example.com}
  personal: {user: me}
dirs:
  ~/code/work: work
ssh_hosts:
  gh-work: work
`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.ManageGh || cfg.Profiles["work"].HostOrDefault() != "github.example.com" || cfg.Profiles["personal"].HostOrDefault() != DefaultHost || cfg.SSHHosts["gh-work"] != "work" {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestParseRejectsInvalid(t *testing.T) {
	cases := map[string]string{
		"missing user":             "profiles: {work: {host: x}}",
		"unknown dir profile":      "profiles: {work: {user: a}}\ndirs: {~/x: nope}",
		"unknown ssh host profile": "profiles: {work: {user: a}}\nssh_hosts: {gh1: nope}",
		"unknown default profile":  "profiles: {work: {user: a}}\ndefault: nope",
		"bad profile git_identity": "profiles: {work: {user: a, git_identity: never}}",
		"bad global git_identity":  "git_identity: never",
		"bad yaml":                 "profiles: [",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(content)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestIdentityForFallsBackToTopLevel(t *testing.T) {
	cfg, err := Parse([]byte(`
name: Global Name
email: global@x.com
git_identity: always
profiles:
  work: {user: a, email: work@x.com, git_identity: once}
  personal: {user: b}
`))
	if err != nil {
		t.Fatal(err)
	}
	work := cfg.IdentityFor(cfg.Profiles["work"])
	if work != (Identity{Name: "Global Name", Email: "work@x.com", Mode: IdentityOnce}) {
		t.Fatalf("work: got %+v", work)
	}
	personal := cfg.IdentityFor(cfg.Profiles["personal"])
	if personal != (Identity{Name: "Global Name", Email: "global@x.com", Mode: IdentityAlways}) {
		t.Fatalf("personal: got %+v", personal)
	}
	if got := (Config{}).IdentityFor(Profile{User: "a"}); !got.IsEmpty() || got.Mode != IdentityOnce {
		t.Fatalf("empty: got %+v", got)
	}
}

func TestLoadMissingFileIsEmpty(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "absent.yaml"))
	if err != nil || len(cfg.Profiles) != 0 {
		t.Fatalf("got %+v, %v", cfg, err)
	}
}

func TestDefaultPathHonorsEnv(t *testing.T) {
	t.Setenv("GHX_CONFIG", "/custom/ghx.yaml")
	if got := DefaultPath(); got != "/custom/ghx.yaml" {
		t.Fatalf("got %s", got)
	}
}

func TestExpandHome(t *testing.T) {
	t.Setenv("HOME", "/home/me")
	t.Setenv("USERPROFILE", "/home/me")
	if got := ExpandHome("~/code"); !strings.HasSuffix(filepath.ToSlash(got), "/home/me/code") {
		t.Fatalf("got %s", got)
	}
	if got := ExpandHome("/abs/~/x"); got != "/abs/~/x" {
		t.Fatalf("got %s", got)
	}
}

func TestAbbreviateHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if got := AbbreviateHome(filepath.Join(home, "code", ".ghx")); got != filepath.Join("~", "code", ".ghx") {
		t.Fatalf("got %s", got)
	}
	if got := AbbreviateHome(home + "other"); got != home+"other" {
		t.Fatalf("got %s", got)
	}
}
