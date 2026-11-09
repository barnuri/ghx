package ghuser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/barnuri/ghx/internal/config"
)

const (
	cacheDirName = "github-identity"
	cacheTTL     = 24 * time.Hour
)

type cachedIdentity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func cachePath(stateDir string, profile config.Profile) string {
	return filepath.Join(stateDir, cacheDirName, profile.HostOrDefault(), profile.User+".json")
}

func readCache(file string) (config.Identity, bool) {
	info, err := os.Stat(file)
	if err != nil || time.Since(info.ModTime()) >= cacheTTL {
		return config.Identity{}, false
	}
	content, err := os.ReadFile(file)
	if err != nil {
		return config.Identity{}, false
	}
	var cached cachedIdentity
	if json.Unmarshal(content, &cached) != nil {
		return config.Identity{}, false
	}
	return config.Identity{Name: cached.Name, Email: cached.Email}, true
}

// writeCache ignores errors: a missing cache only costs another API call.
func writeCache(file string, identity config.Identity) {
	content, err := json.Marshal(cachedIdentity{Name: identity.Name, Email: identity.Email})
	if err != nil || os.MkdirAll(filepath.Dir(file), 0o700) != nil {
		return
	}
	_ = os.WriteFile(file, content, 0o600)
}
