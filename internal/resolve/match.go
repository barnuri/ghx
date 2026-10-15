package resolve

import "github.com/barnuri/ghx/internal/config"

type Match struct {
	ProfileName string
	Profile     config.Profile
	Source      string
}

func (match Match) Found() bool {
	return match.ProfileName != ""
}
