package ghbin

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

const ghPathEnv = "GHX_GH"

var ErrNotFound = errors.New("gh not found on PATH; install gh or set manage_gh: true in the ghx config")

// fallbackDirs are the usual gh install dirs. Apps started from the macOS Dock (VS Code, GUI
// git clients) get a PATH without them, so git credential calls would not find gh.
var fallbackDirs = []string{"/opt/homebrew/bin", "/usr/local/bin", "/home/linuxbrew/.linuxbrew/bin"}

func BinaryName() string {
	if runtime.GOOS == "windows" {
		return "gh.exe"
	}
	return "gh"
}

// Locate finds the real gh. It skips ghx itself and the ghx install dir, so a gh shim that
// points at ghx never resolves to itself.
func Locate() (string, error) {
	if fromEnv := os.Getenv(ghPathEnv); fromEnv != "" {
		return fromEnv, nil
	}
	dirs := filepath.SplitList(os.Getenv("PATH"))
	if runtime.GOOS != "windows" {
		dirs = append(dirs, fallbackDirs...)
	}
	selfPath, selfInfo := self()
	for _, dir := range dirs {
		candidate := filepath.Join(dir, BinaryName())
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		if selfInfo != nil && (os.SameFile(selfInfo, info) || sameDir(selfPath, candidate)) {
			continue
		}
		return candidate, nil
	}
	return "", ErrNotFound
}

func self() (string, os.FileInfo) {
	executable, err := os.Executable()
	if err != nil {
		return "", nil
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	info, err := os.Stat(executable)
	if err != nil {
		return "", nil
	}
	return executable, info
}

func sameDir(selfPath string, candidate string) bool {
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return false
	}
	return filepath.Dir(resolved) == filepath.Dir(selfPath)
}
