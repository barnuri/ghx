package resolve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/barnuri/ghx/internal/config"
)

func newTree(t *testing.T) (string, config.Config) {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"work/api/src", "workshop", "personal/blog", "work/oss"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Config{
		Profiles: map[string]config.Profile{
			"work":     {User: "me-work"},
			"personal": {User: "me"},
		},
		Dirs: map[string]string{
			filepath.Join(root, "work"):     "work",
			filepath.Join(root, "personal"): "personal",
			filepath.Join(root, "work/oss"): "personal",
		},
	}
	return root, cfg
}

func TestProfile(t *testing.T) {
	root, cfg := newTree(t)
	cases := []struct {
		name string
		cwd  string
		want string
	}{
		{"configured dir itself", "work", "work"},
		{"nested subfolder inherits", "work/api/src", "work"},
		{"deepest configured dir wins", "work/oss", "personal"},
		{"sibling with shared prefix does not match", "workshop", ""},
		{"other tree", "personal/blog", "personal"},
		{"outside every dir", ".", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Profile(cfg, filepath.Join(root, tc.cwd), "")
			if got.ProfileName != tc.want {
				t.Fatalf("got %q, want %q", got.ProfileName, tc.want)
			}
		})
	}
}

func TestProfileOverrideWins(t *testing.T) {
	root, cfg := newTree(t)
	got := Profile(cfg, filepath.Join(root, "work"), "personal")
	if got.ProfileName != "personal" || got.Profile.User != "me" || got.Source != overrideSource {
		t.Fatalf("unexpected match %+v", got)
	}
}

func TestProfileDefault(t *testing.T) {
	root, cfg := newTree(t)
	cfg.Default = "personal"
	if got := Profile(cfg, root, ""); got.ProfileName != "personal" || got.Source != defaultSource {
		t.Fatalf("unexpected match %+v", got)
	}
}

func TestRepoFileBeatsDirs(t *testing.T) {
	root, cfg := newTree(t)
	repoFile := filepath.Join(root, "work/api", RepoFileName)
	if err := os.WriteFile(repoFile, []byte("# repo profile\n\npersonal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Profile(cfg, filepath.Join(root, "work/api/src"), "")
	if got.ProfileName != "personal" || got.Source != canonical(repoFile) {
		t.Fatalf("unexpected match %+v", got)
	}
}

func TestRepoFileWithUnknownNameIsGhUser(t *testing.T) {
	root, cfg := newTree(t)
	if err := os.WriteFile(filepath.Join(root, "workshop", RepoFileName), []byte("octocat\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := Profile(cfg, filepath.Join(root, "workshop"), "")
	if got.Profile.User != "octocat" || got.Profile.HostOrDefault() != config.DefaultHost {
		t.Fatalf("unexpected match %+v", got)
	}
}

func TestSymlinkedCwdResolvesToTarget(t *testing.T) {
	root, cfg := newTree(t)
	link := filepath.Join(root, "link")
	if err := os.Symlink(filepath.Join(root, "work/api"), link); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	if got := Profile(cfg, link, ""); got.ProfileName != "work" {
		t.Fatalf("got %q, want work", got.ProfileName)
	}
}

func TestTildeDirExpands(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	if err := os.MkdirAll(filepath.Join(home, "code/work/repo"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Profiles: map[string]config.Profile{"work": {User: "me-work"}},
		Dirs:     map[string]string{"~/code/work": "work"},
	}
	if got := Profile(cfg, filepath.Join(home, "code/work/repo"), ""); got.ProfileName != "work" {
		t.Fatalf("got %q, want work", got.ProfileName)
	}
}
