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
`))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.ManageGh || cfg.Profiles["work"].HostOrDefault() != "github.example.com" || cfg.Profiles["personal"].HostOrDefault() != DefaultHost {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestParseRejectsInvalid(t *testing.T) {
	cases := map[string]string{
		"missing user":            "profiles: {work: {host: x}}",
		"unknown dir profile":     "profiles: {work: {user: a}}\ndirs: {~/x: nope}",
		"unknown default profile": "profiles: {work: {user: a}}\ndefault: nope",
		"bad yaml":                "profiles: [",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(content)); err == nil {
				t.Fatal("expected error")
			}
		})
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
