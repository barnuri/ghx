package main

import (
	"fmt"
	"os"

	"github.com/barnuri/ghx/internal/config"
	"github.com/barnuri/ghx/internal/ghbin"
	"github.com/barnuri/ghx/internal/platform"
	"github.com/barnuri/ghx/internal/proxy"
	"github.com/barnuri/ghx/internal/release"
	"github.com/barnuri/ghx/internal/resolve"
	"github.com/barnuri/ghx/internal/update"
)

var version = "dev"

const (
	metaCommand          = "ghx"
	backgroundUpdateFlag = "--ghx-background-update"
	profileOverrideEnv   = "GHX_PROFILE"
	noUpdateEnv          = "GHX_NO_UPDATE"
	metaUsage            = `Usage: gh ghx <command>

Commands:
  which     Show the profile, gh user and gh binary for the current directory
  current   Print only the profile name for the current directory (empty when none)
  update    Update ghx now (and gh when manage_gh is true)
  version   Show the ghx version
  help      Show this help

Every other command goes to the real gh unchanged.`
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ghx:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		return err
	}
	stateDir := config.StateDir()
	updater := newUpdater(cfg, stateDir)
	if len(args) == 1 && args[0] == backgroundUpdateFlag {
		updater.RunQuietly()
		return nil
	}
	ghPath, err := locateGh(cfg, updater)
	if err != nil {
		return err
	}
	scheduleUpdate(cfg, stateDir)
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	match := resolve.Profile(cfg, cwd, os.Getenv(profileOverrideEnv))
	if len(args) > 0 && args[0] == metaCommand {
		return runMeta(args[1:], match, ghPath, updater)
	}
	env, err := buildEnv(args, match, ghPath)
	if err != nil {
		return err
	}
	exitCode, err := platform.Run(ghPath, args, env)
	if err != nil {
		return fmt.Errorf("run %s: %w", ghPath, err)
	}
	os.Exit(exitCode)
	return nil
}

func runMeta(args []string, match resolve.Match, ghPath string, updater update.Updater) error {
	command := "help"
	if len(args) > 0 {
		command = args[0]
	}
	switch command {
	case "which":
		printWhich(match, ghPath)
	case "current":
		if match.Found() {
			fmt.Println(match.ProfileName)
		}
	case "update":
		return updater.Run(os.Stdout)
	case "version":
		fmt.Printf("ghx %s\n", version)
	case "help", "--help", "-h":
		fmt.Println(metaUsage)
	default:
		return fmt.Errorf("unknown command %q. Run: gh ghx help", command)
	}
	return nil
}

func newUpdater(cfg config.Config, stateDir string) update.Updater {
	client := release.NewClient()
	return update.Updater{
		StateDir: stateDir,
		Self:     update.SelfUpdater{Client: client, CurrentVersion: version},
		Gh:       update.ManagedGh{Client: client, StateDir: stateDir},
		ManageGh: cfg.ManageGh,
	}
}

func locateGh(cfg config.Config, updater update.Updater) (string, error) {
	if os.Getenv("GHX_GH") != "" || !cfg.ManageGh {
		return ghbin.Locate()
	}
	if managedPath := updater.Gh.Path(); managedPath != "" {
		return managedPath, nil
	}
	fmt.Fprintln(os.Stderr, "ghx: installing gh (manage_gh: true)...")
	if _, err := updater.Gh.Update(); err != nil {
		return "", fmt.Errorf("install gh: %w", err)
	}
	return updater.Gh.Path(), nil
}

func scheduleUpdate(cfg config.Config, stateDir string) {
	if cfg.DisableUpdates || os.Getenv(noUpdateEnv) != "" || !update.IsDue(stateDir) {
		return
	}
	executable, err := os.Executable()
	if err != nil || update.MarkChecked(stateDir) != nil {
		return
	}
	_ = platform.StartDetached(executable, backgroundUpdateFlag)
}

func buildEnv(args []string, match resolve.Match, ghPath string) ([]string, error) {
	env := os.Environ()
	if !match.Found() || !proxy.NeedsToken(args) || proxy.HasExplicitToken(env) {
		return env, nil
	}
	token, err := proxy.FetchToken(ghPath, match.Profile)
	if err != nil {
		return nil, err
	}
	return proxy.WithProfile(env, match.Profile, token), nil
}

func printWhich(match resolve.Match, ghPath string) {
	if !match.Found() {
		fmt.Printf("profile: none (plain gh)\ngh:      %s\n", config.AbbreviateHome(ghPath))
		return
	}
	fmt.Printf("profile: %s\nuser:    %s\nhost:    %s\nsource:  %s\ngh:      %s\n",
		match.ProfileName, match.Profile.User, match.Profile.HostOrDefault(),
		config.AbbreviateHome(match.Source), config.AbbreviateHome(ghPath))
}
