package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/barnuri/ghx/internal/ghbin"
	"github.com/barnuri/ghx/internal/release"
)

const (
	ghRepo          = "cli/cli"
	managedDirName  = "gh"
	currentFileName = "current"
)

type ManagedGh struct {
	Client   release.Client
	StateDir string
}

// Path returns the installed managed gh, or "" when none is installed yet.
func (managed ManagedGh) Path() string {
	version, err := os.ReadFile(filepath.Join(managed.root(), currentFileName))
	if err != nil {
		return ""
	}
	binary := filepath.Join(managed.root(), strings.TrimSpace(string(version)), ghbin.BinaryName())
	if _, err := os.Stat(binary); err != nil {
		return ""
	}
	return binary
}

// Update installs the latest gh release. It returns the new version, or "" when current.
func (managed ManagedGh) Update() (string, error) {
	latest, err := managed.Client.Latest(ghRepo)
	if err != nil {
		return "", err
	}
	version := latest.Version()
	if managed.Path() != "" && managed.currentVersion() == version {
		return "", nil
	}
	archive, found := latest.Find(ghAssetName(version))
	checksums, hasChecksums := latest.Find(ghChecksumsName(version))
	if !found || !hasChecksums {
		return "", fmt.Errorf("gh release %s has no %s or checksums", latest.TagName, ghAssetName(version))
	}
	content, err := managed.Client.DownloadVerified(archive, checksums)
	if err != nil {
		return "", err
	}
	binary, err := release.ExtractFile(content, archive.Name, "bin/"+ghbin.BinaryName())
	if err != nil {
		return "", err
	}
	if err := managed.install(version, binary); err != nil {
		return "", err
	}
	return version, nil
}

func (managed ManagedGh) install(version string, binary []byte) error {
	versionDir := filepath.Join(managed.root(), version)
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(versionDir, ghbin.BinaryName()), binary, 0o755); err != nil {
		return err
	}
	previous := managed.currentVersion()
	staged := filepath.Join(managed.root(), currentFileName+".new")
	if err := os.WriteFile(staged, []byte(version), 0o644); err != nil {
		return err
	}
	if err := os.Rename(staged, filepath.Join(managed.root(), currentFileName)); err != nil {
		return err
	}
	managed.removeVersionsExcept(version, previous)
	return nil
}

// removeVersionsExcept keeps the previous version too, because a running gh may still use it.
func (managed ManagedGh) removeVersionsExcept(keep ...string) {
	entries, err := os.ReadDir(managed.root())
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() || contains(keep, entry.Name()) {
			continue
		}
		_ = os.RemoveAll(filepath.Join(managed.root(), entry.Name()))
	}
}

func (managed ManagedGh) currentVersion() string {
	version, err := os.ReadFile(filepath.Join(managed.root(), currentFileName))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(version))
}

func (managed ManagedGh) root() string {
	return filepath.Join(managed.StateDir, managedDirName)
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
