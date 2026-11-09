package gitident

import (
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/barnuri/ghx/internal/config"
)

func newRepo(t *testing.T, localValues map[string]string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	runGit(t, dir, "init", "-q")
	for key, value := range localValues {
		runGit(t, dir, "config", "--local", key, value)
	}
	return dir
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	output, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func localOrEmpty(dir string, key string) string {
	output, _ := exec.Command("git", "-C", dir, "config", "--local", "--get", key).Output()
	return strings.TrimSpace(string(output))
}

func TestApply(t *testing.T) {
	cases := []struct {
		name      string
		existing  map[string]string
		identity  config.Identity
		wantName  string
		wantEmail string
	}{
		{"once fills unset keys", nil,
			config.Identity{Name: "Me", Email: "me@x.com", Mode: config.IdentityOnce}, "Me", "me@x.com"},
		{"once keeps existing keys", map[string]string{nameKey: "Old", emailKey: "old@x.com"},
			config.Identity{Name: "Me", Email: "me@x.com", Mode: config.IdentityOnce}, "Old", "old@x.com"},
		{"once fills only the missing key", map[string]string{nameKey: "Old"},
			config.Identity{Name: "Me", Email: "me@x.com", Mode: config.IdentityOnce}, "Old", "me@x.com"},
		{"always replaces existing keys", map[string]string{nameKey: "Old", emailKey: "old@x.com"},
			config.Identity{Name: "Me", Email: "me@x.com", Mode: config.IdentityAlways}, "Me", "me@x.com"},
		{"empty value leaves key alone", map[string]string{emailKey: "old@x.com"},
			config.Identity{Name: "Me", Mode: config.IdentityAlways}, "Me", "old@x.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := newRepo(t, tc.existing)
			if err := Apply(tc.identity, dir, failLookup(t)); err != nil {
				t.Fatal(err)
			}
			if got := localOrEmpty(dir, nameKey); got != tc.wantName {
				t.Fatalf("name: got %q, want %q", got, tc.wantName)
			}
			if got := localOrEmpty(dir, emailKey); got != tc.wantEmail {
				t.Fatalf("email: got %q, want %q", got, tc.wantEmail)
			}
		})
	}
}

func TestApplyOutsideRepoDoesNothing(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CEILING_DIRECTORIES", t.TempDir())
	dir := t.TempDir()
	if err := Apply(config.Identity{Name: "Me", Mode: config.IdentityAlways}, dir, failLookup(t)); err != nil {
		t.Fatal(err)
	}
}

func failLookup(t *testing.T) Lookup {
	return func() (config.Identity, error) {
		t.Fatal("unexpected GitHub lookup")
		return config.Identity{}, nil
	}
}

func TestApplyFromGithub(t *testing.T) {
	fetched := config.Identity{Name: "Gh Name", Email: "gh@x.com"}
	cases := []struct {
		name      string
		existing  map[string]string
		identity  config.Identity
		wantCalls int
		wantName  string
		wantEmail string
	}{
		{"fills both unset keys with one lookup", nil,
			config.Identity{Mode: config.IdentityOnce, FromGithub: true}, 1, "Gh Name", "gh@x.com"},
		{"explicit value wins over GitHub", nil,
			config.Identity{Email: "me@x.com", Mode: config.IdentityOnce, FromGithub: true}, 1, "Gh Name", "me@x.com"},
		{"once skips lookup when keys are set", map[string]string{nameKey: "Old", emailKey: "old@x.com"},
			config.Identity{Mode: config.IdentityOnce, FromGithub: true}, 0, "Old", "old@x.com"},
		{"always replaces with GitHub values", map[string]string{nameKey: "Old", emailKey: "old@x.com"},
			config.Identity{Mode: config.IdentityAlways, FromGithub: true}, 1, "Gh Name", "gh@x.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := newRepo(t, tc.existing)
			calls := 0
			lookup := func() (config.Identity, error) {
				calls++
				return fetched, nil
			}
			if err := Apply(tc.identity, dir, lookup); err != nil {
				t.Fatal(err)
			}
			if calls != tc.wantCalls {
				t.Fatalf("lookup calls: got %d, want %d", calls, tc.wantCalls)
			}
			if got := localOrEmpty(dir, nameKey); got != tc.wantName {
				t.Fatalf("name: got %q, want %q", got, tc.wantName)
			}
			if got := localOrEmpty(dir, emailKey); got != tc.wantEmail {
				t.Fatalf("email: got %q, want %q", got, tc.wantEmail)
			}
		})
	}
}

func TestApplyReturnsLookupError(t *testing.T) {
	dir := newRepo(t, nil)
	lookup := func() (config.Identity, error) { return config.Identity{}, errors.New("no token") }
	if err := Apply(config.Identity{Mode: config.IdentityOnce, FromGithub: true}, dir, lookup); err == nil {
		t.Fatal("expected error")
	}
}
