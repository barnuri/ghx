package config

const DefaultHost = "github.com"

type Profile struct {
	User        string `yaml:"user"`
	Host        string `yaml:"host"`
	Name        string `yaml:"name"`
	Email       string `yaml:"email"`
	GitIdentity string `yaml:"git_identity"`
}

func (profile Profile) HostOrDefault() string {
	if profile.Host == "" {
		return DefaultHost
	}
	return profile.Host
}
