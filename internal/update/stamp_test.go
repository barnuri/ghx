package update

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStamp(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "ghx")
	if !IsDue(stateDir) {
		t.Fatal("expected due without a stamp")
	}
	if err := MarkChecked(stateDir); err != nil {
		t.Fatal(err)
	}
	if IsDue(stateDir) {
		t.Fatal("expected not due right after a check")
	}
	stale := time.Now().Add(-checkInterval - time.Minute)
	if err := os.Chtimes(filepath.Join(stateDir, stampName), stale, stale); err != nil {
		t.Fatal(err)
	}
	if !IsDue(stateDir) {
		t.Fatal("expected due after the interval")
	}
}

func TestManagedGhPathEmptyWhenNotInstalled(t *testing.T) {
	if got := (ManagedGh{StateDir: t.TempDir()}).Path(); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestManagedGhInstallSetsCurrent(t *testing.T) {
	managed := ManagedGh{StateDir: t.TempDir()}
	if err := managed.install("2.0.0", []byte("old")); err != nil {
		t.Fatal(err)
	}
	if err := managed.install("2.1.0", []byte("new")); err != nil {
		t.Fatal(err)
	}
	if err := managed.install("2.2.0", []byte("newest")); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(managed.Path())
	if err != nil || string(content) != "newest" {
		t.Fatalf("got %q, %v", content, err)
	}
	if _, err := os.Stat(filepath.Join(managed.root(), "2.0.0")); !os.IsNotExist(err) {
		t.Fatal("expected versions older than previous to be removed")
	}
	if _, err := os.Stat(filepath.Join(managed.root(), "2.1.0")); err != nil {
		t.Fatal("expected previous version to be kept")
	}
}
