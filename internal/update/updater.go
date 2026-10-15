package update

import (
	"fmt"
	"io"
	"strings"
)

type Updater struct {
	StateDir string
	Self     SelfUpdater
	Gh       ManagedGh
	ManageGh bool
}

func (updater Updater) Run(out io.Writer) error {
	var failures []string
	if newVersion, err := updater.Self.Update(); err != nil {
		failures = append(failures, "ghx: "+err.Error())
	} else {
		fmt.Fprintln(out, describe("ghx", newVersion))
	}
	if updater.ManageGh {
		if newVersion, err := updater.Gh.Update(); err != nil {
			failures = append(failures, "gh: "+err.Error())
		} else {
			fmt.Fprintln(out, describe("gh", newVersion))
		}
	}
	if len(failures) > 0 {
		return fmt.Errorf("update failed: %s", strings.Join(failures, "; "))
	}
	return nil
}

func (updater Updater) RunQuietly() {
	var log strings.Builder
	if err := updater.Run(&log); err != nil {
		log.WriteString(err.Error())
	}
	WriteLog(updater.StateDir, strings.TrimSpace(log.String()))
}

func describe(name string, newVersion string) string {
	if newVersion == "" {
		return name + " is up to date"
	}
	return name + " updated to " + newVersion
}
