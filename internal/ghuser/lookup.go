package ghuser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/barnuri/ghx/internal/config"
	"github.com/barnuri/ghx/internal/proxy"
)

const noreplyDomain = "users.noreply.github.com"

// Lookup returns the GitHub display name and email of the profile's login, cached for a day.
func Lookup(stateDir string, ghPath string, profile config.Profile) (config.Identity, error) {
	file := cachePath(stateDir, profile)
	if cached, ok := readCache(file); ok {
		return cached, nil
	}
	fetched, err := fetch(ghPath, profile)
	if err != nil {
		return config.Identity{}, err
	}
	writeCache(file, fetched)
	return fetched, nil
}

func fetch(ghPath string, profile config.Profile) (config.Identity, error) {
	token, err := proxy.FetchToken(ghPath, profile)
	if err != nil {
		return config.Identity{}, err
	}
	env := proxy.WithProfile(os.Environ(), profile, token)
	var user account
	if err := api(ghPath, env, "user", &user); err != nil {
		return config.Identity{}, err
	}
	var addresses []emailAddress
	// user/emails needs the "user" token scope. Without it the public or noreply email is used.
	_ = api(ghPath, env, "user/emails", &addresses)
	return config.Identity{Name: user.Name, Email: pickEmail(addresses, user, profile.HostOrDefault())}, nil
}

func pickEmail(addresses []emailAddress, user account, host string) string {
	for _, address := range addresses {
		if address.Primary {
			return address.Email
		}
	}
	if user.Email != "" || host != config.DefaultHost {
		return user.Email
	}
	return fmt.Sprintf("%d+%s@%s", user.ID, user.Login, noreplyDomain)
}

func api(ghPath string, env []string, path string, target any) error {
	command := exec.Command(ghPath, "api", path)
	command.Env = env
	output, err := command.Output()
	if err != nil {
		return fmt.Errorf("gh api %s: %w", path, err)
	}
	return json.Unmarshal(output, target)
}
