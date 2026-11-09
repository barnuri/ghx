package config

import "cmp"

const (
	IdentityOnce   = "once"
	IdentityAlways = "always"
)

type Identity struct {
	Name       string
	Email      string
	Mode       string
	FromGithub bool
}

// IdentityFor takes each field the profile leaves empty from the top-level config.
func (cfg Config) IdentityFor(profile Profile) Identity {
	fromGithub := cfg.IdentityFromGithub
	if profile.IdentityFromGithub != nil {
		fromGithub = *profile.IdentityFromGithub
	}
	return Identity{
		Name:       cmp.Or(profile.Name, cfg.Name),
		Email:      cmp.Or(profile.Email, cfg.Email),
		Mode:       cmp.Or(profile.GitIdentity, cfg.GitIdentity, IdentityOnce),
		FromGithub: fromGithub,
	}
}

func (identity Identity) Enabled() bool {
	return identity.Name != "" || identity.Email != "" || identity.FromGithub
}

func validIdentityMode(mode string) bool {
	return mode == "" || mode == IdentityOnce || mode == IdentityAlways
}
