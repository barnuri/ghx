package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	configPathEnv = "GHX_CONFIG"
	appDirName    = "ghx"
	configName    = "config.yaml"
)

func DefaultPath() string {
	if fromEnv := os.Getenv(configPathEnv); fromEnv != "" {
		return fromEnv
	}
	return filepath.Join(configDir(), appDirName, configName)
}

func configDir() string {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData
		}
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return xdg
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config")
}

func StateDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}
	return filepath.Join(cacheDir, appDirName)
}

func ExpandHome(path string) string {
	if path != "~" && !strings.HasPrefix(path, "~/") && !strings.HasPrefix(path, `~\`) {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}

func AbbreviateHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	for _, prefix := range homeForms(home) {
		if path == prefix || strings.HasPrefix(path, prefix+string(filepath.Separator)) {
			return "~" + path[len(prefix):]
		}
	}
	return path
}

func homeForms(home string) []string {
	resolved, err := filepath.EvalSymlinks(home)
	if err != nil || resolved == home {
		return []string{home}
	}
	return []string{resolved, home}
}
