package config

const DefaultHost = "github.com"

type Profile struct {
	User string `yaml:"user"`
	Host string `yaml:"host"`
}

func (profile Profile) HostOrDefault() string {
	if profile.Host == "" {
		return DefaultHost
	}
	return profile.Host
}
