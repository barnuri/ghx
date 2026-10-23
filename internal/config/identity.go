package config

import "cmp"

const (
	IdentityOnce   = "once"
	IdentityAlways = "always"
)

type Identity struct {
	Name  string
	Email string
	Mode  string
}

// IdentityFor takes each field the profile leaves empty from the top-level config.
func (cfg Config) IdentityFor(profile Profile) Identity {
	return Identity{
		Name:  cmp.Or(profile.Name, cfg.Name),
		Email: cmp.Or(profile.Email, cfg.Email),
		Mode:  cmp.Or(profile.GitIdentity, cfg.GitIdentity, IdentityOnce),
	}
}

func (identity Identity) IsEmpty() bool {
	return identity.Name == "" && identity.Email == ""
}

func validIdentityMode(mode string) bool {
	return mode == "" || mode == IdentityOnce || mode == IdentityAlways
}
