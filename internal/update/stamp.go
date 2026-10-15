package update

import (
	"os"
	"path/filepath"
	"time"
)

const (
	checkInterval = 24 * time.Hour
	stampName     = "last-update-check"
	logName       = "update.log"
)

func IsDue(stateDir string) bool {
	info, err := os.Stat(filepath.Join(stateDir, stampName))
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) >= checkInterval
}

func MarkChecked(stateDir string) error {
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(stateDir, stampName), []byte(time.Now().UTC().Format(time.RFC3339)), 0o644)
}

func WriteLog(stateDir string, message string) {
	_ = os.MkdirAll(stateDir, 0o755)
	line := time.Now().UTC().Format(time.RFC3339) + " " + message + "\n"
	_ = os.WriteFile(filepath.Join(stateDir, logName), []byte(line), 0o644)
}
