package ghbin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocateFindsGhOnPath(t *testing.T) {
	first, second := t.TempDir(), t.TempDir()
	expected := filepath.Join(second, BinaryName())
	if err := os.WriteFile(expected, []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv(ghPathEnv, "")
	t.Setenv("PATH", first+string(os.PathListSeparator)+second)
	got, err := Locate()
	if err != nil || got != expected {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestLocateSkipsSelf(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	shimDir := t.TempDir()
	if err := os.Symlink(executable, filepath.Join(shimDir, BinaryName())); err != nil {
		t.Skip("symlinks unavailable:", err)
	}
	t.Setenv(ghPathEnv, "")
	t.Setenv("PATH", shimDir)
	if _, err := Locate(); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestLocateEnvOverride(t *testing.T) {
	t.Setenv(ghPathEnv, "/opt/gh")
	if got, _ := Locate(); got != "/opt/gh" {
		t.Fatalf("got %q", got)
	}
}
